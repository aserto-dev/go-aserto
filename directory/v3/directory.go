package directory

import (
	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/go-aserto/internal/hosted"
	des "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dis "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dms "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	drs "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dws "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// Client provides access to the Aserto Directory APIs.
type Client struct {
	// Client for the directory reader service.
	Reader drs.ReaderClient

	// Client for the directory writer service.
	Writer dws.WriterClient

	// Client for the directory importer service.
	Importer dis.ImporterClient

	// Client for the directory exporter service.
	Exporter des.ExporterClient

	// Client for the directory model service.
	Model dms.ModelClient

	conns []*grpc.ClientConn
}

// New returns a new Directory with the specified options.
func New(opts ...aserto.ConnectionOption) (*Client, error) {
	options, err := aserto.NewConnectionOptions(opts...)
	if err != nil {
		return nil, err
	}

	if options.ServerAddress() == "" {
		options.Address = hosted.HostedDirectoryHostname + hosted.HostedDirectoryGRPCPort
	}

	conn, err := aserto.Connect(options)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return &Client{
		Reader:   drs.NewReaderClient(conn),
		Writer:   dws.NewWriterClient(conn),
		Importer: dis.NewImporterClient(conn),
		Exporter: des.NewExporterClient(conn),
		Model:    dms.NewModelClient(conn),
		conns:    []*grpc.ClientConn{conn},
	}, nil
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
