package prometheus_client_test

import (
	"crypto/tls"
	"fmt"
	"github.com/influxdata/telegraf/plugins/outputs/prometheus_client"
	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/toml"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

var pki = testutil.NewPKI("../../../testutil/pki")

var configWithTLS = fmt.Sprintf(`
 listen = "127.0.0.1:9090"
 tls_allowed_cacerts = ["%s"]
 tls_cert = "%s"
 tls_key = "%s"
`, pki.TLSServerConfig().TLSAllowedCACerts[0], pki.TLSServerConfig().TLSCert, pki.TLSServerConfig().TLSKey)

var configWithoutTLS = `
  listen = "127.0.0.1:9090"
`

type PrometheusClientTestContext struct {
	Output      *prometheus_client.PrometheusClient
	Accumulator *testutil.Accumulator
	Client      *http.Client
}

func TestWorksWithoutTLS(t *testing.T) {
	tc := buildTestContext(t, []byte(configWithoutTLS))
	err := tc.Output.Connect()
	defer tc.Output.Close()

	require.NoError(t, err)

	response, err := tc.Client.Get("http://localhost:9090/metrics")
	require.NoError(t, err)

	require.NoError(t, err)
	require.Equal(t, response.StatusCode, http.StatusOK)
}

func TestWorksWithTLS(t *testing.T) {
	tc := buildTestContext(t, []byte(configWithTLS))
	err := tc.Output.Connect()
	defer tc.Output.Close()
	require.NoError(t, err)

	response, err := tc.Client.Get("https://localhost:9090/metrics")
	require.NoError(t, err)

	require.NoError(t, err)
	require.Equal(t, response.StatusCode, http.StatusOK)

	response, err = tc.Client.Get("http://localhost:9090/metrics")
	require.Error(t, err)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	response, err = client.Get("https://localhost:9090/metrics")

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
