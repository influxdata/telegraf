package prometheus_client_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"

	"github.com/influxdata/telegraf/plugins/outputs/prometheus_client"
	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/toml"
	"github.com/stretchr/testify/require"
)

var pki = testutil.NewPKI("../../../testutil/pki")

var configWithTLS = fmt.Sprintf(`
 listen = "127.0.0.1:0"
 tls_allowed_cacerts = ["%s"]
 tls_cert = "%s"
 tls_key = "%s"
`, pki.TLSServerConfig().TLSAllowedCACerts[0], pki.TLSServerConfig().TLSCert, pki.TLSServerConfig().TLSKey)

var configWithoutTLS = `
  listen = "127.0.0.1:0"
`

type PrometheusClientTestContext struct {
	Output      *prometheus_client.PrometheusClient
	Accumulator *testutil.Accumulator
	Client      *http.Client
}

func TestWorksWithoutTLS(t *testing.T) {
	tc := buildTestContext(t, []byte(configWithoutTLS))
	err := tc.Output.Connect()
	require.NoError(t, err)
	defer tc.Output.Close()

	response, err := tc.Client.Get(tc.Output.URL())
	require.NoError(t, err)

	require.NoError(t, err)
	require.Equal(t, response.StatusCode, http.StatusOK)
}

func TestWorksWithTLS(t *testing.T) {
	tc := buildTestContext(t, []byte(configWithTLS))
	err := tc.Output.Connect()
	require.NoError(t, err)
	defer tc.Output.Close()

	response, err := tc.Client.Get(tc.Output.URL())
	require.NoError(t, err)

	require.NoError(t, err)
	require.Equal(t, response.StatusCode, http.StatusOK)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	response, err = client.Get(tc.Output.URL())

	require.Error(t, err)
}

func buildTestContext(t *testing.T, config []byte) *PrometheusClientTestContext {
	output := prometheus_client.NewClient()
	err := toml.Unmarshal(config, output)
	require.NoError(t, err)

	var (
		httpClient *http.Client
	)

	if len(output.TLSAllowedCACerts) != 0 {
		httpClient = buildClientWithTLS(t, output)
	} else {
		httpClient = buildClientWithoutTLS()
	}

	return &PrometheusClientTestContext{
		Output:      output,
		Accumulator: &testutil.Accumulator{},
		Client:      httpClient,
	}
}

func buildClientWithoutTLS() *http.Client {
	return &http.Client{}
}

func buildClientWithTLS(t *testing.T, output *prometheus_client.PrometheusClient) *http.Client {
	tlsConfig, err := pki.TLSClientConfig().TLSConfig()
	require.NoError(t, err)

	transport := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: transport}
}
