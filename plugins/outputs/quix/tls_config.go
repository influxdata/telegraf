// tls_config.go
package quix

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
)

// createTLSConfig sets up TLS configuration using the CA certificate
func (q *Quix) createTLSConfig(caCert []byte) (*tls.Config, error) {
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	q.Log.Debugf("TLS configuration created with CA certificate.")
	return &tls.Config{
		RootCAs:            certPool,
		InsecureSkipVerify: false,
	}, nil
}
