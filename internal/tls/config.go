package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// ClientConfig represents the standard client TLS config.
type ClientConfig struct {
	TLSCA              string `toml:"tls_ca"`
	TLSCert            string `toml:"tls_cert"`
	TLSKey             string `toml:"tls_key"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify"`

	// Deprecated in 1.7; use TLS variables above
	SSLCA   string `toml:"ssl_ca"`
	SSLCert string `toml:"ssl_cert"`
	SSLKey  string `toml:"ssl_ca"`
}

// ServerConfig represents the standard server TLS config.
type ServerConfig struct {
	TLSCert           string   `toml:"tls_cert"`
	TLSKey            string   `toml:"tls_key"`
	TLSAllowedCACerts []string `toml:"tls_allowed_cacerts"`
}

func NewClientTLSConfig(config ClientConfig) (*tls.Config, error) {
	if config.TLSCA == "" && config.SSLCA != "" {
		config.TLSCA = config.SSLCA
	}
	if config.TLSCert == "" && config.SSLCert != "" {
		config.TLSCert = config.SSLCert
	}
	if config.TLSKey == "" && config.SSLKey != "" {
		config.TLSKey = config.SSLKey
	}

	// TODO: return default tls.Config; plugins should not call if they don't
	// want TLS, this will require using another option to determine.  In the
	// case of an HTTP plugin, you could use `https`.  Other plugins may need
	// a dedicated option.
	if config.TLSCA == "" && config.TLSKey == "" && config.TLSCert == "" && !config.InsecureSkipVerify {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.InsecureSkipVerify,
	}

	if config.TLSCA != "" {
		pool, err := makeCertPool([]string{config.TLSCA})
		if err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = pool
	}

	if config.TLSCert != "" && config.TLSKey != "" {
		err := loadCertificate(tlsConfig, config.TLSCert, config.TLSKey)
		if err != nil {
			return nil, err
		}
	}

	return tlsConfig, nil
}

// GetServerTLSConfig gets a tls.Config object from the given certs, key, and one or more CA files
// for use with a server.
// The full path to each file must be provided.
// Returns a nil pointer if all files are blank.
func NewServerTLSConfig(config ServerConfig) (*tls.Config, error) {
	if config.TLSCert == "" && config.TLSKey == "" && len(config.TLSAllowedCACerts) == 0 {
		return nil, nil
	}

	tlsConfig := &tls.Config{}

	if len(config.TLSAllowedCACerts) != 0 {
		pool, err := makeCertPool(config.TLSAllowedCACerts)
		if err != nil {
			return nil, err
		}
		tlsConfig.ClientCAs = pool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	if config.TLSCert != "" && config.TLSKey != "" {
		err := loadCertificate(tlsConfig, config.TLSCert, config.TLSKey)
		if err != nil {
			return nil, err
		}
	}

	return tlsConfig, nil
}

func makeCertPool(certFiles []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, certFile := range certFiles {
		pem, err := ioutil.ReadFile(certFile)
		if err != nil {
			return nil, fmt.Errorf(
				"could not read certificate %q: %v", certFile, err)
		}
		ok := pool.AppendCertsFromPEM(pem)
		if !ok {
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
