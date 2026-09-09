module github.com/aserto-dev/go-aserto/middleware/humaz

go 1.26.3

toolchain go1.27.1

replace github.com/aserto-dev/go-aserto => ../../

require (
	github.com/aserto-dev/errors v0.34.1
	github.com/aserto-dev/go-aserto v0.0.0-00010101000000-000000000000
	github.com/aserto-dev/go-authorizer v0.24.1
	github.com/danielgtaylor/huma/v2 v2.37.3
	github.com/lestrrat-go/jwx v1.2.31
	github.com/rs/zerolog v1.35.1
	google.golang.org/protobuf v1.36.12
)

require (
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.1 // indirect
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.30.0 // indirect
	github.com/lestrrat-go/backoff/v2 v2.0.8 // indirect
	github.com/lestrrat-go/blackmagic v1.0.4 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/samber/lo v1.53.0 // indirect
	golang.org/x/crypto v0.57.0 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260908043556-f8649ddbbfe6 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260908043556-f8649ddbbfe6 // indirect
	google.golang.org/grpc v1.83.2 // indirect
)
