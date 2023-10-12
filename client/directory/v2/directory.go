package directory

import (
	"context"

	"github.com/aserto-dev/go-aserto/client"
	"github.com/aserto-dev/go-aserto/internal/hosted"
	des "github.com/aserto-dev/go-directory/aserto/directory/exporter/v2"
	dis "github.com/aserto-dev/go-directory/aserto/directory/importer/v2"
	drs "github.com/aserto-dev/go-directory/aserto/directory/reader/v2"
	dws "github.com/aserto-dev/go-directory/aserto/directory/writer/v2"
	"github.com/pkg/errors"
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
}

// New returns a new Directory with the specified options.
func New(ctx context.Context, opts ...client.ConnectionOption) (*Client, error) {
	options, err := client.NewConnectionOptions(opts...)
	if err != nil {
		return nil, err
	}

	if options.ServerAddress() == "" {
		options.Address = hosted.HostedDirectoryHostname + hosted.HostedDirectoryGRPCPort
	}

	connection, err := client.Connect(ctx, options)
	if err != nil {
		return nil, errors.Wrap(err, "create grpc client failed")
	}

	return &Client{
		Reader:   drs.NewReaderClient(connection.Conn),
		Writer:   dws.NewWriterClient(connection.Conn),
		Importer: dis.NewImporterClient(connection.Conn),
		Exporter: des.NewExporterClient(connection.Conn),
	}, nil
}
