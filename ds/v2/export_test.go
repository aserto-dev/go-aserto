package ds

import (
	"github.com/aserto-dev/go-aserto/ds/internal"
)

func InternalConnect(conns *internal.Connections, cfg *Config) (*Client, error) {
	return connect(conns, cfg)
}
