package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/influxdata/telegraf/internal/config"
)

func NewClientConfig(config config.ClientTLSConfig) (*tls.Config, error) {
	if len(config.TLSCACerts) == 0 && config.SSLCA != "" {
		config.TLSCACerts = append(config.TLSCACerts, config.SSLCA)
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
	if len(config.TLSCACerts) == 0 && config.TLSKey == "" && config.TLSCert == "" && !config.InsecureSkipVerify {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.InsecureSkipVerify,
	}

	if len(config.TLSCACerts) > 0 {
		pool, err := makeCertPool(config.TLSCACerts)
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
func NewServerConfig(config config.ServerTLSConfig) (*tls.Config, error) {
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

func makeCertPool(certs []string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, cert := range certs {
		pem, err := ioutil.ReadFile(cert)
		if err != nil {
			return nil, fmt.Errorf(
				"could not read certificate %q: %v", cert, err)
		}
		ok := pool.AppendCertsFromPEM(pem)
		if !ok {
			return nil, fmt.Errorf(
				"could not parse any PEM certificates %q: %v", cert, err)
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
