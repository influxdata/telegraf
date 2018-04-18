package config

// ClientTLSConfig represents the standard client TLS config.
type ClientTLSConfig struct {
	TLSCACerts         []string `toml:"tls_cacerts"`
	TLSCert            string   `toml:"tls_cert"`
	TLSKey             string   `toml:"tls_key"`
	InsecureSkipVerify bool     `toml:"insecure_skip_verify"`

	// Deprecated in 1.7; use TLS variables above
	SSLCA   string `toml:"ssl_ca"`
	SSLCert string `toml:"ssl_cert"`
	SSLKey  string `toml:"ssl_ca"`
}

// ServerTLSConfig represents the standard server TLS config.
type ServerTLSConfig struct {
	TLSCert           string   `toml:"tls_cert"`
	TLSKey            string   `toml:"tls_key"`
	TLSAllowedCACerts []string `toml:"tls_allowed_cacerts"`
}
