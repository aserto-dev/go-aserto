/*
Package client provides communication with the Aserto services.

There are two groups of services:

1. client/authorizer provides access to the authorizer service and the edge services running alongside it.

2. client/tenant provides access to the Aserto control plane services.
*/
package aserto

import (
	"context"
	"strings"

	"github.com/aserto-dev/go-aserto/internal/hosted"
	"github.com/aserto-dev/header"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// NewConnection creates a gRPC connection with the given options.
func NewConnection(opts ...ConnectionOption) (*grpc.ClientConn, error) {
	options, err := NewConnectionOptions(opts...)
	if err != nil {
		return nil, err
	}

	if options.Address == "" {
		// Backward compatibility: default to authorizer service.
		options.Address = hosted.HostedAuthorizerHostname + hosted.HostedAuthorizerGRPCPort
	}

	return Connect(options)
}

// Connect creates a gRPC connection with the given options.
func Connect(options *ConnectionOptions) (*grpc.ClientConn, error) {
	if options.Address == "" {
		return nil, errors.Wrap(ErrInvalidOptions, "address not specified")
	}

	dialOpts, err := options.ToDialOptions()
	if err != nil {
		return nil, err
	}

	return grpc.NewClient(options.Address, dialOpts...)
}

// SetTenantContext returns a new context with the provided tenant ID embedded as metadata.
func SetTenantContext(ctx context.Context, tenantID string) context.Context {
	if strings.TrimSpace(tenantID) == "" {
		return ctx
	}

	return metadata.AppendToOutgoingContext(ctx, string(header.HeaderAsertoTenantID), tenantID)
}

func SetAccountContext(ctx context.Context, accountID string) context.Context {
	if strings.TrimSpace(accountID) == "" {
		return ctx
	}

	return metadata.AppendToOutgoingContext(ctx, string(header.HeaderAsertoAccountID), accountID)
}
