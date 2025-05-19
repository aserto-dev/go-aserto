package ds

import (
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"google.golang.org/grpc"

	"github.com/aserto-dev/go-aserto"
	"github.com/aserto-dev/go-aserto/ds/internal"
	"github.com/aserto-dev/go-directory/aserto/directory/reader/v3"
	"github.com/aserto-dev/go-directory/aserto/directory/writer/v3"
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
	if allNil([]*aserto.Config{c.Config, c.Reader, c.Writer}) {
		return ErrInvalidConfig
	}

	return nil
}

func connect(conns *internal.Connections, cfg *Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	r, err := getConnection(conns, cfg.Reader, cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "reader connection failed")
	}

	w, err := getConnection(conns, cfg.Writer, cfg.Config)
	if err != nil {
		return nil, errors.Wrap(err, "writer connection failed")
	}

	return &Client{
		Reader: newClient(r, reader.NewReaderClient),
		Writer: newClient(w, writer.NewWriterClient),
		conns:  conns.AsSlice(),
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
