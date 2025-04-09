/*
Package httpz provides authorization middleware for HTTP servers built on top of the standard net/http.

The middleware intercepts incoming requests and calls the Aserto authorizer service to determine if access should
be allowed or denied.
*/
package httpz

import (
	"context"
	"net/http"
	"strings"

	cerr "github.com/aserto-dev/errors"
	"github.com/aserto-dev/go-aserto/middleware"
	"github.com/aserto-dev/go-aserto/middleware/internal"
	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	aerr "github.com/aserto-dev/go-authorizer/pkg/aerr"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/structpb"
)

type (
	Policy           = middleware.Policy
	AuthorizerClient = authz.AuthorizerClient
)

/*
Middleware implements an http.Handler that can be added to routes in net/http servers.

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
	StringMapper func(*http.Request) string

	// ResourceMapper functions are used to extract structured data from incoming requests.
	ResourceMapper func(*http.Request, map[string]any)
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
		Identity:        (&IdentityBuilder{}).FromHeader("Authorization"),
		client:          client,
		policy:          policy,
		resourceMappers: []ResourceMapper{},
		policyMapper:    policyMapper,
	}
}

// Handler returns a middlleware handler that authorizes incoming requests.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		policyContext := m.policyContext()

		if m.policyMapper != nil {
			policyContext.Path = m.policyMapper(r)
		}

		resource, err := m.resourceContext(r)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		allowed, err := m.is(r.Context(), m.Identity.Build(r), policyContext, resource)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !allowed {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// HandlerFunc returns a middleware handler that wraps the given http.HandlerFunc and authorizes incoming requests.
func (m *Middleware) HandlerFunc(next http.HandlerFunc) http.Handler {
	return m.Handler(next)
}

// Check returns a new Check middleware object that can be used to make ReBAC authorization decisions for individual
// routes.
// A check call returns true if a given relation exists between an object and a subject.
func (m *Middleware) Check(options ...CheckOption) *Check {
	return newCheck(m, options...)
}

func (m *Middleware) policyContext() *api.PolicyContext {
	return internal.DefaultPolicyContext(m.policy)
}

func (m *Middleware) resourceContext(r *http.Request) (*structpb.Struct, error) {
	res := map[string]any{}
	for _, mapper := range m.resourceMappers {
		mapper(r, res)
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
		PolicyInstance:  internal.DefaultPolicyInstance(m.policy),
	}

	logger := zerolog.Ctx(ctx).With().Interface("is_request", isRequest).Logger()
	logger.Debug().Msg("authorizing request")
	ctx = logger.WithContext(ctx)

	resp, err := m.client.Is(ctx, isRequest)

	switch {
	case err != nil:
		return false, cerr.WithContext(err, ctx)
	case len(resp.GetDecisions()) != 1:
		return false, cerr.WithContext(aerr.ErrInvalidDecision, ctx)
	}

	if !resp.GetDecisions()[0].GetIs() {
		logger.Info().Msg("authorization failed")
	}

	return resp.GetDecisions()[0].GetIs(), nil
}

// WithPolicyFromURL instructs the middleware to construct the policy path from the path segment
// of the incoming request's URL.
//
// Path separators ('/') are replaced with dots ('.').
// An optional prefix can be specified to be included in all paths.
//
// # Example
//
// Using 'WithPolicyFromURL("myapp")', the route
//
//	POST /api/products/
//
// becomes the policy path
//
//	"myapp.POST.api.products"
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
	m.resourceMappers = []ResourceMapper{func(*http.Request, map[string]any) {}}
	return m
}

// WithResourceMapper sets a custom resource mapper, a function that takes an incoming request
// and returns the resource object to include with the authorization request as a `structpb.Struct`.
func (m *Middleware) WithResourceMapper(mapper ResourceMapper) *Middleware {
	m.resourceMappers = append(m.resourceMappers, mapper)
	return m
}

func urlPolicyPathMapper(prefix string) StringMapper {
	return func(r *http.Request) string {
		policyPath := append([]string{r.Method}, getPathSegments(r)...)

		if prefix != "" {
			policyPath = append([]string{strings.Trim(prefix, ".")}, policyPath...)
		}

		return strings.Join(policyPath, ".")
	}
}

func getPathSegments(r *http.Request) []string {
	return strings.Split(strings.Trim(r.URL.Path, "/"), "/")
}
