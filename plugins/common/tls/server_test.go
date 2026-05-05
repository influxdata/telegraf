package tls_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/common/tls"
)

func TestServerConfig(t *testing.T) {
	tests := []struct {
		name   string
		server tls.ServerConfig
		expNil bool
		expErr bool
	}{
		{
			name:   "unset",
			server: tls.ServerConfig{},
			expNil: true,
		},
		{
			name: "success",
			server: tls.ServerConfig{
				TLSCert:            pki.ServerCertPath(),
				TLSKey:             pki.ServerKeyPath(),
				TLSAllowedCACerts:  []string{pki.CACertPath()},
				TLSCipherSuites:    []string{pki.CipherSuite()},
				TLSAllowedDNSNames: []string{"localhost", "127.0.0.1"},
				TLSMinVersion:      pki.TLSMinVersion(),
				TLSMaxVersion:      pki.TLSMaxVersion(),
			},
		},
		{
			name: "success with tls key password set",
			server: tls.ServerConfig{
				TLSCert:           pki.ServerCertPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSKeyPwd:         "",
				TLSAllowedCACerts: []string{pki.CACertPath()},
				TLSCipherSuites:   []string{pki.CipherSuite()},
				TLSMinVersion:     pki.TLSMinVersion(),
				TLSMaxVersion:     pki.TLSMaxVersion(),
			},
		},
		{
			name: "missing tls cipher suites is okay",
			server: tls.ServerConfig{
				TLSCert:           pki.ServerCertPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
				TLSCipherSuites:   []string{pki.CipherSuite()},
			},
		},
		{
			name: "missing tls max version is okay",
			server: tls.ServerConfig{
				TLSCert:           pki.ServerCertPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
				TLSCipherSuites:   []string{pki.CipherSuite()},
				TLSMaxVersion:     pki.TLSMaxVersion(),
			},
		},
		{
			name: "missing tls min version is okay",
			server: tls.ServerConfig{
				TLSCert:           pki.ServerCertPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
				TLSCipherSuites:   []string{pki.CipherSuite()},
				TLSMinVersion:     pki.TLSMinVersion(),
			},
		},
		{
			name: "missing tls min/max versions is okay",
			server: tls.ServerConfig{
				TLSCert:           pki.ServerCertPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
				TLSCipherSuites:   []string{pki.CipherSuite()},
			},
		},
		{
			name: "invalid ca",
			server: tls.ServerConfig{
				TLSCert:           pki.ServerCertPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.ServerKeyPath()},
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "missing allowed ca is okay",
			server: tls.ServerConfig{
				TLSCert: pki.ServerCertPath(),
				TLSKey:  pki.ServerKeyPath(),
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "invalid cert",
			server: tls.ServerConfig{
				TLSCert:           pki.ServerKeyPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "missing cert",
			server: tls.ServerConfig{
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "missing key",
			server: tls.ServerConfig{
				TLSCert:           pki.ServerCertPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "invalid cipher suites",
			server: tls.ServerConfig{
				TLSCert:           pki.ServerCertPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
				TLSCipherSuites:   []string{pki.CACertPath()},
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "TLS Max Version less than TLS Min version",
			server: tls.ServerConfig{
				TLSCert:           pki.ServerCertPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
				TLSCipherSuites:   []string{pki.CACertPath()},
				TLSMinVersion:     pki.TLSMaxVersion(),
				TLSMaxVersion:     pki.TLSMinVersion(),
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "invalid tls min version",
			server: tls.ServerConfig{
				TLSCert:           pki.ServerCertPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
				TLSCipherSuites:   []string{pki.CipherSuite()},
				TLSMinVersion:     pki.ServerKeyPath(),
				TLSMaxVersion:     pki.TLSMaxVersion(),
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "invalid tls max version",
			server: tls.ServerConfig{
				TLSCert:           pki.ServerCertPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
				TLSCipherSuites:   []string{pki.CACertPath()},
				TLSMinVersion:     pki.TLSMinVersion(),
				TLSMaxVersion:     pki.ServerCertPath(),
			},
			expNil: true,
			expErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tlsConfig, err := tt.server.TLSConfig()
			if !tt.expNil {
				require.NotNil(t, tlsConfig)
			}
			if !tt.expErr {
				require.NoError(t, err)
			}
		})
	}
}
