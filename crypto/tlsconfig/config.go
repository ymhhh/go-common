package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the config for client TLS.
type Config struct {
	CertPath           string `yaml:"cert_path" json:"cert_path"`
	KeyPath            string `yaml:"key_path" json:"key_path"`
	CAPath             string `yaml:"ca_path" json:"ca_path"`
	ServerName         string `yaml:"server_name" json:"server_name"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
}

var _ yaml.Unmarshaler = (*Config)(nil)
var _ yaml.Marshaler = (*Config)(nil)

// MarshalYAML implements yaml.Marshaler.
func (c Config) MarshalYAML() (any, error) {
	type plain Config
	return plain(c), nil
}

// UnmarshalYAML implements yaml.Unmarshaler.
func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		*c = Config{}
		return nil
	}
	type plain Config
	var tmp plain
	if err := value.Decode(&tmp); err != nil {
		return err
	}
	*c = Config(tmp)
	return nil
}

// GetTLSConfig builds a *tls.Config for outgoing TLS connections.
//
// Rules:
// - If both CertPath and KeyPath are set, a client certificate is loaded.
// - If only one of CertPath/KeyPath is set, it returns an error.
// - If CAPath is set, its PEM bundle is appended to RootCAs (starting from the system pool when available).
func (c Config) GetTLSConfig() (*tls.Config, error) {
	tcfg := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		ServerName:         c.ServerName,
		InsecureSkipVerify: c.InsecureSkipVerify,
	}

	switch {
	case c.CertPath != "" && c.KeyPath != "":
		cert, err := tls.LoadX509KeyPair(c.CertPath, c.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("tlsconfig: load client cert/key: %w", err)
		}
		tcfg.Certificates = []tls.Certificate{cert}
	case c.CertPath != "" || c.KeyPath != "":
		return nil, fmt.Errorf("tlsconfig: cert_path and key_path must both be set or both be empty")
	}

	if c.CAPath != "" {
		pem, err := os.ReadFile(c.CAPath)
		if err != nil {
			return nil, fmt.Errorf("tlsconfig: read ca_path: %w", err)
		}

		pool, err := x509.SystemCertPool()
		if err != nil {
			pool = x509.NewCertPool()
		}
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("tlsconfig: failed to parse certificates from ca_path")
		}
		tcfg.RootCAs = pool
	}

	return tcfg, nil
}
