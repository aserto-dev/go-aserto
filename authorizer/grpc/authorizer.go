/*
Package grpc is used to create an AuthorizerClient that communicates with the authorizer using gRPC.

AuthorizerClient is the low-level interface that exposes the raw authorization API.
*/
package grpc

import (
	"context"

	"github.com/aserto-dev/go-aserto/client"
	"github.com/aserto-dev/go-aserto/client/authorizer"
	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"
)

// New returns a new gRPC AuthorizerClient with the specified options.
func New(ctx context.Context, opts ...client.ConnectionOption) (authz.AuthorizerClient, error) {
	c, err := authorizer.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return c.Authorizer, nil
}
