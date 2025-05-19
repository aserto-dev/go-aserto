module gorilla_example

go 1.23.0

toolchain go1.24.3

replace github.com/aserto-dev/go-aserto => ../../../..

replace github.com/aserto-dev/go-aserto/middleware/gorillaz => ../../../../middleware/gorillaz

require (
	github.com/aserto-dev/go-aserto v0.33.9
	github.com/aserto-dev/go-aserto/middleware/gorillaz v0.0.0-20250430215048-a6f7d0d57d40
	github.com/gorilla/mux v1.8.1
)

require (
	github.com/aserto-dev/errors v0.0.17 // indirect
	github.com/aserto-dev/go-authorizer v0.20.14 // indirect
	github.com/aserto-dev/header v0.0.11 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/go-http-utils/headers v0.0.0-20181008091004-fed159eddc2a // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/lestrrat-go/blackmagic v1.0.2 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc v1.0.6 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/jwx/v2 v2.1.4 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	github.com/samber/lo v1.50.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250512202823-5a2f75b736a9 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250512202823-5a2f75b736a9 // indirect
	google.golang.org/grpc v1.72.1 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
