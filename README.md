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

The [Authorizer](https://docs.aserto.com/docs/overview/authorizer) service is is an [open source authorization engine](https://www.topaz.sh)
which uses the [Open Policy Agent](https://www.openpolicyagent.org) (OPA) to make decisions by computing authorization
policies.


### AuthorizerClient

The `AuthorizerClient` interface, defined in
[`github.com/aserto-dev/go-authorizer/aserto/authorizer/v2`](https://github.com/aserto-dev/go-authorizer/blob/main/aserto/authorizer/v2/authorizer_grpc.pb.go#L34),
describes the operations exposed by the Aserto authorizer service.

Two implementation of `AuthorizerClient` are available:

1. `authorizer/grpc` provides a client that communicates with the authorizer using gRPC.

2. `authorizer/http` provides a client that communicates with the authorizer over its REST HTTP endpoints.


Create a new client using `New()` in either package.

The snippet below creates an authorizer client that talks to Aserto's hosted authorizer over gRPC:

```go
import (
	"github.com/aserto-dev/go-aserto/client"
	"github.com/aserto-dev/go-aserto/authorizer/grpc"
)
...
authorizer, err := grpc.New(
	ctx,
	client.WithAPIKeyAuth("<API Key>"),
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


#### Connection Timeout


Connection timeout can be set on the specified context using context.WithTimeout. If no timeout is set on the
context, the default connection timeout is 5 seconds. For example, to increase the timeout to 10 seconds:

```go
ctx := context.Background()

authorizer, err := grpc.New(
	context.WithTimeout(ctx, time.Duration(10) * time.Second),
	aserto.WithAPIKeyAuth("<API Key>"),
)
```


### Making Authorization Calls

Use the client's `Is()` method to request authorization decisions from the Aserto authorizer service.

```go
import (
	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
	"github.com/aserto-dev/go-authorizer/aserto/authorizer/v2/api"
)

resp, err := authorizer.Is(c.Context, &authz.IsRequest{
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

## Directory

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
	"context"
	"github.com/aserto-dev/go-aserto/client"
	"github.com/aserto-dev/go-aserto/client/directory/v3"
)

...

dir, err := directory.New(context.Background(), client.WithAPIKeyAuth('<api key>'))
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

To create a directory client from configuration, call `Connect(context.Context)` on the config struct:

```go
import (
	"context"

	"github.com/aserto-dev/go-aserto/client/directory/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/common/v2"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
)

...

// Use the same address for all services.
cfg := &directory.Config{Address: "localhost:9292"}

dir, err := cfg.Connect(context.Background())
if err != nil {
	panic(err)
}

resp, err := dir.Reader.GetObjects(&reader.GetObjectsRequest{})
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

Two middleware implementations are available in subpackages:

* `middleware/grpc` provides middleware for gRPC servers.
* `middleware/http` provides middleware for HTTP REST servers.

When authorization middleware is configured and attached to a server, it examines incoming requests, extracts
authorization parameters like the caller's identity, calls the Aserto authorizers, and rejects messages if their
access is denied.

Both gRPC and HTTP middleware are created from an `AuthorizerClient` and a `Policy` with parameters that can be shared
by all authorization calls.

```go
// Policy holds global authorization options that apply to all requests.
type Policy struct {
	// Name is the Name of the aserto policy being queried for authorization.
	Name string

	// Path is the package name of the rego policy to evaluate.
	// If left empty, a policy mapper must be attached to the middleware to provide
	// the policy path from incoming messages.
	Path string

	// Decision is the authorization rule to use.
	Decision string

	// Label name of the aserto policy's instance being queried for authorization.
	InstanceLabel string
}
```

The value of several authorization parameters often depends on the content of incoming requests. Those are:

* Identity - the identity (subject or JWT) of the caller.
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
```

In addition, it is possible to provide custom logic to specify the callers identity. For example, in HTTP middleware:

```go
middleware.Identity.Mapper(func(r *http.Request, identity middleware.Identity) {
	username := getUserFromRequest(r) // custom logic to get user identity

	identity.Subject().ID(username) // set it on the middleware
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
middleware.WithPolicyPathMapper(func(ctx context.Context, req interface{}) string {
	path := getPolicyPath(ctx, req) // custom logic to retrieve a JWT token
	return path
})
```

### Resource

A resource can be any structured data that the authorization policy uses to evaluate decisions.
By default, middleware do not include a resource in authorization calls.

To add resource data, use `Middleware.WithResourceMapper()` to attach custom logic. For example, in HTTP middleware:

```go
middleware.WithResourceMapper(func(r *http.Request) *structpb.Struct {
	return structFromBody(r.Body) // custom logic
})
```

In addition to these, each middleware has built-in mappers that can handle common use-cases.

### gRPC Middleware

The gRPC middleware is available in the sub-package `middleware/grpc`.
It implements unary and stream gRPC server interceptors in its `.Unary()` and `.Stream()` methods.

```go
import (
	"github.com/aserto-dev/go-aserto/middleware"
	grpcmw "github.com/aserto-dev/go-aserto/middleware/grpc"
	"google.golang.org/grpc"
)
...
middleware, err := grpcmw.New(
	client,
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

gRPC mappers take as their input the incoming request's context and the message.

```go
type (
	// StringMapper functions are used to extract string values like identity and policy-path from incoming messages.
	StringMapper func(context.Context, interface{}) string

	// ResourceMapper functions are used to extract structured data from incoming message.
	ResourceMapper func(context.Context, interface{}, map[string]interface{})
)
```

In addition to the general `WithIdentityMapper`, `WithPolicyPathMapper`, and `WithResourceMapper`, the gRPC middleware
provides methods to help construct resource contexts from incoming messages.

**`WithResourceFromFields(fields ...string)`** selects a specified set of fields from the incoming message to be
included in the authorization resource.

**WithResourceFromMessageByPath(fieldsByPath map[string][]string, defaults ...string)** is similar to
`WithResourceFromFields` but can select different sets  of fields depending on which service method is called.

**WithResourceFromContextValue(ctxKey interface{}, field string)** reads a value from the incoming request context
and adds it as a field to the resource context.

#### Default Mappers

The default behavior of the gRPC middleware is:

* Identity is pulled form the `"authorization"` metadata field (i.e. `middleware.Identity.FromMetadata("authorization")`).
* Policy path is constructed from `grpc.Method()` with dots (`.`) replacing path delimiters (`/`).
* No Resource Context is included in authorization calls by default.


### HTTP Middleware

The HTTP middleware are available under the sub-package `middleware/http`.

Two flavors are implemented:

* Standard `net/http` middleware with support for [gorilla/mux](https://pkg.go.dev/github.com/gorilla/mux) is implemented in `middleware/http/std`.
* [Gin](https://github.com/gin-gonic/gin) middleware is implemented in `middleware/http/gin`.

Both are constructed and configured in a similar way. They differ in the signature of their `Handler()`
function, which is used to attach them to HTTP routes, and in the signatures of their mapper functions.

#### net/http Middleware

```go
import (
	"github.com/aserto-dev/go-aserto/middleware"
	"github.com/aserto-dev/go-aserto/middleware/http/std"
)
...
mw := std.New(
	client,
	middleware.Policy{
		Decision:	   "allowed",
	},
)
```

Adding the created authorization middleware to a basic `net/http` server may look something like this:

```go
http.Handle("/foo", mw.Handler(fooHandler))
```

The popular [`gorilla/mux`](https://github.com/gorilla/mux) package provides a powerful and flexible HTTP router.
Attaching the standard authorization middleware to a `gorilla/mux` server is as simple as:

```go
router := mux.NewRouter()
router.Use(mw.Handler)

router.HandleFunc("/foo", fooHandler).Methods("GET")
```


#### Mappers

HTTP mappers take the incoming `http.Request` as their sole parameter.

```go
type (
	StringMapper func(*http.Request) string
	StructMapper func(*http.Request) *structpb.Struct
)
```

In addition to the general `WithIdentityMapper`, `WithPolicyMapper`, and `WithResourceMapper`, the HTTP middleware
provides `WithIdentityFromHeader()` to extract identity information from HTTP headers, and `WithNoResourceContext()` to
omit a resource context from authorization calls.

#### Default Mappers

The default behavior of the HTTP middleware is:

* Identity is retrieved from the "Authorization" HTTP Header, if present.
* Policy path is retrieved from the request URL and method to form a path of the form `METHOD.path.to.endpoint`.
  If the server uses [`gorilla/mux`](https://github.com/gorilla/mux) and
  the route contains path parameters (e.g. `"api/products/{id}"`), the surrounding braces are replaced with a
  double-underscore prefix. For example, with policy root `"myApp"`, a request to `GET api/products/{id}` gets the
  policy path `myApp.GET.api.products.__id`.
* Any path parameters defined using [`gorilla/mux`](https://github.com/gorilla/mux) are included in the resource
  context. For example, if the route is defined as `"api/products/{id}"` and the incoming request URL path is
  `"api/products/123"` then the resource context will be `{"id": "123"}`.


#### Gin Middleware

The gin middleware looks and behaves just like the net/http middleware with the following differences:

* Its Handler function is a `gin.HandlerFunc` which can be used with
[`IRoutes.Use(...HandlerFunc)`](https://pkg.go.dev/github.com/gin-gonic/gin#IRoutes).
* Its mappers take `*gin.Context` instead of `*http.Request`:
  ```go
	type (
		StringMapper func(*gin.Context) string
		StructMapper func(*gin.Context) *structpb.Struct
	)
  ```

### Check Middleware (ReBAC)

In addition to the pattern described above, in which each route is authorized by its own policy module,
the HTTP middleware can be used to implement Relation-Based Access Control (rebac) in which authorization
decisions are made by checking if a given subject has the necessary permission or relation to the object being accessed.

This is achieved using the `Check` function on `http.Middleware`.

A check call needs three pieces of information:

* The type and key of the object.
* The name of the relation or permission to look for.
* The type and key of the subject. When omitted, the subject is derived from the middleware's [Identity](#identity)
with type `"user"`.

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


## Other Aserto Services

In addition to the authorizer service, go-aserto provides gRPC clients for Aserto's administrative services,
allowing users to programmatically manage their aserto account.

`client/authorizer` defines a client for services run at the edge and used to serve authorization requests.


An API client is created using `New()` with the same connection options as the authorizer client.

### Edge Client

```go
// Client provides access to Aserto edge services.
type Client struct {
	// Authorizer provides methods for performing authorization requests.
	Authorizer authorizer.AuthorizerClient
}
```
