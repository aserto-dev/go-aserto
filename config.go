package aserto

import "github.com/pkg/errors"

var ErrInvalidConfig = errors.New("invalid configuration")

// gRPC Client Configuration.
type Config struct {
	Address          string            `json:"address"`
	Token            string            `json:"token"`
	TenantID         string            `json:"tenant_id"`
	APIKey           string            `json:"api_key"`
	ClientCertPath   string            `json:"client_cert_path"`
	ClientKeyPath    string            `json:"client_key_path"`
	CACertPath       string            `json:"ca_cert_path"`
	TimeoutInSeconds int               `json:"timeout_in_seconds"`
	Insecure         bool              `json:"insecure"`
	NoTLS            bool              `json:"no_tls"`
	Headers          map[string]string `json:"headers"`
}

func (cfg *Config) ToConnectionOptions(dop DialOptionsProvider) ([]ConnectionOption, error) {
	if cfg.APIKey != "" && cfg.Token != "" {
		return nil, errors.Wrap(ErrInvalidConfig, "api_key and token are mutually exclusive")
	}

	if cfg.Insecure && cfg.NoTLS {
		return nil, errors.Wrap(ErrInvalidConfig, "insecure and no_tls are mutually exclusive")
	}

	options := []ConnectionOption{
		WithInsecure(cfg.Insecure),
		WithNoTLS(cfg.NoTLS),
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

	if cfg.TenantID != "" {
		options = append(options, WithTenantID(cfg.TenantID))
	}

	for key, value := range cfg.Headers {
		options = append(options, WithHeader(key, value))
	}

	opts, err := dop(cfg)
	if err != nil {
		return nil, err
	}

	options = append(options, WithDialOptions(opts...))

	return options, nil
}
