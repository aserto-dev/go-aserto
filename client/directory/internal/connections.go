package internal

import (
	"context"

	"github.com/aserto-dev/go-aserto/client"
	hs "github.com/mitchellh/hashstructure/v2"
	"github.com/samber/lo"
	"google.golang.org/grpc"
)

type Connections struct {
	conns   map[uint64]*client.Connection
	Connect func(context.Context, ...client.ConnectionOption) (*client.Connection, error)
}

func NewConnections() *Connections {
	return &Connections{
		conns:   make(map[uint64]*client.Connection),
		Connect: client.NewConnection,
	}
}

func (cb *Connections) Get(ctx context.Context, cfg *client.Config) (*client.Connection, error) {
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

		conn, err = cb.Connect(ctx, opts...)
		if err != nil {
			return nil, err
		}

		cb.conns[hash] = conn
	}

	return conn, nil
}

func (cb *Connections) AsSlice() []*grpc.ClientConn {
	return lo.MapToSlice(cb.conns, func(_ uint64, conn *client.Connection) *grpc.ClientConn {
		return conn.Conn
	})
}

// Used for testing.
type ConnectCounter struct {
	Count int
}

func (cc *ConnectCounter) Connect(context.Context, ...client.ConnectionOption) (*client.Connection, error) {
	cc.Count++
	return &client.Connection{}, nil
}
