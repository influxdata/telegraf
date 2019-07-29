package testutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/influxdata/telegraf/internal/tls"
)

type pki struct {
	path string
}

func NewPKI(path string) *pki {
	return &pki{path: path}
}

func (p *pki) TLSClientConfig() *tls.ClientConfig {
	return &tls.ClientConfig{
		TLSCA:   p.CACertPath(),
		TLSCert: p.ClientCertPath(),
		TLSKey:  p.ClientKeyPath(),
	}
}

func (p *pki) TLSServerConfig() *tls.ServerConfig {
	return &tls.ServerConfig{
		TLSAllowedCACerts: []string{p.CACertPath()},
		TLSCert:           p.ServerCertPath(),
		TLSKey:            p.ServerKeyPath(),
	}
}

func (p *pki) ReadCACert() string {
	return readCertificate(p.CACertPath())
}

func (p *pki) CACertPath() string {
	return path.Join(p.path, "cacert.pem")
}

func (p *pki) ReadClientCert() string {
	return readCertificate(p.ClientCertPath())
}

func (p *pki) ClientCertPath() string {
	return path.Join(p.path, "clientcert.pem")
}

func (p *pki) ReadClientKey() string {
	return readCertificate(p.ClientKeyPath())
}

func (p *pki) ClientKeyPath() string {
	return path.Join(p.path, "clientkey.pem")
}

func (p *pki) ReadServerCert() string {
	return readCertificate(p.ServerCertPath())
}

func (p *pki) ServerCertPath() string {
	return path.Join(p.path, "servercert.pem")
}

func (p *pki) ReadServerKey() string {
	return readCertificate(p.ServerKeyPath())
}

func (p *pki) ServerKeyPath() string {
	return path.Join(p.path, "serverkey.pem")
}

func readCertificate(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("opening %q: %v", filename, err))
	}
	octets, err := ioutil.ReadAll(file)
	if err != nil {
		panic(fmt.Sprintf("reading %q: %v", filename, err))
	}
	return string(octets)
}
