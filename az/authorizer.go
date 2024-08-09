package az

import (
	"github.com/aserto-dev/go-aserto"
	"google.golang.org/grpc"

	authz "github.com/aserto-dev/go-authorizer/aserto/authorizer/v2"

	"github.com/pkg/errors"
)

// Client provides access to the Aserto authorization services.
type Client struct {
	authz.AuthorizerClient
	conn *grpc.ClientConn
}

// NewClient creates a Client with the specified connection options.
func New(opts ...aserto.ConnectionOption) (*Client, error) {
	conn, err := aserto.NewConnection(opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return &Client{
		AuthorizerClient: authz.NewAuthorizerClient(conn),
		conn:             conn,
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
