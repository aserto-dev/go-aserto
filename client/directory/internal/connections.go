package internal

import (
	"github.com/aserto-dev/go-aserto/client"
	hs "github.com/mitchellh/hashstructure/v2"
	"github.com/samber/lo"
	"google.golang.org/grpc"
)

type Connections struct {
	conns   map[uint64]*grpc.ClientConn
	Connect func(...client.ConnectionOption) (*grpc.ClientConn, error)
}

func NewConnections() *Connections {
	return &Connections{
		conns:   make(map[uint64]*grpc.ClientConn),
		Connect: client.NewConnection,
	}
}

func (cb *Connections) Get(cfg *client.Config) (*grpc.ClientConn, error) {
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

		conn, err = cb.Connect(opts...)
		if err != nil {
			return nil, err
		}

		cb.conns[hash] = conn
	}

	return conn, nil
}

func (cb *Connections) AsSlice() []*grpc.ClientConn {
	return lo.Values(cb.conns)
}

// Used for testing.
type ConnectCounter struct {
	Count int
}

func (cc *ConnectCounter) Connect(...client.ConnectionOption) (*grpc.ClientConn, error) {
	cc.Count++
	return &grpc.ClientConn{}, nil
}
