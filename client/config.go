package client

// gRPC Client Configuration.
type Config struct {
	Address          string            `json:"address"`
	Token            string            `json:"token"`
	APIKey           string            `json:"api_key"`
	ClientCertPath   string            `json:"client_cert_path"`
	ClientKeyPath    string            `json:"client_key_path"`
	CACertPath       string            `json:"ca_cert_path"`
	TimeoutInSeconds int               `json:"timeout_in_seconds"`
	Insecure         bool              `json:"insecure"`
	Headers          map[string]string `json:"headers"`
}
