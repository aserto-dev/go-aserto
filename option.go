package aserto

import (
	"net/url"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/aserto-dev/go-aserto/internal/client"
)

var ErrInvalidOptions = errors.New("invalid connection options")

// ConnectionOption functions are used to configure ConnectionOptions instances.
type ConnectionOption func(*ConnectionOptions) error

// WithInsecure disables TLS verification.
func WithInsecure(insecure bool) ConnectionOption {
	return func(options *ConnectionOptions) error {
		options.Insecure = insecure

		return nil
	}
}

// WithAddr overrides the default authorizer server address.
//
// Note: WithAddr and WithURL are mutually exclusive.
func WithAddr(addr string) ConnectionOption {
	return func(options *ConnectionOptions) error {
		if options.Address != "" {
			return errors.Wrap(ErrInvalidOptions, "address has already been set")
		}

		options.Address = addr

		return nil
	}
}

// WithURL overrides the default authorizer server URL.
// Unlike WithAddr, WithURL lets gRPC users to connect to communicate with a locally running authorizer
// over Unix sockets. See https://github.com/grpc/grpc/blob/master/doc/naming.md#grpc-name-resolution for
// more details about gRPC name resolution.
//
// Note: WithURL and WithAddr are mutually exclusive.
func WithURL(svcURL *url.URL) ConnectionOption {
	return func(options *ConnectionOptions) error {
		if options.Address != "" {
			return errors.Wrap(ErrInvalidOptions, "address has already been set")
		}

		options.Address = svcURL.String()

		return nil
	}
}

// WithCACertPath treats the specified certificate file as a trusted root CA.
//
// Include it when calling a service that uses a self-issued SSL certificate.
func WithCACertPath(path string) ConnectionOption {
	return func(options *ConnectionOptions) error {
		options.CACertPath = path

		return nil
	}
}

// WithClientCert configure the client certificate for mTLS connections.
func WithClientCert(certPath, keyPath string) ConnectionOption {
	return func(options *ConnectionOptions) error {
		if certPath == "" || keyPath == "" {
			return errors.Wrap(ErrInvalidOptions, "both client certificate and private key paths must be specified")
		}

		options.ClientCertPath = certPath
		options.ClientKeyPath = keyPath

		return nil
	}
}

// WithTokenAuth uses an OAuth2.0 token to authenticate with the authorizer service.
func WithTokenAuth(token string) ConnectionOption {
	return func(options *ConnectionOptions) error {
		if options.Creds != nil {
			return errors.Wrap(ErrInvalidOptions, "only one set of credentials allowed")
		}

		options.Creds = client.NewTokenAuth(token)

		return nil
	}
}

// WithAPIKeyAuth uses an Aserto API key to authenticate with the authorizer service.
func WithAPIKeyAuth(key string) ConnectionOption {
	return func(options *ConnectionOptions) error {
		if options.Creds != nil {
			return errors.Wrap(ErrInvalidOptions, "only one set of credentials allowed")
		}

		options.Creds = client.NewAPIKeyAuth(key)

		return nil
	}
}

// WithTenantID sets the Aserto tenant ID.
func WithTenantID(tenantID string) ConnectionOption {
	return func(options *ConnectionOptions) error {
		options.TenantID = tenantID

		return nil
	}
}

// WithAccountID sets the Aserto account ID.
func WithAccountID(accountID string) ConnectionOption {
	return func(options *ConnectionOptions) error {
		options.AccountID = accountID
		return nil
	}
}

// WithChainUnaryInterceptor adds a unary interceptor to grpc dial options.
func WithChainUnaryInterceptor(mw ...grpc.UnaryClientInterceptor) ConnectionOption {
	return func(options *ConnectionOptions) error {
		options.UnaryClientInterceptors = append(options.UnaryClientInterceptors, mw...)
		return nil
	}
}

// WithChainStreamInterceptor adds a stream interceptor to grpc dial options.
func WithChainStreamInterceptor(mw ...grpc.StreamClientInterceptor) ConnectionOption {
	return func(options *ConnectionOptions) error {
		options.StreamClientInterceptors = append(options.StreamClientInterceptors, mw...)
		return nil
	}
}

// WithDialOptions add custom dial options to the grpc connection.
func WithDialOptions(opts ...grpc.DialOption) ConnectionOption {
	return func(options *ConnectionOptions) error {
		options.DialOptions = append(options.DialOptions, opts...)
		return nil
	}
}

// WithHeader adds an header to the client config instance.
func WithHeader(key, value string) ConnectionOption {
	return func(options *ConnectionOptions) error {
		if options.Headers == nil {
			options.Headers = map[string]string{}
		}

		options.Headers[key] = value

		return nil
	}
}

// WithNoTLS disables transport security. The connection is established in plaintext.
func WithNoTLS(noTLS bool) ConnectionOption {
	return func(options *ConnectionOptions) error {
		options.NoTLS = noTLS
		return nil
	}
}
