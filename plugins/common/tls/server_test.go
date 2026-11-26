package tls_test

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestServerConfigAllowedClients(t *testing.T) {
	// The client certificate has CN=localhost, the SANs DNS:localhost and
	// IP:127.0.0.1, serial number 3 and is issued by "Telegraf Test CA".
	tests := []struct {
		name         string
		allowed      []string
		dnsNames     []string
		expectedDeny bool
	}{
		{
			name: "empty filter",
		},
		{
			name:    "allowed by cn",
			allowed: []string{"cn:localhost"},
		},
		{
			name:    "allowed by dns glob",
			allowed: []string{"dns:local*"},
		},
		{
			name:    "allowed by ip range",
			allowed: []string{"ip:127.0.0.0/8"},
		},
		{
			name:    "allowed by serial number",
			allowed: []string{"sn:3"},
		},
		{
			name:    "allowed by issuer",
			allowed: []string{"issuer:Telegraf Test CA"},
		},
		{
			name:    "allowed by any expression",
			allowed: []string{"cn:foo", "ip:192.0.2.0/24", "dns:localhost"},
		},
		{
			name:     "allowed by deprecated dns names",
			dnsNames: []string{"localhost"},
		},
		{
			name:     "allowed by new option with deprecated one present",
			allowed:  []string{"cn:localhost"},
			dnsNames: []string{"foo.example.org"},
		},
		{
			name:         "denied by cn",
			allowed:      []string{"cn:foo"},
			expectedDeny: true,
		},
		{
			name:         "denied by ip range",
			allowed:      []string{"ip:192.0.2.0/24"},
			expectedDeny: true,
		},
		{
			name:         "denied by missing attribute",
			allowed:      []string{"org:*"},
			expectedDeny: true,
		},
		{
			name:         "denied by issuer",
			allowed:      []string{"issuer:localhost"},
			expectedDeny: true,
		},
		{
			name:         "denied by all expressions",
			allowed:      []string{"cn:foo", "ip:192.0.2.0/24", "dns:*.example.org"},
			expectedDeny: true,
		},
		{
			name:         "denied by deprecated dns names",
			dnsNames:     []string{"foo.example.org"},
			expectedDeny: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serverConfig := tls.ServerConfig{
				TLSCert:            pki.ServerCertPath(),
				TLSKey:             pki.ServerKeyPath(),
				TLSAllowedCACerts:  []string{pki.CACertPath()},
				TLSAllowedClients:  tt.allowed,
				TLSAllowedDNSNames: tt.dnsNames,
			}
			serverTLSConfig, err := serverConfig.TLSConfig()
			require.NoError(t, err)

			ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			ts.TLS = serverTLSConfig
			ts.Config.ErrorLog = log.New(io.Discard, "", 0)
			ts.StartTLS()
			defer ts.Close()

			clientConfig := tls.ClientConfig{
				TLSCA:   pki.CACertPath(),
				TLSCert: pki.ClientCertPath(),
				TLSKey:  pki.ClientKeyPath(),
			}
			clientTLSConfig, err := clientConfig.TLSConfig()
			require.NoError(t, err)

			client := http.Client{
				Transport: &http.Transport{TLSClientConfig: clientTLSConfig},
				Timeout:   10 * time.Second,
			}
			defer client.CloseIdleConnections()

			resp, err := client.Get(ts.URL)
			if tt.expectedDeny {
				require.ErrorContains(t, err, "bad certificate")
				return
			}
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestServerConfigAllowedClientsInvalid(t *testing.T) {
	serverConfig := tls.ServerConfig{
		TLSCert:           pki.ServerCertPath(),
		TLSKey:            pki.ServerKeyPath(),
		TLSAllowedCACerts: []string{pki.CACertPath()},
		TLSAllowedClients: []string{"foo:bar"},
	}
	_, err := serverConfig.TLSConfig()
	require.ErrorContains(t, err, "compiling client certificate filter failed")
}
