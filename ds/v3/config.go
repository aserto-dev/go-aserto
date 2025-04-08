package ds

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"google.golang.org/grpc"

	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/go-aserto/ds/internal"
	des "github.com/aserto-dev/go-directory/aserto/directory/exporter/v3"
	dis "github.com/aserto-dev/go-directory/aserto/directory/importer/v3"
	dms "github.com/aserto-dev/go-directory/aserto/directory/model/v3"
	drs "github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	dws "github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
)

var ErrInvalidConfig = errors.New("invalid config")

// Config provides configuration for connecting to the Aserto Directory service.
type Config struct {
	// Base configuration. If non-nil, this configuration is used for any client that doesn't have its own configuration.
	// If nil, only clients that have their own configuration will be created.
	*aserto.Config

	// Reader configuration.
	Reader *aserto.Config `json:"reader"`

	// Writer configuration.
	Writer *aserto.Config `json:"writer"`

	// Importer configuration.
	Importer *aserto.Config `json:"importer"`

	// Exporter configuration.
	Exporter *aserto.Config `json:"exporter"`

	// Model configuration.
	Model *aserto.Config `json:"model"`
}

// Connect create a new directory client from the specified configuration.
func (c *Config) Connect() (*Client, error) {
	return connect(internal.NewConnections(), c)
}

// Validate returns an error if the configuration is invalid.
func (c *Config) Validate() error {
	if c == nil {
		return ErrInvalidConfig
	}

	// At least one client config must be non-nil.
	if allNil([]*aserto.Config{c.Config, c.Reader, c.Writer, c.Importer, c.Exporter}) {
		return ErrInvalidConfig
	}

	return nil
}

func connect(conns *internal.Connections, cfg *Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	reader, err := getConnection(conns, cfg.Reader, cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "reader connection failed")
	}

	writer, err := getConnection(conns, cfg.Writer, cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "writer connection failed")
	}

	importer, err := getConnection(conns, cfg.Importer, cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "importer connection failed")
	}

	exporter, err := getConnection(conns, cfg.Exporter, cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "exporter connection failed")
	}

	model, err := getConnection(conns, cfg.Model, cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "model connection failed")
	}

	return &Client{
		Reader:   newClient(reader, drs.NewReaderClient),
		Writer:   newClient(writer, dws.NewWriterClient),
		Importer: newClient(importer, dis.NewImporterClient),
		Exporter: newClient(exporter, des.NewExporterClient),
		Model:    newClient(model, dms.NewModelClient),
		conns:    conns.AsSlice(),
	}, nil
}

// Returns true if all elements of slice are nil.
func allNil[T any](slice []*T) bool {
	return lo.Every([]*T{nil}, slice)
}

func getConnection(
	conns *internal.Connections,
	cfg, fallback *aserto.Config,
) (*grpc.ClientConn, error) {
	if cfg != nil {
		return conns.Get(cfg)
	}

	if fallback != nil {
		return conns.Get(fallback)
	}

	return nil, nil //nolint: nilnil
}

func newClient[T any](conn *grpc.ClientConn, factory func(conn grpc.ClientConnInterface) T) T {
	if conn == nil {
		var t T
		return t
	}

	return factory(conn)
}
