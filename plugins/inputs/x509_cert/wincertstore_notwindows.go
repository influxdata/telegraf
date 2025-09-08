//go:build !windows

package x509_cert

import (
	"crypto/x509"
	"errors"
)

func (*X509Cert) processWinCertStore(string) ([]*x509.Certificate, error) {
	return nil, errors.New("not supported on this platform")
}
