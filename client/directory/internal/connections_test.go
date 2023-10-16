package internal // nolint:testpackage

import (
	"context"
	"testing"

	asserts "github.com/stretchr/testify/assert"

	"github.com/aserto-dev/go-aserto/client"
)

func TestConnections(t *testing.T) {
	ctx := context.Background()

	counter := &ConnectCounter{}
	conns := NewConnections()
	conns.Connect = counter.Connect

	cfg := &client.Config{Address: "localhost:8282"}

	t.Run("new connection", func(t *testing.T) {
		assert := asserts.New(t)

		conn, err := conns.Get(ctx, cfg)
		assert.NoError(err)
		assert.NotNil(conn)
		assert.Equal(1, counter.Count)
	})

	t.Run("cached connection", func(t *testing.T) {
		assert := asserts.New(t)
		conn, err := conns.Get(ctx, cfg)
		assert.NoError(err)
		assert.NotNil(conn)
		assert.Equal(1, counter.Count) // no new calls to `connect`
	})

	t.Run("second connection", func(t *testing.T) {
		assert := asserts.New(t)
		cfg := &client.Config{Address: "localhost:8282", TenantID: "foobar"}

		conn, err := conns.Get(ctx, cfg)
		assert.NoError(err)
		assert.NotNil(conn)
		assert.Equal(2, counter.Count) // new call to `connect`
	})
}
