package ds

import (
	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/go-aserto/internal/hosted"

	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Client provides access to the Aserto Directory APIs.
type Client struct {
	// Client for the directory reader service.
	Reader reader.ReaderClient

	// Client for the directory writer service.
	Writer writer.WriterClient

	conns []*grpc.ClientConn
}

// New returns a new Client with the specified options.
func New(opts ...aserto.ConnectionOption) (*Client, error) {
	options, err := aserto.NewConnectionOptions(opts...)
	if err != nil {
		return nil, err
	}

	if options.Address == "" {
		options.Address = hosted.HostedDirectoryHostname + hosted.HostedDirectoryGRPCPort
	}

	conn, err := aserto.Connect(options)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return &Client{
		Reader: reader.NewReaderClient(conn),
		Writer: writer.NewWriterClient(conn),
		conns:  []*grpc.ClientConn{conn},
	}, nil
}

// FromConnection returns a new Client using an existing connection.
func FromConnection(conn *grpc.ClientConn) *Client {
	return &Client{
		Reader: reader.NewReaderClient(conn),
		Writer: writer.NewWriterClient(conn),
		conns:  []*grpc.ClientConn{conn},
	}
}

// Close closes the underlying connections.
func (c *Client) Close() error {
	var errs error

	for _, conn := range c.conns {
		if err := conn.Close(); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}
