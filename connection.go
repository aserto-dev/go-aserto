/*
Package client provides communication with the Aserto services.

There are two groups of services:

1. client/authorizer provides access to the authorizer service and the edge services running alongside it.

2. client/tenant provides access to the Aserto control plane services.
*/
package aserto

import (
	"context"
	"crypto/tls"
	"strings"

	"github.com/aserto-dev/go-aserto/internal/hosted"
	"github.com/aserto-dev/go-aserto/internal/tlsconf"
	"github.com/aserto-dev/header"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

/*
NewConnection establishes a gRPC connection.

# Options

Options can be specified to configure the connection or override default behavior:

1. WithAddr() - sets the server address and port. Default: "authorizer.prod.aserto.com:8443".

2. WithAPIKeyAuth() - sets an API key for authentication.

3. WithTokenAuth() - sets an OAuth2 token to be used for authentication.

4. WithTenantID() - sets the aserto tenant ID.

5. WithInsecure() - enables/disables TLS verification. Default: false.

6. WithCACertPath() - adds the specified PEM certificate file to the connection's list of trusted root CAs.

# Timeout

Connection timeout can be set on the specified context using context.WithTimeout. If no timeout is set on the
context, the default connection timeout is 5 seconds. For example, to increase the timeout to 10 seconds:

	ctx := context.Background()

	client, err := authorizer.New(
		context.WithTimeout(ctx, time.Duration(10) * time.Second),
		aserto.WithAPIKeyAuth("<API Key>"),
		aserto.WithTenantID("<Tenant ID>"),
	)
*/
func NewConnection(opts ...ConnectionOption) (*grpc.ClientConn, error) {
	options, err := NewConnectionOptions(opts...)
	if err != nil {
		return nil, err
	}

	if options.ServerAddress() == "" {
		// Backward compatibility: default to authorizer service.
		options.Address = hosted.HostedAuthorizerHostname + hosted.HostedAuthorizerGRPCPort
	}

	return Connect(options)
}

func Connect(options *ConnectionOptions) (*grpc.ClientConn, error) {
	return newConnection(newClient, options)
}

// connectionFactory is introduced in order to test the logic responsible for configuring the underlying gRPC connection
// without really attempting a connection.
type connectionFactory func(
	address string,
	tlsConf *tls.Config,
	callerCreds credentials.PerRPCCredentials,
	tenantID, sessionID string,
	options []grpc.DialOption,
) (*grpc.ClientConn, error)

// newClient is the default dialer that calls grpc.DialContext to establish a connection.
func newClient(
	address string,
	tlsConf *tls.Config,
	callerCreds credentials.PerRPCCredentials,
	tenantID, sessionID string,
	options []grpc.DialOption,
) (*grpc.ClientConn, error) {
	if address == "" {
		return nil, errors.Wrap(ErrInvalidOptions, "address not specified")
	}

	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConf)),
		grpc.WithChainUnaryInterceptor(unary(tenantID, sessionID)),
		grpc.WithChainStreamInterceptor(stream(tenantID, sessionID)),
	}
	if callerCreds != nil {
		dialOptions = append(dialOptions, grpc.WithPerRPCCredentials(callerCreds))
	}

	dialOptions = append(dialOptions, options...)

	return grpc.NewClient(
		address,
		dialOptions...,
	)
}

func newConnection(dialContext connectionFactory, options *ConnectionOptions) (*grpc.ClientConn, error) {
	tlsConf, err := tlsconf.TLSConfig(options.Insecure, options.CACertPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup tls configuration")
	}

	dialOptions := []grpc.DialOption{
		grpc.WithChainStreamInterceptor(options.StreamClientInterceptors...),
		grpc.WithChainUnaryInterceptor(options.UnaryClientInterceptors...),
	}

	dialOptions = append(dialOptions, options.DialOptions...)

	return dialContext(
		options.ServerAddress(),
		tlsConf,
		options.Creds,
		options.TenantID,
		options.SessionID,
		dialOptions,
	)
}

func unary(tenantID, sessionID string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		return invoker(SetTenantContext(SetSessionContext(ctx, sessionID), tenantID), method, req, reply, cc, opts...)
	}
}

func stream(tenantID, sessionID string) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		return streamer(SetTenantContext(SetSessionContext(ctx, sessionID), tenantID), desc, cc, method, opts...)
	}
}

// SetTenantContext returns a new context with the provided tenant ID embedded as metadata.
func SetTenantContext(ctx context.Context, tenantID string) context.Context {
	if strings.TrimSpace(tenantID) == "" {
		return ctx
	}

	return metadata.AppendToOutgoingContext(ctx, string(header.HeaderAsertoTenantID), tenantID)
}

func SetSessionContext(ctx context.Context, sessionID string) context.Context {
	if strings.TrimSpace(sessionID) == "" {
		return ctx
	}

	return metadata.AppendToOutgoingContext(ctx, string(header.HeaderAsertoSessionID), sessionID)
}
