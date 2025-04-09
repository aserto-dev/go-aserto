# aserto-dev/go-aserto

![ci](https://github.com/aserto-dev/go-aserto/workflows/ci/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/aserto-dev/go-aserto.svg)](https://pkg.go.dev/github.com/aserto-dev/go-aserto)
[![Go Report Card](https://goreportcard.com/badge/github.com/aserto-dev/go-aserto)](https://goreportcard.com/report/github.com/aserto-dev/go-aserto)

Package `go-aserto` implements clients and middleware for [Aserto](http://aserto.com) services.

* Docs: https://docs.aserto.com/docs/
* API Reference:  https://aserto.readme.io/


## Install

```sh
go get -u github.com/aserto-dev/go-aserto
```

## Authorizer

The [Authorizer](https://www.topaz.sh/docs/authorizer-guide/overview) service is is an [open source authorization engine](https://www.topaz.sh)
which uses the [Open Policy Agent](https://www.openpolicyagent.org) (OPA) to make decisions by computing authorization
policies.

The `AuthorizerClient` interface, defined in
[`github.com/aserto-dev/go-authorizer/aserto/authorizer/v2`](https://github.com/aserto-dev/go-authorizer/blob/main/aserto/authorizer/v2/authorizer_grpc.pb.go#L34),
describes the operations exposed by the Aserto authorizer service.


### Client

The snippet below creates an authorizer client that connects to a topaz instance running locally:

```go
import (
	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/go-aserto/az"
)
...
azClient, err := az.New(
	aserto.WithAddr("localhost:8282"),
)
```

#### Connection Options

The options below can be specified to override default behaviors:

**`WithAddr()`** - sets the server address and port. Default: "authorizer.prod.aserto.com:8443".

**`WithAPIKeyAuth()`** - sets an API key for authentication.

**`WithTokenAuth()`** - sets an OAuth2 token to be used for authentication.

**`WithTenantID()`** - sets the aserto tenant ID.

**`WithInsecure()`** - enables/disables TLS verification. Default: false.

**`WithCACertPath()`** - adds the specified PEM certificate file to the connection's list of trusted root CAs.


### Making Authorization Calls

Use the client's `Is()` method to request authorization decisions from the Aserto authorizer service.

```go
import (
	"context"
	...
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
)

ctx := context.Background()

resp, err := azClient.Is(ctx, &authorizer.IsRequest{
	PolicyContext: &api.PolicyContext{
		Path:      		"peoplefinder.GET.users.__id",
		Decisions: 		"allowed",
	},
	IdentityContext: &api.IdentityContext{
		Identity: "<user name>",
		Type:     api.IdentityType_IDENTITY_TYPE_SUB,
	},
})
```

## Directory Service

The [Directory](https://docs.aserto.com/docs/overview/directory) stores information required to make authorization
decisions.


### Directory Client

The directory client provides access to the directory services:

1. Reader - provides functions to query the directory.
2. Writer - provides functions to mutate or delete directory data.
3. Exporter - provides bulk export of data from the directory.
4. Importer - provides bulk import of data into the directory.


To create a directory client:

```go

import (
	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/go-aserto/ds/v3"
)

...

dsClient, err := ds.New(aserto.WithAPIKeyAuth('<api key>'))
```

[Connection options](#connection-options) are the same as those for the authorizer client.
If `WithAddr()` is not provided, the default address is `directory.prod.aserto.com:8443`.


### Configuration

The hosted Aserto directory exposes all services on the same address (`directory.prod.aserto.com:8443`).
However, with Topaz or in self-hosted environments, it is possible to configure the services individually and to
disable selected services entirely.

The `directory.Config` structs allows for customization of connection options for directory services.

```go
// Config provides configuration for connecting to the Aserto Directory service.
type Config struct {
	// Base configuration. If non-nil, this configuration is used for any client that doesn't have its own configuration.
	// If nil, only clients that have their own configuration will be created.
	*client.Config

	// Reader configuration.
	Reader *client.Config `json:"reader"`

	// Writer configuration.
	Writer *client.Config `json:"writer"`

	// Importer configuration.
	Importer *client.Config `json:"importer"`

	// Exporter configuration.
	Exporter *client.Config `json:"exporter"`
}
```

The embedded `*client.Config` acts as a fallback. If no configuration is provided for a specific service, the fallback
configuration is used. If no fallback is provided, the client for that service is nil.

To create a directory client from configuration, call `Connect()` on the config struct:

```go
import (
	"context"

	"github.com/aserto-dev/go-aserto/ds/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
)

...

// Use the same address for all services.
cfg := &ds.Config{Address: "localhost:9292"}

dsClient, err := cfg.Connect()
if err != nil {
	panic(err)
}

resp, err := dsClient.Reader.GetObjects(context.Background(), &reader.GetObjectsRequest{})
```

**Examples**

All services use the same configuration:
```json
{
	"address": "directory.prod.aserto.com:8443",
	"api_key": "<API-KEY>",
	"tenant_id": "<TENANT-ID>"
}
```


All services use the same configuration except for the writer, that uses a different address:
```json
{
	"address": "localhost:9292",
	"writer": {
		"address": "localhost:9293"
	}
}
```

Only a reader and writer are configured. `Client.Importer` and `Client.Exporter` are nil:
```json
{
	"reader": {
		"address": "localhost:9292"
	},
	"writer": {
		"address": "localhost:9293"
	}
}
```


## Middleware

To easily integrate Aserto authorization into your own services middleware implementations for common
frameworks are available as submodules of `go-aserto/middleware`.

* `middleware/httpz` provides middleware for HTTP servers using the standard [net/http](https://pkg.go.dev/net/http) package.
* `middleware/gorillaz` provides middleware for HTTP servers using [gorilla/mux](https://github.com/gorilla/mux).
* `middleware/ginz` provides middleware for HTTP servers using the [Gin web framework](https://gin-gonic.com).
* `middleware/grpcz` provides middleware for gRPC servers.

When authorization middleware is configured and attached to a server, it examines incoming requests, extracts
authorization parameters such as the caller's identity, calls the Aserto authorizers, and rejects requests if their
access is denied.

All middleware are created from an `AuthorizerClient` and a `Policy` with parameters that can be shared
by all authorization calls.

```go
// Policy holds global authorization options that apply to all requests.
type Policy struct {
	// Name is the Name of the aserto policy being queried for authorization.
	Name string

	// Path is the name of the policy package to evaluate.
	// If left empty, a policy mapper must be attached to the middleware to provide
	// the policy path from incoming messages.
	Path string

	// Decision is the authorization rule to use.
	Decision string
}
```

The value of several authorization parameters often depends on the content of incoming requests. Those are:

* Identity - the identity (subject name or JWT) of the caller.
* Policy Path - the name of the authorization policy package to evaluate. A default value can be set in `Policy.Path`
  when creating the middleware, but the path is often dependent on the details of the request being authorized.
* Resource Context - Additional data sent to the authorizer as JSON.

### Identity

Middleware offer control over the identity used in authorization calls:

```go
// Use the subject name "george@acmecorp.com".
middleware.Identity.Subject().ID("george@acmecorp.com")

// Use a JWT from the Authorization header.
middleware.Identity.JWT().FromHeader("Authorization")

// Use subject name from the "identity" metadata key in the request `Context`.
middleware.Identity.Subject().FromMetadata("identity")

// Read identity from the context value "user". Middleware infers the identity type from the value.
middleware.Identity.FromContext("user")

// Manually pass the identity to the authorizer without resolving it to a user.
// Manual identities are availabe in the authorizer's policy language through the "input.identity" variable.
middleware.Manual().ID("object_id")
```

In addition, it is possible to provide custom logic to specify the caller's identity. For example, in HTTP middleware:

```go
middleware.Identity.Mapper(func(r *http.Request, identity middleware.Identity) {
	username := getUserFromRequest(r) // custom logic to get user identity

	identity.Subject().ID(username) // set the caller's identity for the request
})
```

In all cases, if a value cannot be retrieved from the specified source (header, context, etc.), the authorization
call checks for unauthenticated access.

### Policy

The authorization policy's ID and the decision to be evaluated are specified when creating authorization Middleware,
but the policy path is often derived from the URL or method being called.

By default, the policy path is derived from the URL path in HTTP middleware and the `grpc.Method` in gRPC middleware.

To provide custom logic, use `middleware.WithPolicyPathMapper()`. For example, in gRPC middleware:

```go
middleware.WithPolicyPathMapper(func(ctx context.Context, req any) string {
	path := getPolicyPath(ctx, req) // custom logic to retrieve a JWT token
	return path
})
```

### Resource

A resource can be any structured data that the authorization policy uses to evaluate decisions.
By default, middleware do not include a resource in authorization calls.

To add resource data, use `Middleware.WithResourceMapper()` to attach custom logic. For example, in HTTP middleware:

```go
middleware.WithResourceMapper(func(r *http.Request, resource map[string]any) {
	accountID := getAccountID(r)         // custom logic to retrieve a value from the request

	resource["account_id"] = accountID   // add the value as a field to the resource context
})
```

`Middleware.WithResourceMapper()` can be called multiple times to add more than one mapper. Each mapper can add
or remove fields from the resoruce context. Mappers are called in the order in which they are added.

In addition to these, each middleware has built-in mappers that can handle common use-cases.


### HTTP Middleware

Two flavors of HTTP middleware are available:

* `middleware/httpz`: Middleware for HTTP servers using the standard [net/http](https://pkg.go.dev/net/http) package.
* `middleware/gorillaz`: Middleware with support for [gorilla/mux](https://pkg.go.dev/github.com/gorilla/mux).
* `middleware/ginz`: Middleware for the [Gin](https://github.com/gin-gonic/gin) web framework.

Both are constructed and configured in a similar way. They differ in the signature of their `Handler()`
function, which is used to attach them to HTTP routes, and in the signatures of their mapper functions.

#### net/http Middleware

```go
import (
	"github.com/aserto-dev/go-aserto/middleware"
	"github.com/aserto-dev/go-aserto/middleware/httpz"
)
...
mw := httpz.New(
	azClient,
	middleware.Policy{
		Decision:	   "allowed",
	},
)
```

Adding the created authorization middleware to a basic `net/http` server may look something like this:

```go
http.Handle("/users", mw.HandlerFunc(usersHandler))
```

The default behavior of the HTTP middleware is:

* Identity is retrieved from the "Authorization" HTTP Header, if present.
* Policy path is retrieved from the request URL and method to form a path of the form `METHOD.path.to.endpoint`.
* No resource context is included in authorization calls by default.


#### gorilla/mux Middleware

```go
import (
	"github.com/aserto-dev/go-aserto/middleware"
	"github.com/aserto-dev/go-aserto/middleware/gorillaz"
)
...
mw := gorillaz.New(
	azClient,
	middleware.Policy{
		Decision:	   "allowed",
	},
)
```

Adding the created authorization middleware to a basic `net/http` server may look something like this:

```go
http.Handle("/users", mw.Handler(usersHandler))
```

The popular [`gorilla/mux`](https://github.com/gorilla/mux) package provides a powerful and flexible HTTP router
with support for URL path paremeters.
Attaching the standard authorization middleware to a `gorilla/mux` server is as simple as:

```go
router := mux.NewRouter()
router.Use(mw.Handler)

router.HandleFunc("/users/{id}", userHandler).Methods("GET")
```

The default behavior of the gorilla/mux middleware is:

* Identity is retrieved from the "Authorization" HTTP Header, if present.
* Policy path is retrieved from the request URL and method to form a path of the form `METHOD.path.to.endpoint`.
  If the route contains path parameters (e.g. `"api/products/{id}"`), the surrounding braces are replaced with a
  double-underscore prefix. For example, a request to `GET api/products/{id}` gets the policy path `GET.api.products.__id`.
* All path parameters are included in the resource context.
  For example, if the route is defined as `"api/products/{id}"` and the incoming request URL path is
  `"api/products/123"` then the resource context will be `{"id": "123"}`.


#### Gin Middleware

The gin middleware looks and behaves just like the net/http middleware but uses `gin.Context` instead of `http.Request`.


### Relation-Based Access Control (ReBAC)

In addition to the pattern described above, in which each route is authorized by its own policy module,
the HTTP middleware can be used to implement Relation-Based Access Control (ReBAC) in which authorization
decisions are made by checking if a given subject has the necessary permission or relation to the object being accessed.

See [here](https://www.topaz.sh/docs/directory) for a more in-depth overview of ReBAC in Aserto.

The canonical policy for ReBAC is [ghcr.io/aserto-policies/policy-rebac](https://github.com/aserto-templates/policy-rebac/tree/main/content).

The `Check()` function on HTTP middleware (`httpz`, `gorillaz`, or `ginz`) to annotate individual routes with
instructions for populating the resource context for ReBAC checks.

A check call needs three pieces of information:

* The type and ID of the object being accessed.
* The name of the relation or permission to check.
* The type and ID of the subject attempting to access the object.

Example:
```go
router := mux.NewRouter()
router.Handle(
	"/items/{id}",
	mw.Check(
		std.WithObjectType("item"),
		std.WithObjectIDFromVar("id"),
		std.WithRelation("read"),
	).HandlerFunc(GetItem),
).Methods("GET")
```

`GetItem()` is an http handler function that serves GET request to the `/items/{id}` route.
The `mw.Check` call only authorizes requests if the calling user has the `read` permission on an object of type `item`
with the object ID extracted from the route's `{id}` parameter.
The subject type is `user` by default and the subject ID is inferred from the `Authorization` header.

#### Check Options

The `Check()` function accepts options that configure the object, subject, and relation sent to the authorizer.

**`WithIdentityMapper(IdentityMapper)`** can be used to override the identity context sent to the authorizer. The `mapper` is a
function that takes the incoming request and a `middleware.Identity` and can set options on the `Identity` object based on
information from the request.
If an identity mapper isn't provided, the check call uses the identity configured on the middleware object on which
the `Check` call is made.

**`WithRelation(string)`** sets the relation name sent to the authorizer.

**`WithRelationMapper(StringMapper)`** can be used in cases where the relation to be checked isn't known ahead of time. It
receives a function that takes the incoming request and returns the name of the relation or permission to check.

**`WithObjectType(string)`** sets the object type sent to the authorizer.

**`WithObjectID(string)`** sets the object ID sent to the authorizer.

**`WithObjectIDMapper(StringMapper)`** is used to determine the object ID sent to the authorizer at runtime. It receives
a function that takes the incoming request and returns an object ID.

**`WithObjectIDFromVar(string)`** (only in `gorillaz` and `ginz` middleware) configures the check call to use the value of
a path parameter as the object ID sent to the authorizer.

**`WithObjectMapper(ObjectMapper)`** can be used to set both the object type and ID at runtime. It receives a function that
takes the incoming request and returns a `(objectType string, objectID string)` pair.

**`WithPolicyPath(string)`** sets the name of the policy module to evaluate in check calls. It defaults to `check`.
If the `Policy` object used to construct the middleware contains the `Root` field, the root is used as a prefix.
For example, if the root is set to `"myPolicy"`, the `Check` call looks for a policy module named `myPolicy.check`.

### gRPC Middleware

The gRPC middleware is available in the sub-package `middleware/grpcz`.
It implements unary and stream gRPC server interceptors in its `.Unary()` and `.Stream()` methods.

```go
import (
	"github.com/aserto-dev/go-aserto/middleware"
	"github.com/aserto-dev/go-aserto/middleware/grpcz"
	"google.golang.org/grpc"
)
...
middleware, err := grpcz.New(
	azClient,
	middleware.Policy{
		Decision: 	   "allowed",
	},
)

server := grpc.NewServer(
	grpc.UnaryInterceptor(middleware.Unary),
	grpc.StreamInterceptor(middleware.Stream),
)
```

#### Mappers

In addition to the general `WithIdentityMapper`, `WithPolicyPathMapper`, and `WithResourceMapper`, the gRPC middleware
provides methods to help construct resource contexts from incoming messages.

**`WithResourceFromFields(fields ...string)`** selects a specified set of fields from the incoming message to be
included in the resource context.

**WithResourceFromMessageByPath(fieldsByPath map[string][]string, defaults ...string)** is similar to
`WithResourceFromFields` but can select different sets  of fields depending on which service method is called.

**WithResourceFromContextValue(ctxKey any, field string)** reads a value from the incoming request context
and adds it as a field to the resource context.

#### Default Mappers

The default behavior of the gRPC middleware is:

* Identity is pulled form the `"authorization"` metadata field (i.e. `middleware.Identity.FromMetadata("authorization")`).
* Policy path is constructed from `grpc.Method()` with dots (`.`) replacing path delimiters (`/`).
* No Resource Context is included in authorization calls by default.
