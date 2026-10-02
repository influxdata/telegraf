package fxmacrodata

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestInitFail(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *FXMacroData
		expected string
	}{
		{
			name:     "invalid scheme",
			plugin:   &FXMacroData{URL: "ftp://example.com", Indicators: []string{"inflation"}},
			expected: `invalid scheme "ftp"`,
		},
		{
			name:     "nothing to gather",
			plugin:   &FXMacroData{URL: "https://example.com/v1"},
			expected: "at least one of indicators or fx_pairs must be set",
		},
		{
			name: "lowercase currency",
			plugin: &FXMacroData{
				URL:        "https://example.com/v1",
				Currencies: []string{"usd"},
				Indicators: []string{"inflation"},
			},
			expected: `invalid currency "usd"`,
		},
		{
			name: "empty indicator",
			plugin: &FXMacroData{
				URL:        "https://example.com/v1",
				Indicators: []string{"inflation", ""},
			},
			expected: "empty entry in indicators",
		},
		{
			name:     "pair without separator",
			plugin:   &FXMacroData{URL: "https://example.com/v1", FXPairs: []string{"EURUSD"}},
			expected: `invalid fx_pair "EURUSD"`,
		},
		{
			name:     "lowercase pair",
			plugin:   &FXMacroData{URL: "https://example.com/v1", FXPairs: []string{"eur/usd"}},
			expected: `invalid fx_pair "eur/usd"`,
		},
		{
			name: "negative timeout",
			plugin: &FXMacroData{
				URL:             "https://example.com/v1",
				Indicators:      []string{"inflation"},
				ResponseTimeout: config.Duration(-time.Second),
			},
			expected: "response_timeout must not be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plugin.Log = testutil.Logger{}
			require.ErrorContains(t, tt.plugin.Init(), tt.expected)
		})
	}
}

func TestInitDefaultURL(t *testing.T) {
	plugin := &FXMacroData{
		Indicators: []string{"inflation"},
		Log:        testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
	require.Equal(t, "https://api.fxmacrodata.com/v1", plugin.URL)
	require.Len(t, plugin.series, 1)
	require.Equal(t, "https://api.fxmacrodata.com/v1/announcements/usd/inflation?limit=1", plugin.series[0].endpoint)
}

func TestInitZeroTimeout(t *testing.T) {
	plugin := &FXMacroData{
		URL:        "https://example.com/v1",
		Indicators: []string{"inflation"},
		Log:        testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
	require.Zero(t, plugin.client.Timeout)
}

func TestAPIKeyHeader(t *testing.T) {
	var header, query string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header = r.Header.Get("X-API-Key")
		query = r.URL.RawQuery
		if _, err := w.Write([]byte(`{"data":[]}`)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
		}
	}))
	defer server.Close()

	plugin := &FXMacroData{
		URL:        server.URL + "/v1",
		APIKey:     config.NewSecret([]byte("test-key")),
		Indicators: []string{"inflation"},
		Log:        testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.Errors)

	require.Equal(t, "test-key", header)
	require.Equal(t, "limit=1", query)
}

func TestNoAPIKeyHeader(t *testing.T) {
	var found bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, found = r.Header["X-Api-Key"]
		if _, err := w.Write([]byte(`{"data":[]}`)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
		}
	}))
	defer server.Close()

	plugin := &FXMacroData{
		URL:        server.URL + "/v1",
		Indicators: []string{"inflation"},
		Log:        testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.Errors)

	require.False(t, found)
}

func TestMissingSeriesRemoved(t *testing.T) {
	requests := make(map[string]int)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests[r.URL.Path]++
		if r.URL.Path == "/v1/announcements/usd/doesnotexist" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if _, err := w.Write([]byte(`{"data":[]}`)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
		}
	}))
	defer server.Close()

	plugin := &FXMacroData{
		URL:        server.URL + "/v1",
		Indicators: []string{"inflation", "doesnotexist"},
		Log:        testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.Errors)

	require.Equal(t, map[string]int{
		"/v1/announcements/usd/inflation":    2,
		"/v1/announcements/usd/doesnotexist": 1,
	}, requests)
	require.Len(t, plugin.series, 1)
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
			expectedErrorFilename := filepath.Join(testcasePath, "expected.err")
			configFilename := filepath.Join(testcasePath, "telegraf.conf")

			// Read the responses. The file name is the request path relative
			// to the API version with slashes replaced by underscores. The
			// extension is either "json" for a successful response or the
			// HTTP status code to reply with.
			responses, err := os.ReadDir(responsesPath)
			require.NoError(t, err)

			pathToResponse := make(map[string][]byte, len(responses))
			pathToStatus := make(map[string]int, len(responses))
			for _, response := range responses {
				if response.IsDir() {
					continue
				}
				fName := response.Name()
				buf, err := os.ReadFile(filepath.Join(responsesPath, fName))
				require.NoError(t, err)

				ext := filepath.Ext(fName)
				key := strings.TrimSuffix(fName, ext)
				pathToResponse[key] = buf
				pathToStatus[key] = http.StatusOK
				if ext != ".json" {
					status, err := strconv.Atoi(strings.TrimPrefix(ext, "."))
					require.NoError(t, err)
					pathToStatus[key] = status
				}
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

			// Read expected errors, if any
			var expectedErrors []string
			if _, err := os.Stat(expectedErrorFilename); err == nil {
				var err error
				expectedErrors, err = testutil.ParseLinesFromFile(expectedErrorFilename)
				require.NoError(t, err)
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("limit") != "1" {
					w.WriteHeader(http.StatusBadRequest)
					t.Errorf("Unexpected query %q", r.URL.RawQuery)
					return
				}

				key := strings.ReplaceAll(strings.TrimPrefix(r.URL.Path, "/v1/"), "/", "_")
				resp, ok := pathToResponse[key]
				if !ok {
					w.WriteHeader(http.StatusInternalServerError)
					t.Errorf("Expected to have path to response: %s", r.URL.Path)
					return
				}

				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(pathToStatus[key])
				if _, err := w.Write(resp); err != nil {
					t.Error(err)
				}
			}))
			defer server.Close()

			// Load the test-specific configuration
			cfg := config.NewConfig()
			cfg.Agent.Quiet = true
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			// Instantiate the plugin and point it to the test server
			plugin := cfg.Inputs[0].Input.(*FXMacroData)
			plugin.URL = server.URL + "/v1"
			plugin.Log = testutil.Logger{}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))

			actualErrors := make([]string, 0, len(acc.Errors))
			for _, err := range acc.Errors {
				actualErrors = append(actualErrors, strings.ReplaceAll(err.Error(), server.URL, ""))
			}
			require.ElementsMatch(t, expectedErrors, actualErrors)

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.SortMetrics())
		})
	}
}
