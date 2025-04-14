package aserto

import (
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

var ErrInvalidConfig = errors.New("invalid configuration")

// gRPC Client Configuration.
type Config struct {
	// Address of the service to connect to.
	//
	// Address is typically in the form "hostname:port" but may also be a Unix socket or DNS URI.
	// See https://github.com/grpc/grpc/blob/master/doc/naming.md#name-syntax for more details.
	Address string `json:"address" yaml:"address"`

	// A JWT to be used for authentication with the service.
	//
	// Token and APIKey are mutually exclusive.
	Token string `json:"token" yaml:"token"`

	// An API key to be used for authentication with the service.
	APIKey string `json:"api_key" yaml:"api_key"`

	// An Aserto tenant ID.
	TenantID string `json:"tenant_id" yaml:"tenant_id"`

	// An Aserto account ID.
	AccountID string `json:"account_id" yaml:"account_id"`

	// In mTLS connections, ClientCertPath is the path of the client's
	// certificate file.
	ClientCertPath string `json:"client_cert_path" yaml:"client_cert_path"`

	// In mTLS connections, ClientKeyPath is the path of the client's
	// private key file.
	ClientKeyPath string `json:"client_key_path" yaml:"client_key_path"`

	// In TLS connections, CACertPath is the path of a CA certificate to
	// validate the server's certificate against.
	CACertPath string `json:"ca_cert_path" yaml:"ca_cert_path"`

	// In TLS connections, skip verification of the server certificate.
	Insecure bool `json:"insecure" yaml:"insecure"`

	// Disable TLS and use a plaintext connection.
	NoTLS bool `json:"no_tls" yaml:"no_tls"`

	// NoProxy bypasses any configured HTTP proxy.
	NoProxy bool `json:"no_proxy" yaml:"no_proxy"`

	// Additional headers to include in requests to the service.
	Headers map[string]string `json:"headers" yaml:"headers"`

	// Deprecated: no longer used. Timeouts are controlled on a per-call basis
	// by the provided context.
	TimeoutInSeconds int `json:"timeout_in_seconds" yaml:"timeout_in_seconds"`
}

// Connects to the service specified in Config, possibly with additional
// connection options.
func (cfg *Config) Connect(opts ...ConnectionOption) (*grpc.ClientConn, error) {
	if cfg.APIKey != "" {
		opts = append(opts, WithAPIKeyAuth(cfg.APIKey))
	}

	if cfg.Token != "" {
		opts = append(opts, WithTokenAuth(cfg.Token))
	}

	connOpts := &ConnectionOptions{Config: *cfg}
	if err := connOpts.Apply(opts...); err != nil {
		return nil, err
	}

	return Connect(connOpts)
}

// Converts the Config into a ConnectionOption slice that can be passed to NewConnection().
func (cfg *Config) ToConnectionOptions() ([]ConnectionOption, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
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

	if cfg.ClientCertPath != "" {
		options = append(options, WithClientCert(cfg.ClientCertPath, cfg.ClientKeyPath))
	}

	for key, value := range cfg.Headers {
		options = append(options, WithHeader(key, value))
	}

	return options, nil
}

func (cfg *Config) validate() error {
	if cfg.APIKey != "" && cfg.Token != "" {
		return errors.Wrap(ErrInvalidConfig, "api_key and token are mutually exclusive")
	}

	if cfg.Insecure && cfg.NoTLS {
		return errors.Wrap(ErrInvalidConfig, "insecure and no_tls are mutually exclusive")
	}

	if cfg.NoTLS && (cfg.ClientCertPath != "" || cfg.ClientKeyPath != "") {
		return errors.Wrap(ErrInvalidConfig, "mtls (client_cert_path and client_cert_key) and no_tls are mutually exclusive")
	}

	if !cfg.NoTLS && ((cfg.ClientCertPath == "") != (cfg.ClientKeyPath == "")) {
		return errors.Wrap(ErrInvalidConfig, "client_cert_path and client_key_path must be specified together")
	}

	return nil
}
