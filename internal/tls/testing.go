package tls

import (
	"crypto/tls"
	"path"
)

func NewTestingTLSClientConfig(pki string) (*tls.Config, error) {
	clientConfig := ClientConfig{
		TLSCA:   path.Join(pki, "cacert.pem"),
		TLSCert: path.Join(pki, "clientcert.pem"),
		TLSKey:  path.Join(pki, "clientkey.pem"),
	}
	return NewClientTLSConfig(clientConfig)
}

func NewTestingTLSServerConfig(pki string) (*tls.Config, error) {
	serverConfig := ServerConfig{
		TLSAllowedCACerts: []string{path.Join(pki, "cacert.pem")},
		TLSCert:           path.Join(pki, "servercert.pem"),
		TLSKey:            path.Join(pki, "serverkey.pem"),
	}
	return NewServerTLSConfig(serverConfig)
}
