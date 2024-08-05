package directory

import (
	"github.com/aserto-dev/go-aserto/client/directory/internal"
)

func InternalConnect(conns *internal.Connections, cfg *Config) (*Client, error) {
	return connect(conns, cfg)
}
