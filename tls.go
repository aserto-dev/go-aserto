package aserto

import (
	"crypto/tls"
	"os"

	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/aserto-dev/go-aserto/internal/tlsconf"
)

type NoTLSVerify bool

const (
	VerifyTLS     = NoTLSVerify(false)
	SkipVerifyTLS = NoTLSVerify(true)
)

// TLSConfig contains paths to an X509 certificate's key-pair and CA files.
// It can be used to create client or server tls.Config or grpc TransportCredentials.
type TLSConfig struct {
	Cert string `json:"tls_cert_path"`
	Key  string `json:"tls_key_path"`
	CA   string `json:"tls_ca_cert_path"`
}

func (c *TLSConfig) IsTLS() bool {
	return c != nil && c.Cert != "" && c.Key != "" && c.CA != ""
}

func (c *TLSConfig) NoTLS() bool {
	return !c.IsTLS()
}

// ServerConfig returns TLS configuration for a server.
func (c *TLSConfig) ServerConfig() (*tls.Config, error) {
	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	if c.NoTLS() {
		return cfg, nil
	}

	certificate, err := tls.LoadX509KeyPair(c.Cert, c.Key)
	if err != nil {
		return cfg, errors.Wrapf(err, "failed to load gateway certs")
	}

	cfg.Certificates = []tls.Certificate{certificate}

	return cfg, nil
}

// ClientConfig returns TLS configuration for a client.
func (c *TLSConfig) ClientConfig(skipVerify NoTLSVerify) (*tls.Config, error) {
	conf, err := c.ServerConfig()
	if err != nil {
		return &tls.Config{MinVersion: tls.VersionTLS12}, err
	}

	if skipVerify == SkipVerifyTLS {
		conf.InsecureSkipVerify = true
		return conf, nil
	}

	certPool, err := tlsconf.CertPool(c.CA)
	if err != nil {
		return conf, errors.Wrap(err, "failed to create certificate pool")
	}

	caCertBytes, err := os.ReadFile(c.CA)
	if err != nil {
		return conf, errors.Wrapf(err, "failed to read ca cert: %s", c.CA)
	}

	if !certPool.AppendCertsFromPEM(caCertBytes) {
		return conf, errors.Wrap(err, "failed to append client ca cert: %s")
	}

	conf.RootCAs = certPool

	return conf, nil
}

// ServerCredentials returns transport credentials for a GRPC server.
func (c *TLSConfig) ServerCredentials() (credentials.TransportCredentials, error) {
	if c.NoTLS() {
		return insecure.NewCredentials(), nil
	}

	tlsConfig, err := c.ServerConfig()
	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(tlsConfig), nil
}

// ClientCredentials returns transport credentials for a GRPC client.
func (c *TLSConfig) ClientCredentials(skipVerify NoTLSVerify) (credentials.TransportCredentials, error) {
	if c.NoTLS() {
		return insecure.NewCredentials(), nil
	}

	tlsConfig, err := c.ClientConfig(skipVerify)
	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(tlsConfig), nil
}
