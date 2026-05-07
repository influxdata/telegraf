package tls

import (
	"crypto/tls"
	"fmt"
)

// ClientConfig represents the standard client TLS config.
type ClientConfig struct {
	TLSCA               string   `toml:"tls_ca"`
	TLSCert             string   `toml:"tls_cert"`
	TLSKey              string   `toml:"tls_key"`
	TLSKeyPwd           string   `toml:"tls_key_pwd"`
	TLSMinVersion       string   `toml:"tls_min_version"`
	TLSCipherSuites     []string `toml:"tls_cipher_suites"`
	InsecureSkipVerify  bool     `toml:"insecure_skip_verify"`
	ServerName          string   `toml:"tls_server_name"`
	RenegotiationMethod string   `toml:"tls_renegotiation_method"`
	Enable              *bool    `toml:"tls_enable"`
}

// TLSConfig returns a tls.Config, may be nil without error if TLS is not
// configured.
func (c *ClientConfig) TLSConfig() (*tls.Config, error) {
	// Check if TLS config is forcefully disabled
	if c.Enable != nil && !*c.Enable {
		return nil, nil
	}

	// This check returns a nil (aka "disabled") or an empty config
	// (aka, "use the default") if no field is set that would have an effect on
	// a TLS connection. That is, any of:
	//     * client certificate settings,
	//     * peer certificate authorities,
	//     * disabled security,
	//     * an SNI server name, or
	//     * empty/never renegotiation method
	empty := c.TLSCA == "" && c.TLSKey == "" && c.TLSCert == ""
	empty = empty && !c.InsecureSkipVerify && c.ServerName == ""
	empty = empty && (c.RenegotiationMethod == "" || c.RenegotiationMethod == "never")

	if empty {
		// Check if TLS config is forcefully enabled and supposed to
		// use the system defaults.
		if c.Enable != nil && *c.Enable {
			return &tls.Config{}, nil
		}

		return nil, nil
	}

	var renegotiationMethod tls.RenegotiationSupport
	switch c.RenegotiationMethod {
	case "", "never":
		renegotiationMethod = tls.RenegotiateNever
	case "once":
		renegotiationMethod = tls.RenegotiateOnceAsClient
	case "freely":
		renegotiationMethod = tls.RenegotiateFreelyAsClient
	default:
		return nil, fmt.Errorf("unrecognized renegotiation method %q, choose from: 'never', 'once', 'freely'", c.RenegotiationMethod)
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.InsecureSkipVerify,
		Renegotiation:      renegotiationMethod,
	}

	if c.TLSCA != "" {
		pool, err := makeCertPool([]string{c.TLSCA})
		if err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = pool
	}

	if c.TLSCert != "" && c.TLSKey != "" {
		err := loadCertificate(tlsConfig, c.TLSCert, c.TLSKey, c.TLSKeyPwd)
		if err != nil {
			return nil, err
		}
	}

	// Explicitly and consistently set the minimal accepted version using the
	// defined default. We use this setting for both clients and servers
	// instead of relying on Golang's default that is different for clients
	// and servers and might change over time.
	tlsConfig.MinVersion = TLSMinVersionDefault
	if c.TLSMinVersion != "" {
		version, err := ParseTLSVersion(c.TLSMinVersion)
		if err != nil {
			return nil, fmt.Errorf("could not parse tls min version %q: %w", c.TLSMinVersion, err)
		}
		tlsConfig.MinVersion = version
	}

	if c.ServerName != "" {
		tlsConfig.ServerName = c.ServerName
	}

	if len(c.TLSCipherSuites) != 0 {
		cipherSuites, err := ParseCiphers(c.TLSCipherSuites)
		if err != nil {
			return nil, fmt.Errorf("could not parse client cipher suites: %w", err)
		}
		tlsConfig.CipherSuites = cipherSuites
	}

	return tlsConfig, nil
}
