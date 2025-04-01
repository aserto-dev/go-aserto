module github.com/aserto-dev/go-aserto/middleware/humaz

go 1.23.0

toolchain go1.23.7

replace github.com/aserto-dev/go-aserto => ../../

require (
	github.com/aserto-dev/errors v0.0.13
	github.com/aserto-dev/go-aserto v0.0.0-00010101000000-000000000000
	github.com/aserto-dev/go-authorizer v0.20.13
	github.com/danielgtaylor/huma/v2 v2.18.0
	github.com/lestrrat-go/jwx v1.2.26
	github.com/rs/zerolog v1.33.0
	google.golang.org/protobuf v1.36.5
)

require (
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.2.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.26.3 // indirect
	github.com/lestrrat-go/backoff/v2 v2.0.8 // indirect
	github.com/lestrrat-go/blackmagic v1.0.1 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/samber/lo v1.49.1 // indirect
	golang.org/x/crypto v0.35.0 // indirect
	golang.org/x/net v0.36.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250303144028-a0af3efb3deb // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250303144028-a0af3efb3deb // indirect
	google.golang.org/grpc v1.71.0 // indirect
)
