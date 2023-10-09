package directory

import (
	"context"

	"github.com/aserto-dev/go-aserto/client"
	hs "github.com/mitchellh/hashstructure/v2"
)

type connections struct {
	conns   map[uint64]*client.Connection
	connect func(context.Context, ...client.ConnectionOption) (*client.Connection, error)
}

func newConnections() *connections {
	return &connections{
		conns:   make(map[uint64]*client.Connection),
		connect: client.NewConnection,
	}
}

func (cb *connections) Get(ctx context.Context, cfg *client.Config) (*client.Connection, error) {
	if cfg == nil {
		return nil, nil
	}

	hash, err := hs.Hash(cfg, hs.FormatV2, nil)
	if err != nil {
		return nil, err
	}

	conn := cb.conns[hash]
	if conn == nil {
		dop := client.NewDialOptionsProvider()

		opts, err := cfg.ToConnectionOptions(dop)
		if err != nil {
			return nil, err
		}

		conn, err = cb.connect(ctx, opts...)
		if err != nil {
			return nil, err
		}

		cb.conns[hash] = conn
	}

	return conn, nil
}
