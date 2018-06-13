package tls_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var pki = testutil.NewPKI("../../testutil/pki")

func TestClientConfig(t *testing.T) {
	tests := []struct {
		name   string
		client tls.ClientConfig
		expNil bool
		expErr bool
	}{
		{
			name:   "unset",
			client: tls.ClientConfig{},
			expNil: true,
		},
		{
			name: "success",
			client: tls.ClientConfig{
				TLSCA:   pki.CACertPath(),
				TLSCert: pki.ClientCertPath(),
				TLSKey:  pki.ClientKeyPath(),
			},
		},
		{
			name: "invalid ca",
			client: tls.ClientConfig{
				TLSCA:   pki.ClientKeyPath(),
				TLSCert: pki.ClientCertPath(),
				TLSKey:  pki.ClientKeyPath(),
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "missing ca is okay",
			client: tls.ClientConfig{
				TLSCert: pki.ClientCertPath(),
				TLSKey:  pki.ClientKeyPath(),
			},
		},
		{
			name: "invalid cert",
			client: tls.ClientConfig{
				TLSCA:   pki.CACertPath(),
				TLSCert: pki.ClientKeyPath(),
				TLSKey:  pki.ClientKeyPath(),
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "missing cert skips client keypair",
			client: tls.ClientConfig{
				TLSCA:  pki.CACertPath(),
				TLSKey: pki.ClientKeyPath(),
			},
			expNil: false,
			expErr: false,
		},
		{
			name: "missing key skips client keypair",
			client: tls.ClientConfig{
				TLSCA:   pki.CACertPath(),
				TLSCert: pki.ClientCertPath(),
			},
			expNil: false,
			expErr: false,
		},
		{
			name: "support deprecated ssl field names",
			client: tls.ClientConfig{
				SSLCA:   pki.CACertPath(),
				SSLCert: pki.ClientCertPath(),
				SSLKey:  pki.ClientKeyPath(),
			},
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
				TLSCert:           pki.ServerCertPath(),
				TLSKey:            pki.ServerKeyPath(),
				TLSAllowedCACerts: []string{pki.CACertPath()},
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
	clientConfig := tls.ClientConfig{
		TLSCA:   pki.CACertPath(),
		TLSCert: pki.ClientCertPath(),
		TLSKey:  pki.ClientKeyPath(),
	}

	serverConfig := tls.ServerConfig{
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
