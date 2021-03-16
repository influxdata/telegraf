package tls_test

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tlsConfig "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var pki = testutil.NewPKI("../../../testutil/pki")

func TestClientConfig(t *testing.T) {
	tests := []struct {
		name   string
		client tlsConfig.ClientConfig
		expNil bool
		expErr bool
	}{
		{
			name:   "unset",
			client: tlsConfig.ClientConfig{},
			expNil: true,
		},
		{
			name: "success",
			client: tlsConfig.ClientConfig{
				TLSCA:   pki.CACertPath(),
				TLSCert: pki.ClientCertPath(),
				TLSKey:  pki.ClientKeyPath(),
			},
		},
		{
			name: "invalid ca",
			client: tlsConfig.ClientConfig{
				TLSCA:   pki.ClientKeyPath(),
				TLSCert: pki.ClientCertPath(),
				TLSKey:  pki.ClientKeyPath(),
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "missing ca is okay",
			client: tlsConfig.ClientConfig{
				TLSCert: pki.ClientCertPath(),
				TLSKey:  pki.ClientKeyPath(),
			},
		},
		{
			name: "invalid cert",
			client: tlsConfig.ClientConfig{
				TLSCA:   pki.CACertPath(),
				TLSCert: pki.ClientKeyPath(),
				TLSKey:  pki.ClientKeyPath(),
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "missing cert skips client keypair",
			client: tlsConfig.ClientConfig{
				TLSCA:  pki.CACertPath(),
				TLSKey: pki.ClientKeyPath(),
			},
			expNil: false,
			expErr: false,
		},
		{
			name: "missing key skips client keypair",
			client: tlsConfig.ClientConfig{
				TLSCA:   pki.CACertPath(),
				TLSCert: pki.ClientCertPath(),
			},
			expNil: false,
			expErr: false,
		},
		{
			name: "support deprecated ssl field names",
			client: tlsConfig.ClientConfig{
				SSLCA:   pki.CACertPath(),
				SSLCert: pki.ClientCertPath(),
				SSLKey:  pki.ClientKeyPath(),
			},
		},
		{
			name: "set SNI server name",
			client: tlsConfig.ClientConfig{
				ServerName: "foo.example.com",
			},
			expNil: false,
			expErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tlsConfig, err := tt.client.TLSConfig()
			if !tt.expNil {
				require.NotNil(t, tlsConfig)
			} else {
				require.Nil(t, tlsConfig)
			}

			if !tt.expErr {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestClientConfigTLSMinVersion(t *testing.T) {
	tests := []struct {
		name    string
		client  tlsConfig.ClientConfig
		expVers uint16
		expErr  bool
	}{
		{
			name: "default",
			client: tlsConfig.ClientConfig{
				TLSCA:   pki.CACertPath(),
				TLSCert: pki.ClientCertPath(),
				TLSKey:  pki.ClientKeyPath(),
			},
			expVers: tls.VersionTLS12,
			expErr:  false,
		},
		{
			name: "TLS13",
			client: tlsConfig.ClientConfig{
				TLSCA:         pki.CACertPath(),
				TLSCert:       pki.ClientCertPath(),
				TLSKey:        pki.ClientKeyPath(),
				TLSMinVersion: "TLS13",
			},
			expVers: tls.VersionTLS13,
			expErr:  false,
		},
		{
			name: "explicit TLS12",
			client: tlsConfig.ClientConfig{
				TLSCA:         pki.CACertPath(),
				TLSCert:       pki.ClientCertPath(),
				TLSKey:        pki.ClientKeyPath(),
				TLSMinVersion: "TLS12",
			},
			expVers: tls.VersionTLS12,
			expErr:  false,
		},
		{
			name: "TLS11",
			client: tlsConfig.ClientConfig{
				TLSCA:         pki.CACertPath(),
				TLSCert:       pki.ClientCertPath(),
				TLSKey:        pki.ClientKeyPath(),
				TLSMinVersion: "TLS11",
			},
			expVers: tls.VersionTLS11,
			expErr:  false,
		},
		{
			name: "TLS10",
			client: tlsConfig.ClientConfig{
				TLSCA:         pki.CACertPath(),
				TLSCert:       pki.ClientCertPath(),
				TLSKey:        pki.ClientKeyPath(),
				TLSMinVersion: "TLS10",
			},
			expVers: tls.VersionTLS10,
			expErr:  false,
		},
		{
			name: "bad version",
			client: tlsConfig.ClientConfig{
				TLSCA:         pki.CACertPath(),
				TLSCert:       pki.ClientCertPath(),
				TLSKey:        pki.ClientKeyPath(),
				TLSMinVersion: "bad",
			},
			expErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tlsConfig, err := tt.client.TLSConfig()
			if !tt.expErr {
				require.NoError(t, err)
				require.NotNil(t, tlsConfig)
				require.Equal(t, tlsConfig.MinVersion, tt.expVers)
			} else {
				require.Error(t, err)
				require.Nil(t, tlsConfig)
			}
		})
	}
}

func TestServerConfig(t *testing.T) {
	tests := []struct {
		name   string
		server tlsConfig.ServerConfig
		expNil bool
		expErr bool
	}{
		{
			name:   "unset",
			server: tlsConfig.ServerConfig{},
			expNil: true,
		},
		{
			name: "success",
			server: tlsConfig.ServerConfig{
				TLSCert:           pki.ServerCertPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
				TLSCipherSuites:   []string{pki.CipherSuite()},
				TLSMinVersion:     pki.TLSMinVersion(),
				TLSMaxVersion:     pki.TLSMaxVersion(),
			},
		},
		{
			name: "missing tls cipher suites is okay",
			server: tlsConfig.ServerConfig{
				TLSCert:           pki.ServerCertPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
				TLSCipherSuites:   []string{pki.CipherSuite()},
			},
		},
		{
			name: "missing tls max version is okay",
			server: tlsConfig.ServerConfig{
				TLSCert:           pki.ServerCertPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
				TLSCipherSuites:   []string{pki.CipherSuite()},
				TLSMaxVersion:     pki.TLSMaxVersion(),
			},
		},
		{
			name: "missing tls min version is okay",
			server: tlsConfig.ServerConfig{
				TLSCert:           pki.ServerCertPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
				TLSCipherSuites:   []string{pki.CipherSuite()},
				TLSMinVersion:     pki.TLSMinVersion(),
			},
		},
		{
			name: "missing tls min/max versions is okay",
			server: tlsConfig.ServerConfig{
				TLSCert:           pki.ServerCertPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
				TLSCipherSuites:   []string{pki.CipherSuite()},
			},
		},
		{
			name: "invalid ca",
			server: tlsConfig.ServerConfig{
				TLSCert:           pki.ServerCertPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.ServerKeyPath()},
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "missing allowed ca is okay",
			server: tlsConfig.ServerConfig{
				TLSCert: pki.ServerCertPath(),
				TLSKey:  pki.ServerKeyPath(),
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "invalid cert",
			server: tlsConfig.ServerConfig{
				TLSCert:           pki.ServerKeyPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "missing cert",
			server: tlsConfig.ServerConfig{
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "missing key",
			server: tlsConfig.ServerConfig{
				TLSCert:           pki.ServerCertPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "invalid cipher suites",
			server: tlsConfig.ServerConfig{
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
			server: tlsConfig.ServerConfig{
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
			server: tlsConfig.ServerConfig{
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
			server: tlsConfig.ServerConfig{
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

func TestConnect(t *testing.T) {
	clientConfig := tlsConfig.ClientConfig{
		TLSCA:   pki.CACertPath(),
		TLSCert: pki.ClientCertPath(),
		TLSKey:  pki.ClientKeyPath(),
	}

	serverConfig := tlsConfig.ServerConfig{
		TLSCert:           pki.ServerCertPath(),
		TLSKey:            pki.ServerKeyPath(),
		TLSAllowedCACerts: []string{pki.CACertPath()},
	}

	serverTLSConfig, err := serverConfig.TLSConfig()
	require.NoError(t, err)

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	ts.TLS = serverTLSConfig

	ts.StartTLS()
	defer ts.Close()

	clientTLSConfig, err := clientConfig.TLSConfig()
	require.NoError(t, err)

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: clientTLSConfig,
		},
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(ts.URL)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
}

func TestConnectFailMinVersion(t *testing.T) {
	clientConfig := tlsConfig.ClientConfig{
		TLSCA:   pki.CACertPath(),
		TLSCert: pki.ClientCertPath(),
		TLSKey:  pki.ClientKeyPath(),
	}

	serverConfig := tlsConfig.ServerConfig{
		TLSCert:           pki.ServerCertPath(),
		TLSKey:            pki.ServerKeyPath(),
		TLSAllowedCACerts: []string{pki.CACertPath()},
		TLSMaxVersion:     "TLS11",
	}

	serverTLSConfig, err := serverConfig.TLSConfig()
	require.NoError(t, err)

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	ts.TLS = serverTLSConfig

	ts.StartTLS()
	defer ts.Close()

	clientTLSConfig, err := clientConfig.TLSConfig()
	require.NoError(t, err)

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: clientTLSConfig,
		},
		Timeout: 10 * time.Second,
	}

	_, err = client.Get(ts.URL)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "protocol version not supported"))
}
