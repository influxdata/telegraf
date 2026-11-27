package tls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
)

// ServerConfig represents the standard server TLS config.
type ServerConfig struct {
	TLSCert            string   `toml:"tls_cert"`
	TLSKey             string   `toml:"tls_key"`
	TLSKeyPwd          string   `toml:"tls_key_pwd"`
	TLSAllowedCACerts  []string `toml:"tls_allowed_cacerts"`
	TLSCipherSuites    []string `toml:"tls_cipher_suites"`
	TLSMinVersion      string   `toml:"tls_min_version"`
	TLSMaxVersion      string   `toml:"tls_max_version"`
	TLSAllowedDNSNames []string `toml:"tls_allowed_dns_names" deprecated:"1.41.0;1.50.0;use 'tls_allowed_clients' with 'dns' prefix instead"`
	TLSAllowedClients  []string `toml:"tls_allowed_clients"`

	clientFilter *certificateFilter
}

// TLSConfig returns a tls.Config, may be nil without error if TLS is not
// configured.
func (c *ServerConfig) TLSConfig() (*tls.Config, error) {
	if c.TLSCert == "" && c.TLSKey == "" && len(c.TLSAllowedCACerts) == 0 {
		return nil, nil
	}

	tlsConfig := &tls.Config{}

	if len(c.TLSAllowedCACerts) != 0 {
		pool, err := makeCertPool(c.TLSAllowedCACerts)
		if err != nil {
			return nil, err
		}
		tlsConfig.ClientCAs = pool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	if c.TLSCert != "" && c.TLSKey != "" {
		err := loadCertificate(tlsConfig, c.TLSCert, c.TLSKey, c.TLSKeyPwd)
		if err != nil {
			return nil, err
		}
	}

	if len(c.TLSCipherSuites) != 0 {
		cipherSuites, err := ParseCiphers(c.TLSCipherSuites)
		if err != nil {
			return nil, fmt.Errorf("could not parse server cipher suites: %w", err)
		}
		tlsConfig.CipherSuites = cipherSuites
	}

	if c.TLSMaxVersion != "" {
		version, err := ParseTLSVersion(c.TLSMaxVersion)
		if err != nil {
			return nil, fmt.Errorf(
				"could not parse tls max version %q: %w", c.TLSMaxVersion, err)
		}
		tlsConfig.MaxVersion = version
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

	if tlsConfig.MinVersion != 0 && tlsConfig.MaxVersion != 0 && tlsConfig.MinVersion > tlsConfig.MaxVersion {
		return nil, fmt.Errorf("tls min version %d can't be greater than tls max version %d", tlsConfig.MinVersion, tlsConfig.MaxVersion)
	}

	// Prepare the client filtering if enabled
	if len(c.TLSAllowedCACerts) > 0 && (len(c.TLSAllowedDNSNames) > 0 || len(c.TLSAllowedClients) > 0) {
		// For backward compatibility
		allowed := c.TLSAllowedClients
		if len(c.TLSAllowedDNSNames) > 0 {
			allowed = make([]string, 0, len(c.TLSAllowedDNSNames)+len(c.TLSAllowedClients))
			for _, n := range c.TLSAllowedDNSNames {
				allowed = append(allowed, "dns:"+n)
			}
			allowed = append(allowed, c.TLSAllowedClients...)
		}

		f, err := compileCertificateFilter(allowed)
		if err != nil {
			return nil, fmt.Errorf("compiling client certificate filter failed: %w", err)
		}
		c.clientFilter = f
		tlsConfig.VerifyPeerCertificate = c.verifyPeerCertificate
	}

	return tlsConfig, nil
}

func (c *ServerConfig) verifyPeerCertificate(rawCerts [][]byte, _ [][]*x509.Certificate) error {
	// The certificate chain is client + intermediate + root.
	// Let's review the client certificate.
	cert, err := x509.ParseCertificate(rawCerts[0])
	if err != nil {
		return fmt.Errorf("could not parse peer certificate: %w", err)
	}

	// Try to match the certificate against the "allowed" filter
	if c.clientFilter.match(cert) {
		return nil
	}

	return errors.New("peer certificate not in allowed list")
}
