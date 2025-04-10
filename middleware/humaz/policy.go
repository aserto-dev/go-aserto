package humaz

import (
	"context"
	"net/http"
	"strings"

	cerr "github.com/aserto-dev/errors"
	"github.com/aserto-dev/go-aserto/middleware"
	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/aserto-dev/go-authorizer/pkg/aerr"
	"github.com/danielgtaylor/huma/v2"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/structpb"
)

type (
	Policy           = middleware.Policy
	AuthorizerClient = authz.AuthorizerClient
)

/*
Middleware implements middleware that can be added to routes in Gin servers.

To authorize incoming requests, the middleware needs information about:

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
}

type (
	// StringMapper functions are used to extract string values from incoming requests.
	// They are used to define policy mappers.
	StringMapper func(huma.Context) string

	// ResourceMapper functions are used to extract structured data from incoming requests.
	// The optional resource mapper is a ResourceMapper.
	ResourceMapper func(huma.Context, map[string]interface{})
)

// New creates middleware for the specified policy.
//
// The new middleware is created with default identity and policy path mapper.
// Those can be overridden using `Middleware.Identity` to specify the caller's identity, or using
// the middleware's ".With...()" functions to set policy path and resource mappers.
func New(client AuthorizerClient, policy *Policy) *Middleware {
	policyMapper := urlPolicyPathMapper("")
	if policy.Path != "" {
		policyMapper = nil
	}

	return &Middleware{
		client:          client,
		Identity:        (&IdentityBuilder{}).FromHeader("Authorization"),
		policy:          policy,
		resourceMappers: []ResourceMapper{defaultResourceMapper},
		policyMapper:    policyMapper,
	}
}

// Handler is the middleware implementation. It is how an Authorizer is wired to a Huma router.
func (m *Middleware) Handler(c huma.Context, next func(huma.Context)) {
	policyContext := m.policyContext()

	if m.policyMapper != nil {
		policyContext.Path = m.policyMapper(c)
	}

	resource, err := m.resourceContext(c)
	if err != nil {
		c.SetStatus(http.StatusInternalServerError)
		return
	}

	allowed, err := m.is(c.Context(), m.Identity.Build(c), policyContext, resource)
	if err != nil {
		c.SetStatus(http.StatusInternalServerError)
		return
	}

	if !allowed {
		c.SetStatus(http.StatusForbidden)
		return
	}

	next(c)
}

// Check returns a new middleware handler that can be used to make ReBAC authorization decisions for individual
// routes.
// The check handler authorizers requests if the caller has a given relation to or permission on a specified object.
func (m *Middleware) Check(options ...CheckOption) func(c huma.Context, next func(huma.Context)) {
	return func(c huma.Context, next func(huma.Context)) {
		newCheck(m, options...).Handler(c, next)
	}
}

// Allowed returns a function that can be used to check if the request is allowed or not.
// It returns false if the request is not allowed or an error in the check happens.
// The function can be used in a route handler to check if the request is allowed.
func (m *Middleware) Allowed(options ...CheckOption) func(c huma.Context) (bool, error) {
	return func(c huma.Context) (bool, error) {
		return newCheck(m, options...).Allowed(c)
	}
}

func (m *Middleware) policyContext() *api.PolicyContext {
	return &api.PolicyContext{
		Path:      m.policy.Path,
		Decisions: []string{m.policy.Decision},
	}
}

func (m *Middleware) resourceContext(ctx huma.Context) (*structpb.Struct, error) {
	res := map[string]any{}
	for _, mapper := range m.resourceMappers {
		mapper(ctx, res)
	}

	return structpb.NewStruct(res)
}

func (m *Middleware) is(
	ctx context.Context,
	identityContext *api.IdentityContext,
	policyContext *api.PolicyContext,
	resourceContext *structpb.Struct,
) (bool, error) {
	isRequest := &authz.IsRequest{
		IdentityContext: identityContext,
		PolicyContext:   policyContext,
		ResourceContext: resourceContext,
		PolicyInstance: &api.PolicyInstance{
			Name:          m.policy.Name,
			InstanceLabel: m.policy.Name,
		},
	}

	logger := zerolog.Ctx(ctx).With().Interface("is_request", isRequest).Logger()
	logger.Debug().Msg("authorizing request")
	ctx = logger.WithContext(ctx)

	resp, err := m.client.Is(ctx, isRequest)

	switch {
	case err != nil:
		return false, cerr.WithContext(err, ctx)
	case len(resp.Decisions) != 1:
		return false, cerr.WithContext(aerr.ErrInvalidDecision, ctx)
	}

	if !resp.Decisions[0].Is {
		logger.Info().Msg("authorization failed")
	}

	return resp.Decisions[0].Is, nil
}

// WithPolicyFromURL instructs the middleware to construct the policy path from the path segment
// of the incoming request's URL.
//
// Path separators ('/') are replaced with dots ('.'). If the request uses gorilla/mux to define path
// parameters, those are added to the path with two leading underscores.
// An optional prefix can be specified to be included in all paths.
//
// # Example
//
// Using 'WithPolicyFromURL("myapp")', the route
//
//	POST /products/{id}
//
// becomes the policy path
//
//	"myapp.POST.products.__id"
func (m *Middleware) WithPolicyFromURL(prefix string) *Middleware {
	m.policyMapper = urlPolicyPathMapper(prefix)
	return m
}

// WithPolicyPathMapper sets a custom policy mapper, a function that takes an incoming request
// and returns the path within the policy of the package to query.
func (m *Middleware) WithPolicyPathMapper(mapper StringMapper) *Middleware {
	m.policyMapper = mapper
	return m
}

// WithNoResourceContext causes the middleware to include no resource context in authorization request instead
// of the default behavior that sends all URL path parameters.
func (m *Middleware) WithNoResourceContext() *Middleware {
	m.resourceMappers = []ResourceMapper{}
	return m
}

// WithResourceMapper sets a custom resource mapper, a function that takes an incoming request
// and returns the resource object to include with the authorization request as a `structpb.Struct`.
func (m *Middleware) WithResourceMapper(mapper ResourceMapper) *Middleware {
	m.resourceMappers = append(m.resourceMappers, mapper)
	return m
}

func defaultResourceMapper(ctx huma.Context, resource map[string]interface{}) {
	for _, param := range ctx.Operation().Parameters {
		resource[param.Name] = ctx.Param(param.Name)
	}
}

func urlPolicyPathMapper(prefix string) StringMapper {
	return func(c huma.Context) string {
		policyPath := []string{c.Method()}

		segments := getPathSegments(c)

		if len(c.Operation().Parameters) > 0 {
			for i, segment := range segments {
				if strings.HasPrefix(segment, ":") {
					segments[i] = "__" + segment[1:]
				}
			}
		}

		policyPath = append(policyPath, segments...)

		if prefix != "" {
			policyPath = append([]string{strings.Trim(prefix, ".")}, policyPath...)
		}

		return strings.Join(policyPath, ".")
	}
}

func getPathSegments(ctx huma.Context) []string {
	path := ctx.Operation().Path
	if len(ctx.Operation().Parameters) > 0 {
		path = ctx.Operation().Path
	}

	return strings.Split(strings.Trim(path, "/"), "/")
}
