/*
Package grpc provides authorization middleware for gRPC servers.

The middleware intercepts incoming requests/streams and calls the Aserto authorizer service to
determine if access should be granted or denied.
*/
package grpcz

import (
	"context"
	"fmt"
	"maps"

	cerr "github.com/aserto-dev/errors"
	"github.com/aserto-dev/go-aserto/middleware"
	"github.com/aserto-dev/go-aserto/middleware/grpcz/internal/pbutil"
	"github.com/aserto-dev/go-aserto/middleware/internal"
	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

type (
	Policy           = middleware.Policy
	AuthorizerClient = authz.AuthorizerClient
)

/*
Middleware implements unary and stream server interceptors that can be attached to gRPC servers.

To authorize incoming RPC calls, the middleware needs information about:

1. The user making the request.

2. The Aserto authorization policy to evaluate.

3. Optional, additional input data to the authorization policy.

The values for these parameters can be set globally or extracted dynamically from incoming messages.
*/
type Middleware struct {
	// Identity determines the caller identity used in authorization calls.
	Identity *IdentityBuilder

	client          AuthorizerClient
	policy          *Policy
	policyMapper    StringMapper
	resourceMappers []ResourceMapper
	ignoredPaths    internal.Lookup[string]
	allowedMethods  internal.Lookup[string]
}

type (
	// StringMapper functions are used to extract string values from incoming messages.
	// They are used to define identity and policy mappers.
	StringMapper func(context.Context, any) string

	// ResourceMapper functions are used to extract structured data from incoming message.
	ResourceMapper func(context.Context, any, map[string]any)
)

// New creates middleware for the specified policy.
//
// The new middleware is created with default identity and policy path mapper.
// Those can be overridden using `Middleware.Identity` to specify the caller's identity, or using
// the middleware's ".With...()" functions to set policy path and resource mappers.
func New(authzClient AuthorizerClient, policy *Policy) *Middleware {
	policyMapper := methodPolicyMapper("")
	if policy.Path != "" {
		policyMapper = nil
	}

	return &Middleware{
		Identity:        (&IdentityBuilder{}).FromMetadata("authorization"),
		client:          authzClient,
		policy:          policy,
		policyMapper:    policyMapper,
		resourceMappers: []ResourceMapper{},
	}
}

// Deprecated: Use WithAllowedMethods instead.
// WithIgnoredMethods takes as its input a list of policy paths in Rego dot notation
// (e.g. "myservice.GET.user.__id") that are ignored by the middleware. Requests that
// would normally evaluate one of these paths will be allowed to proceed without authorization.
func (m *Middleware) WithIgnoredMethods(paths []string) *Middleware {
	m.ignoredPaths = internal.NewLookup(paths...)
	return m
}

// WithAllowedMethods takes a list of gRPC methods that are allowed to proceed without authorization.
// Method paths are in the format "/package.Service/Method".
// For example: "/grpc.reflection.v1.ServerReflection/ServerReflectionInfo".
func (m *Middleware) WithAllowedMethods(methods ...string) *Middleware {
	m.allowedMethods = internal.NewLookup(methods...)
	return m
}

// WithPolicyPathMapper takes a custom StringMapper for extracting the authorization policy path form
// incoming message.
func (m *Middleware) WithPolicyPathMapper(mapper StringMapper) *Middleware {
	m.policyMapper = mapper
	return m
}

/*
WithResourceFromFields instructs the middleware to select the specified fields from incoming messages and
use them as the resource in authorization calls. Fields are expressed as a field mask.

Note: Protobuf message fields are identified using their JSON names.

Example:

	middleware.WithResourceFromFields("product.type", "address")

This call would result in an authorization resource with the following structure:

	  {
		  "product": {
			  "type": <value from message>
		  },
		  "address": <value from message>
	  }

If the value of "address" is itself a message, all of its fields are included.
*/
func (m *Middleware) WithResourceFromFields(fields ...string) *Middleware {
	if len(fields) == 1 && fields[0] == "*" {
		m.resourceMappers = append(m.resourceMappers, reqMessageResourceMapper())
		return m
	}

	m.resourceMappers = append(m.resourceMappers, messageResourceMapper(map[string][]string{}, fields...))

	return m
}

/*
WithResourceFromMessageByPath behaves similarly to `WithResourceFromFields` but allows specifying different sets
of fields for different method paths.

Example:

	  middleware.WithResourceFromMessageByPath(
		  "/example.ExampleService/Method1": []string{"field1", "field2"},
		  "/example.ExampleService/Method2": []string{"field1", "field2"},
		  "id", "name",
	  )

When Method1 or Method2 are called, the middleware constructs in a authorization resource with the following structure:

	  {
		  "field1": <value from message>,
		  "field2": <value from message>
	  }

For all other methods, the middleware constructs in a authorization resource with the following structure:

	  {
		  "id": <value from message>,
		  "name": <value from message>
	  }
*/
func (m *Middleware) WithResourceFromMessageByPath(fieldsByPath map[string][]string, defaults ...string) *Middleware {
	m.resourceMappers = append(m.resourceMappers, messageResourceMapper(fieldsByPath, defaults...))
	return m
}

/*
WithResourceFromContextValue instructs the middleware to read the specified value from the incoming request
context and add it to the authorization resource context.

Example:

	middleware.WithResourceFromContextValue("account_id", "account")

In each incoming request, the middleware reads the value of the "account_id" key from the request context and
adds its value to the "account" field in the authorization resource context.
*/
func (m *Middleware) WithResourceFromContextValue(ctxKey any, field string) *Middleware {
	m.resourceMappers = append(m.resourceMappers, contextValueResourceMapper(ctxKey, field))
	return m
}

// WithResourceMapper takes a custom StructMapper for extracting the authorization resource context from
// incoming messages.
func (m *Middleware) WithResourceMapper(mapper ResourceMapper) *Middleware {
	m.resourceMappers = append(m.resourceMappers, mapper)
	return m
}

// Unary returns a grpc.UnaryServiceInterceptor that authorizes incoming messages.
func (m *Middleware) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if err := m.authorize(ctx, req); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// Stream returns a grpc.StreamServerInterceptor that authorizes incoming messages.
func (m *Middleware) Stream() grpc.StreamServerInterceptor {
	return func(
		srv any,
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := stream.Context()

		if err := m.authorize(ctx, nil); err != nil {
			return err
		}

		return handler(srv, stream)
	}
}

func (m *Middleware) authorize(ctx context.Context, req any) error {
	if m.isAllowedMethod(ctx) {
		return nil
	}

	policyContext := internal.DefaultPolicyContext(m.policy)
	if m.policyMapper != nil {
		policyContext.Path = m.policyMapper(ctx, req)
	}

	if m.ignoredPaths.Contains(policyContext.GetPath()) {
		return nil
	}

	resource, err := m.resourceContext(ctx, req)
	if err != nil {
		return cerr.WrapContext(err, ctx, "failed to apply resource mapper")
	}

	isReq := &authz.IsRequest{
		IdentityContext: m.Identity.build(ctx, req),
		PolicyContext:   policyContext,
		ResourceContext: resource,
		PolicyInstance:  internal.DefaultPolicyInstance(m.policy),
	}

	logger := zerolog.Ctx(ctx).With().Interface("is", isReq).Logger()
	logger.Debug().Msg("authorizing request")
	ctx = logger.WithContext(ctx)

	resp, err := m.client.Is(ctx, isReq)
	if err != nil {
		return cerr.WrapContext(err, ctx, "authorization call failed")
	}

	if len(resp.GetDecisions()) == 0 {
		return cerr.WithContext(aerr.ErrInvalidDecision, ctx)
	}

	if !resp.GetDecisions()[0].GetIs() {
		return cerr.WithContext(aerr.ErrAuthorizationFailed, ctx)
	}

	return nil
}

func (m *Middleware) isAllowedMethod(ctx context.Context) bool {
	method, _ := grpc.Method(ctx)
	return m.allowedMethods.Contains(method)
}

func (m *Middleware) resourceContext(ctx context.Context, req any) (*structpb.Struct, error) {
	res := map[string]any{}
	for _, mapper := range m.resourceMappers {
		mapper(ctx, req, res)
	}

	return structpb.NewStruct(res)
}

func methodPolicyMapper(policyRoot string) StringMapper {
	return func(ctx context.Context, _ any) string {
		method, _ := grpc.Method(ctx)
		path := internal.ToPolicyPath(method)

		if policyRoot == "" {
			return path
		}

		return fmt.Sprintf("%s.%s", policyRoot, internal.ToPolicyPath(method))
	}
}

func messageResourceMapper(fieldsByPath map[string][]string, defaults ...string) ResourceMapper {
	return func(ctx context.Context, req any, res map[string]any) {
		method, _ := grpc.Method(ctx)

		fields, ok := fieldsByPath[method]
		if !ok || len(fields) == 0 {
			fields = defaults
		}

		if len(fields) > 0 && req != nil {
			msg, ok := req.(protoreflect.ProtoMessage)
			if !ok {
				panic("not a proto message")
			}

			resource, _ := pbutil.Select(msg, fields...)

			maps.Copy(res, resource.AsMap())
		}
	}
}

func reqMessageResourceMapper() ResourceMapper {
	return func(ctx context.Context, req any, res map[string]any) {
		if req != nil {
			protoReq, ok := req.(protoreflect.ProtoMessage)
			if !ok {
				return
			}

			message := protoReq.ProtoReflect()
			fields := message.Descriptor().Fields()

			for idx := range fields.Len() {
				field := fields.Get(idx)
				value := protoReq.ProtoReflect().Get(field).String()

				var err error

				val, err := structpb.NewValue(value)
				if err != nil {
					continue
				}

				res[string(field.Name())] = val.AsInterface()
			}
		}
	}
}

func contextValueResourceMapper(ctxKey any, field string) ResourceMapper {
	return func(ctx context.Context, _ any, res map[string]any) {
		if v := ctx.Value(ctxKey); v != nil {
			res[field] = v
		}
	}
}
