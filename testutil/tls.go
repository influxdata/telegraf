package testutil

import (
	"path"

	"github.com/influxdata/telegraf/plugins/common/tls"
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
		TLSCipherSuites:   []string{p.CipherSuite()},
		TLSMinVersion:     p.TLSMinVersion(),
		TLSMaxVersion:     p.TLSMaxVersion(),
	}
}

func (p *pki) ReadCACert() string {
	return tls.ReadCertificate(p.CACertPath())
}

func (p *pki) CACertPath() string {
	return path.Join(p.path, "cacert.pem")
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
	return tls.ReadCertificate(p.ClientCertPath())
}

func (p *pki) ClientCertPath() string {
	return path.Join(p.path, "clientcert.pem")
}

func (p *pki) ReadClientKey() string {
	return tls.ReadKey(p.ClientKeyPath())
}

func (p *pki) ClientKeyPath() string {
	return path.Join(p.path, "clientkey.pem")
}

func (p *pki) ReadClientCertAndKey() string {
	return tls.ReadKey(p.ClientCertAndKeyPath())
}

func (p *pki) ClientCertAndKeyPath() string {
	return path.Join(p.path, "client.pem")
}

func (p *pki) ReadClientEncKey() string {
	return tls.ReadKey(p.ClientEncKeyPath())
}

func (p *pki) ClientEncKeyPath() string {
	return path.Join(p.path, "clientenckey.pem")
}

func (p *pki) ReadClientCertAndEncKey() string {
	return tls.ReadKey(p.ClientCertAndEncKeyPath())
}

func (p *pki) ClientCertAndEncKeyPath() string {
	return path.Join(p.path, "clientenc.pem")
}

func (p *pki) ReadServerCert() string {
	return tls.ReadCertificate(p.ServerCertPath())
}

func (p *pki) ServerCertPath() string {
	return path.Join(p.path, "servercert.pem")
}

func (p *pki) ReadServerKey() string {
	return tls.ReadKey(p.ServerKeyPath())
}

func (p *pki) ServerKeyPath() string {
	return path.Join(p.path, "serverkey.pem")
}

func (p *pki) ReadServerCertAndKey() string {
	return tls.ReadKey(p.ServerCertAndKeyPath())
}

func (p *pki) ServerCertAndKeyPath() string {
	return path.Join(p.path, "server.pem")
}

func (p *pki) ReadServerEncKey() string {
	return tls.ReadKey(p.ServerEncKeyPath())
}

func (p *pki) ServerEncKeyPath() string {
	return path.Join(p.path, "serverenckey.pem")
}

func (p *pki) ReadServerCertAndEncKey() string {
	return tls.ReadKey(p.ServerCertAndEncKeyPath())
}

func (p *pki) ServerCertAndEncKeyPath() string {
	return path.Join(p.path, "serverenc.pem")
}
