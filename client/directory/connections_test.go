package directory // nolint:testpackage

import (
	"context"
	"testing"

	asserts "github.com/stretchr/testify/assert"

	"github.com/aserto-dev/go-aserto/client"
)

type connectCounter struct {
	count int
}

func (cc *connectCounter) connect(context.Context, ...client.ConnectionOption) (*client.Connection, error) {
	cc.count++
	return &client.Connection{}, nil
}

func TestConnections(t *testing.T) {
	ctx := context.Background()

	counter := &connectCounter{}
	conns := newConnections()
	conns.connect = counter.connect

	cfg := &client.Config{Address: "localhost:8282"}

	t.Run("new connection", func(t *testing.T) {
		assert := asserts.New(t)

		conn, err := conns.Get(ctx, cfg)
		assert.NoError(err)
		assert.NotNil(conn)
		assert.Equal(1, counter.count)
	})

	t.Run("cached connection", func(t *testing.T) {
		assert := asserts.New(t)
		conn, err := conns.Get(ctx, cfg)
		assert.NoError(err)
		assert.NotNil(conn)
		assert.Equal(1, counter.count) // no new calls to `connect`
	})

	t.Run("second connection", func(t *testing.T) {
		assert := asserts.New(t)
		cfg := &client.Config{Address: "localhost:8282", TenantID: "foobar"}

		conn, err := conns.Get(ctx, cfg)
		assert.NoError(err)
		assert.NotNil(conn)
		assert.Equal(2, counter.count) // new call to `connect`
	})
}

func TestConnect(t *testing.T) {
	ctx := context.Background()

	t.Run("base only", func(t *testing.T) {
		assert := asserts.New(t)

		counter := &connectCounter{}
		conns := newConnections()
		conns.connect = counter.connect

		cfg := &Config{
			Config: &client.Config{Address: "localhost:8282"},
		}

		dir, err := connect(ctx, conns, cfg)
		assert.NoError(err)
		assert.NotNil(dir)
		assert.NotNil(dir.Reader)
		assert.NotNil(dir.Writer)
		assert.NotNil(dir.Importer)
		assert.NotNil(dir.Exporter)
		assert.Equal(1, counter.count)
	})

	t.Run("base with overrides", func(t *testing.T) {
		assert := asserts.New(t)

		counter := &connectCounter{}
		conns := newConnections()
		conns.connect = counter.connect

		cfg := &Config{
			Config: &client.Config{Address: "localhost:8282"},
			Reader: &client.Config{Address: "localhost:9292"},
		}

		dir, err := connect(ctx, conns, cfg)
		assert.NoError(err)
		assert.NotNil(dir)
		assert.NotNil(dir.Reader)
		assert.NotNil(dir.Writer)
		assert.NotNil(dir.Importer)
		assert.NotNil(dir.Exporter)
		assert.Equal(2, counter.count)
	})

	t.Run("no base", func(t *testing.T) {
		assert := asserts.New(t)

		counter := &connectCounter{}
		conns := newConnections()
		conns.connect = counter.connect

		cfg := &Config{
			Reader: &client.Config{Address: "localhost:9292"},
			Writer: &client.Config{Address: "localhost:9393"},
		}

		dir, err := connect(ctx, conns, cfg)
		assert.NoError(err)
		assert.NotNil(dir)
		assert.NotNil(dir.Reader)
		assert.NotNil(dir.Writer)
		assert.Nil(dir.Importer)
		assert.Nil(dir.Exporter)
		assert.Equal(2, counter.count)
	})
}
