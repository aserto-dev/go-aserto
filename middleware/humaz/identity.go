package humaz

import (
	"strings"

	"github.com/aserto-dev/go-aserto/middleware"
	"github.com/aserto-dev/go-aserto/middleware/internal"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
	"github.com/danielgtaylor/huma/v2"
	"github.com/lestrrat-go/jwx/jwt"
)

// IdentityMapper is the type of callback functions that can inspect incoming HTTP requests
// and set the caller's identity.
type IdentityMapper func(huma.Context, middleware.Identity)

// IdentityBuilder is used to configure what information about caller identity is sent in authorization calls.
type IdentityBuilder struct {
	identityType    api.IdentityType
	defaultIdentity string
	mapper          IdentityMapper
}

// Static values

// Call JWT() to indicate that the user's identity is expressed as a string-encoded JWT.
//
// JWT() is always called in conjunction with another method that provides the user ID itself.
// For example:
//
//	idBuilder.JWT().FromHeader("Authorization")
func (b *IdentityBuilder) JWT() *IdentityBuilder {
	b.identityType = api.IdentityType_IDENTITY_TYPE_JWT
	return b
}

// Call Subject() to indicate that the user's identity is a subject name (email, userid, etc.).

// Subject() is always used in conjunction with another method that provides the user ID itself.
// For example:
//
//	idBuilder.Subject().FromContextValue("username")
func (b *IdentityBuilder) Subject() *IdentityBuilder {
	b.identityType = api.IdentityType_IDENTITY_TYPE_SUB
	return b
}

// Call Manual() to indicate that the user's identity is set manually and isn't resolved to a user by the authorizer.
//
// Manually set identities are available in the authorizer's policy language through the "input.identity" variable.
func (b *IdentityBuilder) Manual() *IdentityBuilder {
	b.identityType = api.IdentityType_IDENTITY_TYPE_MANUAL
	return b
}

// Call None() to indicate that requests are unauthenticated.
func (b *IdentityBuilder) None() *IdentityBuilder {
	b.identityType = api.IdentityType_IDENTITY_TYPE_NONE
	b.defaultIdentity = ""

	return b
}

// Call ID(...) to set the user's identity. If neither JWT() or Subject() are called too, IdentityMapper
// tries to infer whether the specified identity is a JWT or not.
// Passing an empty string is the same as calling .None() and results in an authorization check for anonymous access.
func (b *IdentityBuilder) ID(identity string) *IdentityBuilder {
	b.defaultIdentity = identity
	return b
}

// FromHeader retrieves caller identity from request headers.
//
// Headers are attempted in order. The first non-empty header is used.
// If none of the specified headers have a value, the request is considered anonymous.
func (b *IdentityBuilder) FromHeader(header ...string) *IdentityBuilder {
	b.mapper = func(ctx huma.Context, identity middleware.Identity) {
		for _, h := range header {
			id := ctx.Header(h)
			if id == "" {
				continue
			}

			if strings.EqualFold(h, "authorization") {
				// Authorization header is special. Need to remove "Bearer" auth scheme.
				id = b.fromAuthzHeader(id)
			}

			identity.ID(id)

			return
		}

		// None of the specified headers are present in the request.
		identity.None()
	}

	return b
}

// FromHumaContextValue extracts caller identity from a value in the incoming request context from the huma context.
//
// If the value is not present, not a string, or an empty string then the request is considered anonymous.
func (b *IdentityBuilder) FromContextValue(key string) *IdentityBuilder {
	b.mapper = func(ctx huma.Context, identity middleware.Identity) {
		// uData := tokenintrospection.GetUserFromContext(c.Context())
		b.ID(internal.ValueOrEmpty(ctx.Context(), key))
	}
	return b
}

// FromHostname extracts caller identity from the incoming request's host name.
//
// The function returns the specified hostname segment. Indexing is zero-based and starts from the left.
// Negative indices start from the right.
//
// For Example, if the hostname is "service.user.company.com" then both FromHostname(1) and
// FromHostname(-3) return the value "user".
func (b *IdentityBuilder) FromHostname(segment int) *IdentityBuilder {
	b.mapper = func(ctx huma.Context, identity middleware.Identity) {
		url := ctx.URL()
		identity.ID(internal.HostnameSegment(&url, segment))
	}

	return b
}

// Mapper takes a custom IdentityMapper to be used for extracting identity information from incoming requests.
func (b *IdentityBuilder) Mapper(mapper IdentityMapper) *IdentityBuilder {
	b.mapper = mapper
	return b
}

// Build constructs an IdentityContext that can be used in authorization requests.
func (b *IdentityBuilder) Build(ctx huma.Context) *api.IdentityContext {
	identity := internal.NewIdentity(b.identityType, b.defaultIdentity)

	if b.mapper != nil {
		b.mapper(ctx, identity)
	}

	return identity.Context()
}

func (b *IdentityBuilder) fromAuthzHeader(value string) string {
	// Authorization header is special. Need to remove "Bearer" auth scheme.
	value = strings.TrimSpace(strings.TrimPrefix(value, "Bearer"))
	if b.identityType == api.IdentityType_IDENTITY_TYPE_SUB {
		// Try to parse subject out of token
		token, err := jwt.ParseString(value)
		if err == nil {
			value = token.Subject()
		}
	}

	return value
}
