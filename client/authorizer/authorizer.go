package authorizer

import (
	"context"

	"github.com/aserto-dev/go-aserto/client"
	"google.golang.org/grpc"

	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"

	"github.com/pkg/errors"
)

// Client provides access to the Aserto authorization services.
type Client struct {
	conn *grpc.ClientConn

	// Authorizer provides methods for performing authorization requests.
	Authorizer authz.AuthorizerClient
}

// NewClient creates a Client with the specified connection options.
func New(ctx context.Context, opts ...client.ConnectionOption) (*Client, error) {
	conn, err := client.NewConnection(opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return &Client{
		conn:       conn,
		Authorizer: authz.NewAuthorizerClient(conn),
	}, err
}

// Close closes the underlying connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Connection returns the underlying grpc connection.
func (c *Client) Connection() grpc.ClientConnInterface {
	return c.conn
}
