package tls_test

import (
	cryptotls "crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/testutil"
)

var pki = testutil.NewPKI("../../../testutil/pki")

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
			name: "success with tls key password set",
			client: tls.ClientConfig{
				TLSCA:     pki.CACertPath(),
				TLSCert:   pki.ClientCertPath(),
				TLSKey:    pki.ClientKeyPath(),
				TLSKeyPwd: "",
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
		{
			name: "set SNI server name",
			client: tls.ClientConfig{
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

func TestConnect(t *testing.T) {
	clientConfig := tls.ClientConfig{
		TLSCA:   pki.CACertPath(),
		TLSCert: pki.ClientCertPath(),
		TLSKey:  pki.ClientKeyPath(),
	}

	serverConfig := tls.ServerConfig{
		TLSCert:            pki.ServerCertPath(),
		TLSKey:             pki.ServerKeyPath(),
		TLSAllowedCACerts:  []string{pki.CACertPath()},
		TLSAllowedDNSNames: []string{"localhost", "127.0.0.1"},
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

	defer resp.Body.Close()
	require.Equal(t, 200, resp.StatusCode)
}

func TestConnectClientMinTLSVersion(t *testing.T) {
	serverConfig := tls.ServerConfig{
		TLSCert:            pki.ServerCertPath(),
		TLSKey:             pki.ServerKeyPath(),
		TLSAllowedCACerts:  []string{pki.CACertPath()},
		TLSAllowedDNSNames: []string{"localhost", "127.0.0.1"},
	}

	tests := []struct {
		name string
		cfg  tls.ClientConfig
	}{
		{
			name: "TLS version default",
			cfg: tls.ClientConfig{
				TLSCA:   pki.CACertPath(),
				TLSCert: pki.ClientCertPath(),
				TLSKey:  pki.ClientKeyPath(),
			},
		},
		{
			name: "TLS version 1.0",
			cfg: tls.ClientConfig{
				TLSCA:         pki.CACertPath(),
				TLSCert:       pki.ClientCertPath(),
				TLSKey:        pki.ClientKeyPath(),
				TLSMinVersion: "TLS10",
			},
		},
		{
			name: "TLS version 1.1",
			cfg: tls.ClientConfig{
				TLSCA:         pki.CACertPath(),
				TLSCert:       pki.ClientCertPath(),
				TLSKey:        pki.ClientKeyPath(),
				TLSMinVersion: "TLS11",
			},
		},
		{
			name: "TLS version 1.2",
			cfg: tls.ClientConfig{
				TLSCA:         pki.CACertPath(),
				TLSCert:       pki.ClientCertPath(),
				TLSKey:        pki.ClientKeyPath(),
				TLSMinVersion: "TLS12",
			},
		},
		{
			name: "TLS version 1.3",
			cfg: tls.ClientConfig{
				TLSCA:         pki.CACertPath(),
				TLSCert:       pki.ClientCertPath(),
				TLSKey:        pki.ClientKeyPath(),
				TLSMinVersion: "TLS13",
			},
		},
	}

	tlsVersions := []uint16{
		cryptotls.VersionTLS10,
		cryptotls.VersionTLS11,
		cryptotls.VersionTLS12,
		cryptotls.VersionTLS13,
	}

	tlsVersionNames := []string{
		"TLS 1.0",
		"TLS 1.1",
		"TLS 1.2",
		"TLS 1.3",
	}

	for _, tt := range tests {
		clientTLSConfig, err := tt.cfg.TLSConfig()
		require.NoError(t, err)

		client := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: clientTLSConfig,
			},
			Timeout: 1 * time.Second,
		}

		clientMinVersion := clientTLSConfig.MinVersion
		if tt.cfg.TLSMinVersion == "" {
			clientMinVersion = tls.TLSMinVersionDefault
		}

		for i, serverTLSMaxVersion := range tlsVersions {
			serverVersionName := tlsVersionNames[i]
			t.Run(tt.name+" vs "+serverVersionName, func(t *testing.T) {
				// Constrain the server's maximum TLS version
				serverTLSConfig, err := serverConfig.TLSConfig()
				require.NoError(t, err)
				serverTLSConfig.MinVersion = cryptotls.VersionTLS10
				serverTLSConfig.MaxVersion = serverTLSMaxVersion

				// Start the server
				ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
				ts.TLS = serverTLSConfig
				ts.StartTLS()

				// Do the connection and cleanup
				resp, err := client.Get(ts.URL)
				ts.Close()

				// Things should fail if the currently tested "serverTLSMaxVersion"
				// is below the client minimum version.
				if serverTLSMaxVersion < clientMinVersion {
					require.ErrorContains(t, err, "tls: protocol version not supported")
				} else {
					require.NoErrorf(t, err, "server=%v client=%v", serverTLSMaxVersion, clientMinVersion)
					require.Equal(t, 200, resp.StatusCode)
					resp.Body.Close()
				}
			})
		}
	}
}

func TestConnectClientInvalidMinTLSVersion(t *testing.T) {
	clientConfig := tls.ClientConfig{
		TLSCA:         pki.CACertPath(),
		TLSCert:       pki.ClientCertPath(),
		TLSKey:        pki.ClientKeyPath(),
		TLSMinVersion: "garbage",
	}

	_, err := clientConfig.TLSConfig()
	expected := `could not parse tls min version "garbage": unsupported version "garbage" (available: TLS10,TLS11,TLS12,TLS13)`
	require.EqualError(t, err, expected)
}

func TestConnectWrongDNS(t *testing.T) {
	clientConfig := tls.ClientConfig{
		TLSCA:   pki.CACertPath(),
		TLSCert: pki.ClientCertPath(),
		TLSKey:  pki.ClientKeyPath(),
	}

	serverConfig := tls.ServerConfig{
		TLSCert:            pki.ServerCertPath(),
		TLSKey:             pki.ServerKeyPath(),
		TLSAllowedCACerts:  []string{pki.CACertPath()},
		TLSAllowedDNSNames: []string{"localhos", "127.0.0.2"},
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
	require.Error(t, err)
	if resp != nil {
		err = resp.Body.Close()
		require.NoError(t, err)
	}
}
