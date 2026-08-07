package prometheus_client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/testutil"
)

func TestLandingPage(t *testing.T) {
	output := PrometheusClient{
		Listen:            ":0",
		CollectorsExclude: []string{"process"},
		MetricVersion:     1,
		Log:               &testutil.Logger{Name: "outputs.prometheus_client"},
	}
	expected := "Telegraf Output Plugin: Prometheus Client"

	require.NoError(t, output.Init())
	require.NoError(t, output.Connect())

	u, err := url.Parse(fmt.Sprintf("http://%s/", output.url.Host))
	require.NoError(t, err)

	resp, err := http.Get(u.String())
	require.NoError(t, err)
	defer resp.Body.Close()

	actual, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, expected, strings.TrimSpace(string(actual)))
}

func TestFormatHeader(t *testing.T) {
	tests := []struct {
		name     string
		accept   string
		expected string
	}{
		{
			name:     "invalid accept ",
			accept:   "applications/json",
			expected: "text/plain; version=0.0.4; charset=utf-8; escaping=underscores",
		},
		{
			name:     "no accept header",
			expected: "text/plain; version=0.0.4; charset=utf-8; escaping=underscores",
		},
		{
			name:     "text no version",
			accept:   "text/plain",
			expected: "text/plain; version=0.0.4; charset=utf-8; escaping=underscores",
		},
		{
			name:     "text with version 0.0.4",
			accept:   "text/plain; version=0.0.4",
			expected: "text/plain; version=0.0.4; charset=utf-8; escaping=underscores",
		},
		{
			name:     "protobuf text format",
			accept:   "application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily; encoding=text",
			expected: "application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily; encoding=text; escaping=underscores",
		},
		{
			name:     "protobuf compact text format",
			accept:   "application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily; encoding=compact-text",
			expected: "application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily; encoding=compact-text; escaping=underscores",
		},
		{
			name:     "protobuf delimited format",
			accept:   "application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily; encoding=delimited",
			expected: "application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily; encoding=delimited; escaping=underscores",
		},
		{
			name:     "multiple accept preferring protobuf",
			accept:   "application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily; encoding=delimited, text/plain",
			expected: "application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily; encoding=delimited; escaping=underscores",
		},
		{
			name:     "multiple accept preferring text",
			accept:   "text/plain, application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily; encoding=delimited",
			expected: "text/plain; version=0.0.4; charset=utf-8; escaping=underscores",
		},
	}

	// Setup the plugin
	plugin := PrometheusClient{
		Listen: ":0",
		Log:    testutil.Logger{Name: "outputs.prometheus_client"},
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())

	// Get the plugin's address so we can send data to it
	addr := fmt.Sprintf("http://%s/metrics", plugin.url.Host)

	// Run the actual tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Construct a request with the given "Accept" header
			req, err := http.NewRequest("GET", addr, nil)
			require.NoError(t, err)
			if tt.accept != "" {
				req.Header.Add("Accept", tt.accept)
			}

			// Get the metrics
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Test the result
			require.NotEmpty(t, resp.Body)
			require.Equal(t, tt.expected, resp.Header.Get("Content-Type"))
		})
	}
}

func TestNameSanitizationValidation(t *testing.T) {
	tests := []struct {
		name             string
		nameSanitization string
		expected         string
		err              string
	}{
		{
			name:             "legacy",
			nameSanitization: "legacy",
			expected:         "legacy",
		},
		{
			name:             "utf8",
			nameSanitization: "utf8",
			expected:         "utf8",
		},
		{
			name:             "empty value defaults to legacy",
			nameSanitization: "",
			expected:         defaultNameSanitization,
		},
		{
			name:             "invalid value",
			nameSanitization: "gzip",
			err:              "invalid name_sanitization \"gzip\": must be \"legacy\" or \"utf8\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := PrometheusClient{
				Listen:            ":0",
				MetricVersion:     1,
				CollectorsExclude: []string{"gocollector", "process"},
				NameSanitization:  tt.nameSanitization,
				Log:               testutil.Logger{Name: "outputs.prometheus_client"},
			}

			err := plugin.Init()
			if tt.err != "" {
				require.EqualError(t, err, tt.err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.expected, plugin.NameSanitization)
		})
	}
}

func TestNameSanitizationDefaultFromConstructor(t *testing.T) {
	creator, found := outputs.Outputs["prometheus_client"]
	require.True(t, found)

	plugin := creator().(*PrometheusClient)
	require.Equal(t, defaultNameSanitization, plugin.NameSanitization)
	require.NoError(t, plugin.Init())
}
