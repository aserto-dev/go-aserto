module github.com/aserto-dev/go-aserto/middleware/grpcz

go 1.22.10

toolchain go1.23.4

replace github.com/aserto-dev/go-aserto => ../../

require (
	github.com/aserto-dev/errors v0.0.12
	github.com/aserto-dev/go-aserto v0.33.4
	github.com/aserto-dev/go-authorizer v0.20.12
	github.com/aserto-dev/go-directory v0.33.3
	github.com/lestrrat-go/jwx/v2 v2.1.3
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.33.0
	github.com/samber/lo v1.47.0
	github.com/stretchr/testify v1.10.0
	google.golang.org/grpc v1.68.1
	google.golang.org/protobuf v1.36.0
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.35.2-20241127180247-a33202765966.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.3.0 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.24.0 // indirect
	github.com/lestrrat-go/blackmagic v1.0.2 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc v1.0.6 // indirect
	github.com/lestrrat-go/iter v1.0.2 // indirect
	github.com/lestrrat-go/option v1.0.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241209162323-e6fa225c2576 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241216192217-9240e9c98484 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
