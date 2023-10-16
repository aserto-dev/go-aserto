package directory

import (
	"context"

	"github.com/aserto-dev/go-aserto/client/directory/internal"
)

func InternalConnect(ctx context.Context, conns *internal.Connections, cfg *Config) (*Client, error) {
	return connect(ctx, conns, cfg)
}
