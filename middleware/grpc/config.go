package grpc

import (
	"bytes"
	"encoding/json"

	"github.com/aserto-dev/aserto-grpc/grpcclient"
	"github.com/aserto-dev/aserto-grpc/grpcutil/middlewares/request"
	"github.com/aserto-dev/go-aserto/client"
	"github.com/aserto-dev/go-aserto/middleware"

	"github.com/pkg/errors"
)

type AuthorizationConfig struct {
	// Can be "none", "self" or "remote".
	Mode       string            `json:"mode"`
	ModeParsed AuthorizationType `json:"-"`
	TenantID   string            `json:"tenant_id"`
	Policy     middleware.Policy `json:"policy"`
	Authorizer grpcclient.Config `json:"authorizer"`
}

func (cfg *AuthorizationConfig) ToClientOptions(dop grpcclient.DialOptionsProvider) ([]client.ConnectionOption, error) {
	reqMiddleware := request.NewRequestIDMiddleware()
	options := []client.ConnectionOption{
		client.WithChainUnaryInterceptor(reqMiddleware.UnaryClient()),
		client.WithChainStreamInterceptor(reqMiddleware.StreamClient()),
		client.WithInsecure(cfg.Authorizer.Insecure),
	}

	if cfg.Authorizer.APIKey != "" && cfg.Authorizer.Token != "" {
		return nil, errors.New("both api_key and token are set")
	}

	if cfg.Authorizer.Token != "" {
		options = append(options, client.WithTokenAuth(cfg.Authorizer.Token))
	}

	if cfg.Authorizer.APIKey != "" {
		options = append(options, client.WithAPIKeyAuth(cfg.Authorizer.APIKey))
	}

	if cfg.Authorizer.Address != "" {
		options = append(options, client.WithAddr(cfg.Authorizer.Address))
	}

	if cfg.Authorizer.CACertPath != "" {
		options = append(options, client.WithCACertPath(cfg.Authorizer.CACertPath))
	}

	opts, err := dop(&cfg.Authorizer)
	if err != nil {
		return nil, err
	}

	options = append(options, client.WithDialOptions(opts...))

	return options, nil
}

// AuthorizationType represents the type of authorization to use.
type AuthorizationType int

const (
	// Authorization type was not set.
	Unknown AuthorizationType = iota
	// Don't use any authorization.
	None
	// Use a loaded policy from an in-memory runtime. TODO: implement.
	Self
	// Use a loaded policy from a remote server. TODO: implement.
	Remote
)

func (t AuthorizationType) String() string {
	return authorizationTypeToString[t]
}

var authorizationTypeToString = map[AuthorizationType]string{ //nolint: gochecknoglobals
	None:   "none",
	Self:   "self",
	Remote: "remote",
}

var authorizationTypeToID = map[string]AuthorizationType{ //nolint: gochecknoglobals
	"none":   None,
	"self":   Self,
	"remote": Remote,
}

// MarshalJSON marshals the enum as a quoted json string.
func (t AuthorizationType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(authorizationTypeToString[t])
	buffer.WriteString(`"`)

	return buffer.Bytes(), nil
}

// UnmarshalJSON un-marshalls a quoted json string to the enum value.
func (t *AuthorizationType) UnmarshalJSON(b []byte) error {
	var j string

	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}

	var ok bool

	*t, ok = authorizationTypeToID[j]
	if !ok {
		return errors.Errorf("'%s' is not a valid authorization type", j)
	}

	return nil
}
