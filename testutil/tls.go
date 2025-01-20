package testutil

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/influxdata/telegraf/plugins/common/tls"
)

type PKIPaths struct {
	ServerPem  string
	ServerCert string
	ServerKey  string
	ClientCert string
}

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

func (*pki) CipherSuite() string {
	return "TLS_RSA_WITH_3DES_EDE_CBC_SHA"
}

func (*pki) TLSMinVersion() string {
	return "TLS11"
}

func (*pki) TLSMaxVersion() string {
	return "TLS13"
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
	return path.Join(p.keyPath, "clientenckey.pem")
}

func (p *pki) ClientPKCS8KeyPath() string {
	return path.Join(p.keyPath, "clientkey.pkcs8.pem")
}

func (p *pki) ClientEncPKCS8KeyPath() string {
	return path.Join(p.keyPath, "clientenckey.pkcs8.pem")
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

func (p *pki) AbsolutePaths() (*PKIPaths, error) {
	tlsPem, err := filepath.Abs(p.ServerCertAndKeyPath())
	if err != nil {
		return nil, err
	}
	tlsCert, err := filepath.Abs(p.ServerCertPath())
	if err != nil {
		return nil, err
	}
	tlsKey, err := filepath.Abs(p.ServerKeyPath())
	if err != nil {
		return nil, err
	}
	cert, err := filepath.Abs(p.ClientCertPath())
	if err != nil {
		return nil, err
	}

	return &PKIPaths{
		ServerPem:  tlsPem,
		ServerCert: tlsCert,
		ServerKey:  tlsKey,
		ClientCert: cert,
	}, nil
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
