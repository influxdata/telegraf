package logql

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestInitSuccess(t *testing.T) {
	username := config.NewSecret([]byte("john"))
	password := config.NewSecret([]byte("secret"))
	token := config.NewSecret([]byte("a token"))
	defer username.Destroy()
	defer password.Destroy()
	defer token.Destroy()

	tests := []struct {
		name   string
		plugin *LogQL
	}{
		{
			name: "no authentication",
			plugin: &LogQL{
				InstantQueries: []InstantQuery{{query: query{Query: `{job="varlogs"}`}}},
			},
		},
		{
			name: "basic auth without password",
			plugin: &LogQL{
				Username:       username,
				InstantQueries: []InstantQuery{{query: query{Query: `{job="varlogs"}`}}},
			},
		},
		{
			name: "basic auth with password",
			plugin: &LogQL{
				Username:       username,
				Password:       password,
				InstantQueries: []InstantQuery{{query: query{Query: `{job="varlogs"}`}}},
			},
		},
		{
			name: "token auth",
			plugin: &LogQL{
				Token:          token,
				InstantQueries: []InstantQuery{{query: query{Query: `{job="varlogs"}`}}},
			},
		},
		{
			name: "time range mix",
			plugin: &LogQL{
				RangeQueries: []RangeQuery{
					{
						Start: config.Duration(5 * time.Minute),
						End:   config.Duration(-1 * time.Minute),
						query: query{Query: `{job="varlogs"}`},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plugin.Log = testutil.Logger{}
			require.NoError(t, tt.plugin.Init())
			require.Equal(t, "http://localhost:3100", tt.plugin.URL)
		})
	}
}

func TestInitFail(t *testing.T) {
	username := config.NewSecret([]byte("john"))
	password := config.NewSecret([]byte("secret"))
	token := config.NewSecret([]byte("a token"))
	defer username.Destroy()
	defer password.Destroy()
	defer token.Destroy()

	tests := []struct {
		name     string
		plugin   *LogQL
		expected string
	}{
		{
			name:     "no queries",
			plugin:   &LogQL{},
			expected: "no queries configured",
		},
		{
			name: "invalid sorting",
			plugin: &LogQL{
				InstantQueries: []InstantQuery{{query: query{Query: `{job="varlogs"}`, Sorting: "random"}}},
			},
			expected: "invalid sorting direction",
		},
		{
			name: "invalid time range zero",
			plugin: &LogQL{
				RangeQueries: []RangeQuery{
					{
						query: query{Query: `{job="varlogs"}`},
					},
				},
			},
			expected: "invalid range",
		},
		{
			name: "invalid time range past",
			plugin: &LogQL{
				RangeQueries: []RangeQuery{
					{
						Start: config.Duration(5 * time.Minute),
						End:   config.Duration(15 * time.Minute),
						query: query{Query: `{job="varlogs"}`},
					},
				},
			},
			expected: "invalid range",
		},
		{
			name: "invalid time range future",
			plugin: &LogQL{
				RangeQueries: []RangeQuery{
					{
						Start: config.Duration(-5 * time.Minute),
						query: query{Query: `{job="varlogs"}`},
					},
				},
			},
			expected: "invalid range",
		},
		{
			name: "invalid stepping",
			plugin: &LogQL{
				RangeQueries: []RangeQuery{
					{
						Start: config.Duration(5 * time.Minute),
						Step:  config.Duration(-1 * time.Minute),
						query: query{Query: `{job="varlogs"}`},
					},
				},
			},
			expected: "'step' must be non-negative for query",
		},
		{
			name: "invalid interval",
			plugin: &LogQL{
				RangeQueries: []RangeQuery{
					{
						Start:    config.Duration(5 * time.Minute),
						Interval: config.Duration(-5 * time.Minute),
						query:    query{Query: `{job="varlogs"}`},
					},
				},
			},
			expected: "'interval' must be non-negative for query",
		},
		{
			name: "password without username",
			plugin: &LogQL{
				Password:       password,
				InstantQueries: []InstantQuery{{query: query{Query: `{job="varlogs"}`}}},
			},
			expected: "expecting username for basic authentication",
		},
		{
			name: "basic and token auth",
			plugin: &LogQL{
				Username:       username,
				Token:          token,
				InstantQueries: []InstantQuery{{query: query{Query: `{job="varlogs"}`}}},
			},
			expected: "cannot use both basic and bearer authentication",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plugin.Log = testutil.Logger{}
			require.ErrorContains(t, tt.plugin.Init(), tt.expected)
		})
	}
}

func TestSigleTenant(t *testing.T) {
	// Start a mock server
	var orgs []string
	var orgMu sync.Mutex
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mark the server as ready
		if r.URL.Path == "/ready" {
			if _, err := w.Write([]byte("ready")); err != nil {
				t.Logf("writing 'ready' response failed: %v", err)
				t.Fail()
			}
			return
		}

		// Store the query
		orgMu.Lock()
		orgs = r.Header.Values("X-Scope-OrgID")
		orgMu.Unlock()

		// Send the response
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	// Configure and initialize the plugin
	plugin := &LogQL{
		URL:            ts.URL,
		InstantQueries: []InstantQuery{{query: query{Query: `{job="varlogs"}`}}},
		Log:            &testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	// Start the plugin and collect data
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()
	require.NoError(t, plugin.Gather(&acc))

	// Check the query including the parameters
	orgMu.Lock()
	defer orgMu.Unlock()
	require.Empty(t, orgs)
}

func TestMultiTenant(t *testing.T) {
	// Start a mock server
	var orgs []string
	var orgMu sync.Mutex
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mark the server as ready
		if r.URL.Path == "/ready" {
			if _, err := w.Write([]byte("ready")); err != nil {
				t.Logf("writing 'ready' response failed: %v", err)
				t.Fail()
			}
			return
		}

		// Store the query
		orgMu.Lock()
		orgs = r.Header.Values("X-Scope-OrgID")
		orgMu.Unlock()

		// Send the response
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	// Configure and initialize the plugin
	plugin := &LogQL{
		URL:            ts.URL,
		Organizations:  []string{"CompanyA", "CompanyB"},
		InstantQueries: []InstantQuery{{query: query{Query: `{job="varlogs"}`}}},
		Log:            &testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	// Start the plugin and collect data
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()
	require.NoError(t, plugin.Gather(&acc))

	// Check the query including the parameters
	orgMu.Lock()
	defer orgMu.Unlock()
	require.Equal(t, []string{"CompanyA|CompanyB"}, orgs)
}

func TestCases(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Register the plugin
	inputs.Add("logql", func() telegraf.Input {
		return &LogQL{Timeout: config.Duration(5 * time.Second)}
	})

	// Prepare the influx parser for expectations
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}
		testcasePath := filepath.Join("testcases", f.Name())
		configFilename := filepath.Join(testcasePath, "telegraf.conf")
		expectedFilename := filepath.Join(testcasePath, "expected.out")
		expectedQueryFilename := filepath.Join(testcasePath, "expected.query")

		// Compare options
		options := []cmp.Option{
			testutil.IgnoreTime(),
			testutil.SortMetrics(),
		}

		t.Run(f.Name(), func(t *testing.T) {
			// Read the input data
			matches, err := filepath.Glob(filepath.Join(testcasePath, "response_*.json"))
			require.NoError(t, err)

			responses := make(map[string][]byte, len(matches))
			for _, fn := range matches {
				response, err := os.ReadFile(fn)
				require.NoError(t, err)

				endpoint := filepath.Base(fn)
				endpoint = strings.TrimSuffix(strings.TrimPrefix(endpoint, "response_"), ".json")
				responses["/loki/api/v1/"+endpoint] = response
			}

			// Read the expected output and query
			expected, err := testutil.ParseMetricsFromFile(expectedFilename, parser)
			require.NoError(t, err)
			expectedQuery, err := os.ReadFile(expectedQueryFilename)
			require.NoError(t, err)
			expectedURL, err := url.Parse(strings.TrimSpace(string(expectedQuery)))
			require.NoError(t, err)

			// Start a mock server
			var query string
			var queryMu sync.Mutex
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Mark the server as ready
				if r.URL.Path == "/ready" {
					if _, err := w.Write([]byte("ready")); err != nil {
						t.Logf("writing 'ready' response failed: %v", err)
						t.Fail()
					}
					return
				}

				// Store the query
				queryMu.Lock()
				query = r.URL.String()
				queryMu.Unlock()

				// Send the response
				response, found := responses[r.URL.Path]
				if !found {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				if _, err := w.Write(response); err != nil {
					t.Logf("writing 'ready' response failed: %v", err)
					t.Fail()
				}
			}))
			defer ts.Close()

			// Configure and initialize the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)
			plugin := cfg.Inputs[0].Input.(*LogQL)
			plugin.URL = ts.URL
			require.NoError(t, plugin.Init())

			// Start the plugin and collect data
			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()
			require.NoError(t, plugin.Gather(&acc))

			// Check the query including the parameters
			queryMu.Lock()
			actualQuery := query
			queryMu.Unlock()

			actualURL, err := url.Parse(actualQuery)
			require.NoError(t, err)

			// Wipe actual time values
			actualParams := actualURL.Query()
			for _, k := range []string{"time", "start", "end"} {
				if actualParams.Has(k) {
					actualParams.Set(k, "")
				}
			}

			require.Equal(t, expectedURL.Path, actualURL.Path, "query path differs")
			require.Equal(t, actualParams, expectedURL.Query(), "query parameters differ")

			// Check the received metrics
			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), options...)
		})
	}
}
