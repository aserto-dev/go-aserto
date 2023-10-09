package directory_test

import (
	"encoding/json"
	"testing"

	asserts "github.com/stretchr/testify/assert"

	"github.com/aserto-dev/go-aserto/client/directory"
)

const (
	base      = `{"address": "localhost:8282"}`
	noBase    = `{"reader": {"address": "localhost:9292"}, "writer": {"address": "localhost:9393"}}`
	overrides = `{"address": "localhost:8282", "reader": {"address": "localhost:9292"}}`
)

func TestUnmarshalConfig(t *testing.T) {
	t.Run("base only", func(t *testing.T) {
		assert := asserts.New(t)

		cfg := directory.Config{}
		if err := json.Unmarshal([]byte(base), &cfg); err != nil {
			assert.FailNow("failed to unmarshal config", err)
		}

		assert.NotNil(cfg.Config)
		assert.Equal("localhost:8282", cfg.Address)
		assert.Nil(cfg.Reader)

		assert.NoError(cfg.Validate())
	})

	t.Run("no base", func(t *testing.T) {
		assert := asserts.New(t)

		var cfg directory.Config
		if err := json.Unmarshal([]byte(noBase), &cfg); err != nil {
			assert.FailNow("failed to unmarshal config", err)
		}

		assert.Nil(cfg.Config)
		assert.Nil(cfg.Importer)
		assert.Nil(cfg.Exporter)
		assert.NotNil(cfg.Reader)
		assert.NotNil(cfg.Writer)
		assert.Equal("localhost:9292", cfg.Reader.Address)
		assert.Equal("localhost:9393", cfg.Writer.Address)
	})

	t.Run("overrides", func(t *testing.T) {
		assert := asserts.New(t)

		var cfg directory.Config
		if err := json.Unmarshal([]byte(overrides), &cfg); err != nil {
			assert.FailNow("failed to unmarshal config", err)
		}

		assert.NotNil(cfg.Config)
		assert.NotNil(cfg.Reader)
		assert.Nil(cfg.Writer)
		assert.Nil(cfg.Importer)
		assert.Nil(cfg.Exporter)
		assert.Equal("localhost:8282", cfg.Address)
		assert.Equal("localhost:9292", cfg.Reader.Address)
	})
}
