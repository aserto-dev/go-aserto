package directory

import (
	"context"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"google.golang.org/grpc"

	"github.com/aserto-dev/go-aserto/client"
	"github.com/aserto-dev/go-aserto/client/directory/internal"
	des "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dis "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	drs "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dws "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
)

var (
	ErrInvalidConfig = errors.New("invalid config")
)

// Config provides configuration for connecting to the Aserto Directory service.
type Config struct {
	// Base configuration. If non-nil, this configuration is used for any client that doesn't have its own configuration.
	// If nil, only clients that have their own configuration will be created.
	*client.Config

	// Reader configuration.
	Reader *client.Config `json:"reader"`

	// Writer configuration.
	Writer *client.Config `json:"writer"`

	// Importer configuration.
	Importer *client.Config `json:"importer"`

	// Exporter configuration.
	Exporter *client.Config `json:"exporter"`
}

// Connect create a new directory client from the specified configuration.
func (c *Config) Connect(ctx context.Context) (*Client, error) {
	return connect(ctx, internal.NewConnections(), c)
}

// Validate returns an error if the configuration is invalid.
func (c *Config) Validate() error {
	if c == nil {
		return ErrInvalidConfig
	}

	// At least one client config must be non-nil.
	if allNil([]*client.Config{c.Config, c.Reader, c.Writer, c.Importer, c.Exporter}) {
		return ErrInvalidConfig
	}

	return nil
}

func connect(ctx context.Context, conns *internal.Connections, cfg *Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	reader, err := getConnection(ctx, conns, cfg.Reader, cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "reader connection failed")
	}

	writer, err := getConnection(ctx, conns, cfg.Writer, cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "writer connection failed")
	}

	importer, err := getConnection(ctx, conns, cfg.Importer, cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "importer connection failed")
	}

	exporter, err := getConnection(ctx, conns, cfg.Exporter, cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "exporter connection failed")
	}

	return &Client{
		Reader:   newClient(reader, drs.NewReaderClient),
		Writer:   newClient(writer, dws.NewWriterClient),
		Importer: newClient(importer, dis.NewImporterClient),
		Exporter: newClient(exporter, des.NewExporterClient),
	}, nil
}

// Returns true if all elements of slice are nil.
func allNil[T any](slice []*T) bool {
	return lo.Every([]*T{nil}, slice)
}

func getConnection(
	ctx context.Context,
	conns *internal.Connections,
	cfg, fallback *client.Config,
) (*client.Connection, error) {
	if cfg != nil {
		return conns.Get(ctx, cfg)
	}

	if fallback != nil {
		return conns.Get(ctx, fallback)
	}

	return nil, nil
}

func newClient[T any](conn *client.Connection, factory func(conn grpc.ClientConnInterface) T) T {
	if conn == nil {
		var t T
		return t
	}

	return factory(conn.Conn)
}
