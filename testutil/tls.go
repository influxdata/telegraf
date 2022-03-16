package testutil

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/influxdata/telegraf/plugins/common/tls"
)

type pki struct {
	keyPath string
}

func NewPKI(keyPath string) *pki {
	return &pki{keyPath: keyPath}
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
		TLSCipherSuites:   []string{p.CipherSuite()},
		TLSMinVersion:     p.TLSMinVersion(),
		TLSMaxVersion:     p.TLSMaxVersion(),
	}
}

func (p *pki) ReadCACert() string {
	return readCertificate(p.CACertPath())
}

func (p *pki) CACertPath() string {
	return path.Join(p.keyPath, "cacert.pem")
}

func (p *pki) CipherSuite() string {
	return "TLS_RSA_WITH_3DES_EDE_CBC_SHA"
}

func (p *pki) TLSMinVersion() string {
	return "TLS11"
}

func (p *pki) TLSMaxVersion() string {
	return "TLS12"
}

func (p *pki) ReadClientCert() string {
	return readCertificate(p.ClientCertPath())
}

func (p *pki) ClientCertPath() string {
	return path.Join(p.keyPath, "clientcert.pem")
}

func (p *pki) ReadClientKey() string {
	return readCertificate(p.ClientKeyPath())
}

func (p *pki) ClientKeyPath() string {
	return path.Join(p.keyPath, "clientkey.pem")
}

func (p *pki) ClientCertAndKeyPath() string {
	return path.Join(p.keyPath, "client.pem")
}

func (p *pki) ClientEncKeyPath() string {
	return path.Join(p.keyPath, "clientkeyenc.pem")
}

func (p *pki) ClientCertAndEncKeyPath() string {
	return path.Join(p.keyPath, "clientenc.pem")
}

func (p *pki) ReadServerCert() string {
	return readCertificate(p.ServerCertPath())
}

func (p *pki) ServerCertPath() string {
	return path.Join(p.keyPath, "servercert.pem")
}

func (p *pki) ReadServerKey() string {
	return readCertificate(p.ServerKeyPath())
}

func (p *pki) ServerKeyPath() string {
	return path.Join(p.keyPath, "serverkey.pem")
}

func (p *pki) ServerCertAndKeyPath() string {
	return path.Join(p.keyPath, "server.pem")
}

func readCertificate(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("opening %q: %v", filename, err))
	}
	octets, err := io.ReadAll(file)
	if err != nil {
		panic(fmt.Sprintf("reading %q: %v", filename, err))
	}
	return string(octets)
}
