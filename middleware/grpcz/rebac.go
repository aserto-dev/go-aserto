package grpcz

import (
	"context"
	"fmt"
	"strings"

	"github.com/aserto-dev/go-aserto/middleware/internal"
	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

const MaxPermissionLen = 64

type RebacMiddleware struct {
	policy          *Policy
	client          AuthorizerClient
	policyMapper    StringMapper
	Identity        *IdentityBuilder
	resourceMappers []ResourceMapper
	subjType        string
	objType         string
	ignoredPaths    internal.Lookup[string]
	allowedMethods  internal.Lookup[string]
}

/*
WithResourceFromContextValue instructs the middleware to read the specified value from the incoming request
context and add it to the authorization resource context.

Example:

	checkMiddleware.WithResourceFromContextValue("account_id", "account")

In each incoming request, the middleware reads the value of the "account_id" key from the request context and
adds its value to the "account" field in the authorization resource context.
*/
func (c *RebacMiddleware) WithResourceFromContextValue(ctxKey any, field string) *RebacMiddleware {
	c.resourceMappers = append(c.resourceMappers, contextValueResourceMapper(ctxKey, field))
	return c
}

/*
WithSubjectType instructs the middleware to read the specified value for the subject type in the resource context.

Example:

	checkMiddleware.WithSubjectType("user")
*/
func (c *RebacMiddleware) WithSubjectType(value string) *RebacMiddleware {
	c.subjType = value
	return c
}

/*
WithObjectType instructs the middleware to read the specified value for the object type in the resource context.

Example:

	checkMiddleware.WithSubjectType("tenant")
*/
func (c *RebacMiddleware) WithObjectType(value string) *RebacMiddleware {
	c.objType = value
	return c
}

// Deprecated: Use WithAllowedMethods instead.
// WithIgnoredMethods takes as its input a list of policy paths in Rego dot notation
// (e.g. "myservice.GET.user.__id") that are ignored by the middleware. Requests that
// would normally evaluate one of these paths will be allowed to proceed without authorization.
func (c *RebacMiddleware) WithIgnoredMethods(methods []string) *RebacMiddleware {
	c.ignoredPaths = internal.NewLookup(
		lo.Map(methods, func(m string, _ int) string { return strings.ToLower(m) })...,
	)

	return c
}

// WithAllowedMethods takes a list of gRPC methods that are allowed to proceed without authorization.
// Method paths are in the format "/package.Service/Method".
// For example: "/grpc.reflection.v1.ServerReflection/ServerReflectionInfo".
func (c *RebacMiddleware) WithAllowedMethods(methods ...string) *RebacMiddleware {
	c.allowedMethods = internal.NewLookup(methods...)
	return c
}

func NewRebacMiddleware(authzClient AuthorizerClient, policy *Policy) *RebacMiddleware {
	policyMapper := methodPolicyMapper("")
	if policy.Path != "" {
		policyMapper = nil
	}

	return &RebacMiddleware{
		Identity:     (&IdentityBuilder{}).Subject().FromMetadata("authorization"),
		client:       authzClient,
		policy:       policy,
		policyMapper: policyMapper,
	}
}

// Unary returns a grpc.UnaryServiceInterceptor that authorizes incoming messages.
func (c *RebacMiddleware) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if err := c.authorize(ctx, req); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// Stream returns a grpc.StreamServerInterceptor that authorizes incoming messages.
func (c *RebacMiddleware) Stream() grpc.StreamServerInterceptor {
	return func(
		srv any,
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := stream.Context()

		if err := c.authorize(ctx, nil); err != nil {
			return err
		}

		return handler(srv, stream)
	}
}

func (c *RebacMiddleware) authorize(ctx context.Context, req any) error {
	if c.isAllowedMethod(ctx) {
		return nil
	}

	policyContext := c.policyContext()

	resource, err := c.resourceContext(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to apply resource mapper")
	}

	if c.ignoredPaths.Contains(permissionFromMethod(ctx)) {
		return nil
	}

	resp, err := c.client.Is(
		ctx,
		&authz.IsRequest{
			IdentityContext: c.identityContext(ctx, req),
			PolicyContext:   policyContext,
			ResourceContext: resource,
			PolicyInstance:  internal.DefaultPolicyInstance(c.policy),
		},
	)
	if err != nil {
		return errors.Wrap(err, "authorization call failed")
	}

	if len(resp.GetDecisions()) == 0 {
		return aerr.ErrInvalidDecision
	}

	if !resp.GetDecisions()[0].GetIs() {
		return aerr.ErrAuthorizationFailed
	}

	return nil
}

func (c *RebacMiddleware) isAllowedMethod(ctx context.Context) bool {
	method, _ := grpc.Method(ctx)
	return c.allowedMethods.Contains(method)
}

func (c *RebacMiddleware) policyContext() *api.PolicyContext {
	policyContext := internal.DefaultPolicyContext(c.policy)
	policyContext.Path = ""

	if c.policy.Path != "" {
		policyContext.Path = c.policy.Path
	}

	if policyContext.GetPath() == "" {
		path := "check"
		if c.policy.Root != "" {
			path = fmt.Sprintf("%s.%s", c.policy.Root, path)
		}

		policyContext.Path = path
	}

	return policyContext
}

func (c *RebacMiddleware) identityContext(ctx context.Context, req any) *api.IdentityContext {
	return c.Identity.build(ctx, req)
}

func (c *RebacMiddleware) resourceContext(ctx context.Context, req any) (*structpb.Struct, error) {
	res := map[string]any{}
	for _, mapper := range c.resourceMappers {
		mapper(ctx, req, res)
	}

	res["object_type"] = c.objectType()
	res["relation"] = permissionFromMethod(ctx)
	res["subject_type"] = c.subjectType()

	return structpb.NewStruct(res)
}

func permissionFromMethod(ctx context.Context) string {
	method, _ := grpc.Method(ctx)

	path := strings.ToLower(internal.ToPolicyPath(method))
	if len(path) > MaxPermissionLen {
		path = path[:MaxPermissionLen]
	}

	return path
}

func (c *RebacMiddleware) subjectType() string {
	if c.subjType != "" {
		return c.subjType
	}

	return internal.DefaultSubjType
}

func (c *RebacMiddleware) objectType() string {
	return c.objType
}
