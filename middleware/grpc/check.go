package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/aserto-dev/go-aserto/middleware/internal"
	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

type CheckMiddleware struct {
	policy          *Policy
	client          AuthorizerClient
	policyMapper    StringMapper
	Identity        *IdentityBuilder
	resourceMappers []ResourceMapper
	subjType        string
	objType         string
}

/*
WithResourceFromContextValue instructs the middleware to read the specified value from the incoming request
context and add it to the authorization resource context.

Example:

	checkMiddleware.WithResourceFromContextValue("account_id", "account")

In each incoming request, the middleware reads the value of the "account_id" key from the request context and
adds its value to the "account" field in the authorization resource context.
*/
func (c *CheckMiddleware) WithResourceFromContextValue(ctxKey interface{}, field string) *CheckMiddleware {
	c.resourceMappers = append(c.resourceMappers, contextValueResourceMapper(ctxKey, field))
	return c
}

func (c *CheckMiddleware) WithSubjectType(value string) *CheckMiddleware {
	c.subjType = value
	return c
}

func (c *CheckMiddleware) WithObjectType(value string) *CheckMiddleware {
	c.objType = value
	return c
}

func NewCheckMiddleware(authzClient AuthorizerClient, policy *Policy) *CheckMiddleware {
	policyMapper := methodPolicyMapper("")
	if policy.Path != "" {
		policyMapper = nil
	}

	return &CheckMiddleware{
		Identity:     (&IdentityBuilder{}).FromMetadata("authorization"),
		client:       authzClient,
		policy:       policy,
		policyMapper: policyMapper,
	}
}

// Unary returns a grpc.UnaryServiceInterceptor that authorizes incoming messages.
func (c *CheckMiddleware) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if err := c.authorize(ctx, req); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// Stream returns a grpc.StreamServerInterceptor that authorizes incoming messages.
func (c *CheckMiddleware) Stream() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
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

func (c *CheckMiddleware) authorize(ctx context.Context, req interface{}) error {
	policyContext := c.policyContext()
	resource, err := c.resourceContext(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to apply resource mapper")
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

	if len(resp.Decisions) == 0 {
		return aerr.ErrInvalidDecision
	}

	if !resp.Decisions[0].Is {
		return aerr.ErrAuthorizationFailed
	}

	return nil
}

func (c *CheckMiddleware) policyContext() *api.PolicyContext {
	policyContext := internal.DefaultPolicyContext(c.policy)
	policyContext.Path = ""

	if c.policy.Path != "" {
		policyContext.Path = c.policy.Path
	}

	if policyContext.Path == "" {
		path := "check"
		if c.policy.Root != "" {
			path = fmt.Sprintf("%s.%s", c.policy.Root, path)
		}

		policyContext.Path = path
	}

	return policyContext
}

func (c *CheckMiddleware) identityContext(ctx context.Context, req interface{}) *api.IdentityContext {
	return c.Identity.build(ctx, req)
}

func (c *CheckMiddleware) resourceContext(ctx context.Context, req interface{}) (*structpb.Struct, error) {
	res := map[string]interface{}{}
	for _, mapper := range c.resourceMappers {
		mapper(ctx, req, res)
	}
	res["object_type"] = c.objectType()
	res["relation"] = methodResource(ctx)
	res["subject_type"] = c.subjectType()

	return structpb.NewStruct(res)
}

func methodResource(ctx context.Context) string {
	method, _ := grpc.Method(ctx)
	path := strings.ToLower(internal.ToPolicyPath(method))
	return path
}

func (c *CheckMiddleware) subjectType() string {
	if c.subjType != "" {
		return c.subjType
	}

	return "user"
}

func (c *CheckMiddleware) objectType() string {
	if c.objType != "" {
		return c.objType
	}

	return "tenant"
}
