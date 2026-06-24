package tls_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/common/tls"
)

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

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
