package tls_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal/tls"
	"github.com/stretchr/testify/require"
)

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
				TLSCA:   "./pki/cacert.pem",
				TLSCert: "./pki/clientcert.pem",
				TLSKey:  "./pki/clientkey.pem",
			},
		},
		{
			name: "invalid ca",
			client: tls.ClientConfig{
				TLSCA:   "./pki/invalid.pem",
				TLSCert: "./pki/clientcert.pem",
				TLSKey:  "./pki/clientkey.pem",
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "missing ca is okay",
			client: tls.ClientConfig{
				TLSCert: "./pki/clientcert.pem",
				TLSKey:  "./pki/clientkey.pem",
			},
		},
		{
			name: "invalid cert",
			client: tls.ClientConfig{
				TLSCA:   "./pki/cacert.pem",
				TLSCert: "./pki/invalid.pem",
				TLSKey:  "./pki/clientkey.pem",
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "missing cert",
			client: tls.ClientConfig{
				TLSCA:  "./pki/cacert.pem",
				TLSKey: "./pki/clientkey.pem",
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "missing key",
			client: tls.ClientConfig{
				TLSCA:   "./pki/cacert.pem",
				TLSCert: "./pki/clientcert.pem",
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "ssl option names",
			client: tls.ClientConfig{
				SSLCA:   "./pki/cacert.pem",
				SSLCert: "./pki/clientcert.pem",
				SSLKey:  "./pki/clientkey.pem",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tlsConfig, err := tls.NewClientTLSConfig(tt.client)
			if !tt.expNil {
				require.NotNil(t, tlsConfig)
			}
			if !tt.expErr {
				require.NoError(t, err)
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
				TLSCert:           "./pki/servercert.pem",
				TLSKey:            "./pki/serverkey.pem",
				TLSAllowedCACerts: []string{"./pki/cacert.pem"},
			},
		},
		{
			name: "invalid ca",
			server: tls.ServerConfig{
				TLSCert:           "./pki/servercert.pem",
				TLSKey:            "./pki/serverkey.pem",
				TLSAllowedCACerts: []string{"./pki/invalid.pem"},
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "missing allowed ca is okay",
			server: tls.ServerConfig{
				TLSCert: "./pki/servercert.pem",
				TLSKey:  "./pki/serverkey.pem",
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "invalid cert",
			server: tls.ServerConfig{
				TLSCert:           "./pki/invalid.pem",
				TLSKey:            "./pki/serverkey.pem",
				TLSAllowedCACerts: []string{"./testdata/cacert.pem"},
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "missing cert",
			server: tls.ServerConfig{
				TLSKey:            "./pki/serverkey.pem",
				TLSAllowedCACerts: []string{"./pki/cacert.pem"},
			},
			expNil: true,
			expErr: true,
		},
		{
			name: "missing key",
			server: tls.ServerConfig{
				TLSCert:           "./pki/servercert.pem",
				TLSAllowedCACerts: []string{"./pki/cacert.pem"},
			},
			expNil: true,
			expErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tlsConfig, err := tls.NewServerTLSConfig(tt.server)
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
		TLSCA:   "./pki/cacert.pem",
		TLSCert: "./pki/clientcert.pem",
		TLSKey:  "./pki/clientkey.pem",
	}

	serverConfig := tls.ServerConfig{
		TLSAllowedCACerts: []string{"./pki/cacert.pem"},
		TLSCert:           "./pki/servercert.pem",
		TLSKey:            "./pki/serverkey.pem",
	}

	serverTLSConfig, err := tls.NewServerTLSConfig(serverConfig)
	require.NoError(t, err)

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	ts.TLS = serverTLSConfig

	ts.StartTLS()
	defer ts.Close()

	clientTLSConfig, err := tls.NewClientTLSConfig(clientConfig)
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
