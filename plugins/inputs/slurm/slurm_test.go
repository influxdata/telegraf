package slurm

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestGoodURLs(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"http", "http://example.com:6820"},
		{"https", "https://example.com:6820"},
		{"http no port", "http://example.com"},
		{"https no port", "https://example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := Slurm{
				URL: tt.url,
			}
			require.NoError(t, plugin.Init())
		})
	}
}

func TestWrongURLs(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"wrong http scheme", "httpp://example.com:6820"},
		{"wrong https scheme", "httpss://example.com:6820"},
		{"empty url", ""},
		{"empty hostname", "http://:6820"},
		{"only scheme", "http://"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := Slurm{
				URL: tt.url,
			}
			require.Error(t, plugin.Init())
		})
	}
}

func TestWrongEndpoints(t *testing.T) {
	tests := []struct {
		name             string
		enabledEndpoints []string
	}{
		{"empty endpoint", []string{"diag", "", "jobs"}},
		{"mistyped endpoint", []string{"diagg", "jobs", "partitions"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := Slurm{
				URL:              "http://example.net",
				EnabledEndpoints: tt.enabledEndpoints,
			}
			require.Error(t, plugin.Init())
		})
	}
}

func TestCases(t *testing.T) {
	entries, err := os.ReadDir("testcases")
	require.NoError(t, err)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			testcasePath := filepath.Join("testcases", entry.Name())
			responsesPath := filepath.Join(testcasePath, "responses")
			expectedFilename := filepath.Join(testcasePath, "expected.out")
			configFilename := filepath.Join(testcasePath, "telegraf.conf")

			responses, err := os.ReadDir(responsesPath)
			require.NoError(t, err)

			pathToResponse := map[string][]byte{}
			for _, response := range responses {
				if response.IsDir() {
					continue
				}
				fName := response.Name()
				buf, err := os.ReadFile(filepath.Join(responsesPath, fName))
				require.NoError(t, err)
				pathToResponse[strings.TrimSuffix(fName, filepath.Ext(fName))] = buf
			}

			// Prepare the influx parser for expectations
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())

			// Read expected values, if any
			var expected []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				var err error
				expected, err = testutil.ParseMetricsFromFile(expectedFilename, parser)
				require.NoError(t, err)
			}

			ts := httptest.NewServer(http.NotFoundHandler())
			defer ts.Close()

			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp, ok := pathToResponse[strings.TrimPrefix(r.URL.Path, "/slurm/v0.0.38/")]
				if !ok {
					w.WriteHeader(http.StatusInternalServerError)
					t.Errorf("Expected to have path to response: %s", r.URL.Path)
					return
				}
				w.Header().Add("Content-Type", "application/json")

				if _, err := w.Write(resp); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					t.Error(err)
					return
				}
				w.WriteHeader(http.StatusOK)
			})

			// Load the test-specific configuration
			cfg := config.NewConfig()
			cfg.Agent.Quiet = true
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			// Instantiate the plugin. As seen on NewConfig's documentation,
			// parsing the configuration will instantiate the plugins, so that
			// we only need to assert the plugin's type!
			plugin := cfg.Inputs[0].Input.(*Slurm)
			plugin.URL = "http://" + ts.Listener.Addr().String()
			plugin.Log = testutil.Logger{}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics(), testutil.IgnoreTime())
		})
	}
}
