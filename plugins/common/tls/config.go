package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/influxdata/telegraf/internal/choice"
)

// ClientConfig represents the standard client TLS config.
type ClientConfig struct {
	TLSCA              string `toml:"tls_ca"`
	TLSCert            string `toml:"tls_cert"`
	TLSKey             string `toml:"tls_key"`
	TLSKeyPwd          string `toml:"tls_key_pwd"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`
	ServerName         string `toml:"tls_server_name"`

	SSLCA   string `toml:"ssl_ca" deprecated:"1.7.0;use 'tls_ca' instead"`
	SSLCert string `toml:"ssl_cert" deprecated:"1.7.0;use 'tls_cert' instead"`
	SSLKey  string `toml:"ssl_key" deprecated:"1.7.0;use 'tls_key' instead"`
}

// ServerConfig represents the standard server TLS config.
type ServerConfig struct {
	TLSCert            string   `toml:"tls_cert"`
	TLSKey             string   `toml:"tls_key"`
	TLSKeyPwd          string   `toml:"tls_key_pwd"`
	TLSAllowedCACerts  []string `toml:"tls_allowed_cacerts"`
	TLSCipherSuites    []string `toml:"tls_cipher_suites"`
	TLSMinVersion      string   `toml:"tls_min_version"`
	TLSMaxVersion      string   `toml:"tls_max_version"`
	TLSAllowedDNSNames []string `toml:"tls_allowed_dns_names"`
}

// TLSConfig returns a tls.Config, may be nil without error if TLS is not
// configured.
func (c *ClientConfig) TLSConfig() (*tls.Config, error) {
	// Support deprecated variable names
	if c.TLSCA == "" && c.SSLCA != "" {
		c.TLSCA = c.SSLCA
	}
	if c.TLSCert == "" && c.SSLCert != "" {
		c.TLSCert = c.SSLCert
	}
	if c.TLSKey == "" && c.SSLKey != "" {
		c.TLSKey = c.SSLKey
	}

	// This check returns a nil (aka, "use the default")
	// tls.Config if no field is set that would have an effect on
	// a TLS connection. That is, any of:
	//     * client certificate settings,
	//     * peer certificate authorities,
	//     * disabled security, or
	//     * an SNI server name.
	if c.TLSCA == "" && c.TLSKey == "" && c.TLSCert == "" && !c.InsecureSkipVerify && c.ServerName == "" {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.InsecureSkipVerify,
		Renegotiation:      tls.RenegotiateNever,
	}

	if c.TLSCA != "" {
		pool, err := makeCertPool([]string{c.TLSCA})
		if err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = pool
	}

	if c.TLSCert != "" && c.TLSKey != "" {
		err := loadCertificate(tlsConfig, c.TLSCert, c.TLSKey)
		if err != nil {
			return nil, err
		}
	}

	if c.ServerName != "" {
		tlsConfig.ServerName = c.ServerName
	}

	return tlsConfig, nil
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
		err := loadCertificate(tlsConfig, c.TLSCert, c.TLSKey)
		if err != nil {
			return nil, err
		}
	}

	if len(c.TLSCipherSuites) != 0 {
		cipherSuites, err := ParseCiphers(c.TLSCipherSuites)
		if err != nil {
			return nil, fmt.Errorf(
				"could not parse server cipher suites %s: %v", strings.Join(c.TLSCipherSuites, ","), err)
		}
		tlsConfig.CipherSuites = cipherSuites
	}

	if c.TLSMaxVersion != "" {
		version, err := ParseTLSVersion(c.TLSMaxVersion)
		if err != nil {
			return nil, fmt.Errorf(
				"could not parse tls max version %q: %v", c.TLSMaxVersion, err)
		}
		tlsConfig.MaxVersion = version
	}

	if c.TLSMinVersion != "" {
		version, err := ParseTLSVersion(c.TLSMinVersion)
		if err != nil {
			return nil, fmt.Errorf(
				"could not parse tls min version %q: %v", c.TLSMinVersion, err)
		}
		tlsConfig.MinVersion = version
	}

	if tlsConfig.MinVersion != 0 && tlsConfig.MaxVersion != 0 && tlsConfig.MinVersion > tlsConfig.MaxVersion {
		return nil, fmt.Errorf(
			"tls min version %q can't be greater than tls max version %q", tlsConfig.MinVersion, tlsConfig.MaxVersion)
	}

	// Since clientAuth is tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	// there must be certs to validate.
	if len(c.TLSAllowedCACerts) > 0 && len(c.TLSAllowedDNSNames) > 0 {
		tlsConfig.VerifyPeerCertificate = c.verifyPeerCertificate
	}

	return tlsConfig, nil
}

func makeCertPool(certFiles []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, certFile := range certFiles {
		pem, err := os.ReadFile(certFile)
		if err != nil {
			return nil, fmt.Errorf(
				"could not read certificate %q: %v", certFile, err)
		}
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf(
				"could not parse any PEM certificates %q: %v", certFile, err)
		}
	}
	return pool, nil
}

func loadCertificate(config *tls.Config, certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return fmt.Errorf(
			"could not load keypair %s:%s: %v", certFile, keyFile, err)
	}

	config.Certificates = []tls.Certificate{cert}
	config.BuildNameToCertificate()
	return nil
}

func (c *ServerConfig) verifyPeerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	// The certificate chain is client + intermediate + root.
	// Let's review the client certificate.
	cert, err := x509.ParseCertificate(rawCerts[0])
	if err != nil {
		return fmt.Errorf("could not validate peer certificate: %v", err)
	}

	for _, name := range cert.DNSNames {
		if choice.Contains(name, c.TLSAllowedDNSNames) {
			return nil
		}
	}

	return fmt.Errorf("peer certificate not in allowed DNS Name list: %v", cert.DNSNames)
}
