module gorilla_example

go 1.26.0

toolchain go1.27.3

replace github.com/aserto-dev/go-aserto => ../../../..

replace github.com/aserto-dev/go-aserto/middleware/gorillaz => ../../../../middleware/gorillaz

require (
	github.com/aserto-dev/go-aserto v0.33.6
	github.com/aserto-dev/go-aserto/middleware/gorillaz v0.0.0-00010101000000-000000000000
	github.com/gorilla/mux v1.8.1
)

require (
	github.com/aserto-dev/errors v0.34.2-0.20261006083714-760eef8e0b6c // indirect
	github.com/aserto-dev/go-authorizer/aserto v0.24.4 // indirect
	github.com/aserto-dev/go-authorizer/pkg v0.24.3 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.1 // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.33.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/lestrrat-go/blackmagic v1.0.4 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc v1.0.6 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/jwx/v2 v2.1.7 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.16 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/zerolog v1.35.1 // indirect
	github.com/samber/lo v1.53.0 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20261104224025-d1e00e9f33ab // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261104224025-d1e00e9f33ab // indirect
	google.golang.org/grpc v1.85.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)
