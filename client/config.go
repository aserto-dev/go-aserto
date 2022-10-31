package client

import (
	"github.com/aserto-dev/aserto-grpc/grpcutil/middlewares/request"
	"github.com/pkg/errors"
)

type Config struct {
	Address          string            `json:"address"`
	CACertPath       string            `json:"ca_cert_path"`
	ClientCertPath   string            `json:"client_cert_path"`
	ClientKeyPath    string            `json:"client_key_path"`
	APIKey           string            `json:"api_key"`
	Insecure         bool              `json:"insecure"`
	TimeoutInSeconds int               `json:"timeout_in_seconds"`
	Token            string            `json:"token"`
	Headers          map[string]string `json:"headers"`
}

func (cfg *Config) ToClientOptions(dop DialOptionsProvider) ([]ConnectionOption, error) {
	middleware := request.NewRequestIDMiddleware()
	options := []ConnectionOption{
		WithChainUnaryInterceptor(middleware.UnaryClient()),
		WithChainStreamInterceptor(middleware.StreamClient()),
		WithInsecure(cfg.Insecure),
	}

	if cfg.APIKey != "" && cfg.Token != "" {
		return nil, errors.New("both api_key and token are set")
	}

	if cfg.Token != "" {
		options = append(options, WithTokenAuth(cfg.Token))
	}

	if cfg.APIKey != "" {
		options = append(options, WithAPIKeyAuth(cfg.APIKey))
	}

	if cfg.Address != "" {
		options = append(options, WithAddr(cfg.Address))
	}

	if cfg.CACertPath != "" {
		options = append(options, WithCACertPath(cfg.CACertPath))
	}

	opts, err := dop(cfg)
	if err != nil {
		return nil, err
	}

	options = append(options, WithDialOptions(opts...))

	return options, nil
}
