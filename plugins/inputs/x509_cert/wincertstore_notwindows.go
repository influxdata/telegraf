//go:build !windows

package x509_cert

import (
	"crypto/x509"
	"errors"

	"github.com/influxdata/telegraf"
)

type wincertstore struct {
	source string
}

func newWincertStore(string, telegraf.Logger) (*wincertstore, error) {
	return nil, errors.New("not supported on this platform")
}

func (*wincertstore) read() ([]*x509.Certificate, error) {
	return nil, nil
}
