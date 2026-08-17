package rdap

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

// Make sure RDAP implements telegraf.Input
var _ telegraf.Input = &RDAP{}

func mustUnix(t *testing.T, value string) int64 {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, value)
	require.NoError(t, err)
	return ts.Unix()
}

const fullResponse = `{
  "objectClassName": "domain",
  "ldhName": "example.com",
  "status": ["client transfer prohibited", "client delete prohibited"],
  "secureDNS": {"delegationSigned": true},
  "events": [
    {"eventAction": "registration", "eventDate": "1995-08-14T04:00:00Z"},
    {"eventAction": "expiration", "eventDate": "2028-08-13T04:00:00Z"},
    {"eventAction": "last changed", "eventDate": "2023-08-14T07:01:44Z"}
  ],
  "nameservers": [
    {"objectClassName": "nameserver", "ldhName": "A.IANA-SERVERS.NET"},
    {"objectClassName": "nameserver", "ldhName": "b.iana-servers.net"}
  ],
  "entities": [
    {
      "objectClassName": "entity",
      "roles": ["registrar"],
      "vcardArray": ["vcard", [["version", {}, "text", "4.0"], ["fn", {}, "text", "Example Registrar LLC"]]]
    },
    {
      "objectClassName": "entity",
      "roles": ["registrant"],
      "vcardArray": ["vcard", [["version", {}, "text", "4.0"], ["fn", {}, "text", "Jane Doe"]]]
    }
  ]
}`

const minimalResponse = `{
  "objectClassName": "domain",
  "ldhName": "minimal.dev",
  "events": [
    {"eventAction": "expiration", "eventDate": "2030-01-02T00:00:00Z"}
  ]
}`

func newMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/domain/example.com", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/rdap+json")
		_, _ = w.Write([]byte(fullResponse))
	})
	mux.HandleFunc("/domain/minimal.dev", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/rdap+json")
		_, _ = w.Write([]byte(minimalResponse))
	})
	mux.HandleFunc("/domain/missing.com", func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})
	return httptest.NewServer(mux)
}

func TestInit(t *testing.T) {
	plugin := &RDAP{
		Domains: []string{"example.com"},
		Server:  "https://rdap.org",
		Timeout: config.Duration(5 * time.Second),
		Log:     testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
}

func TestInitFail(t *testing.T) {
	tests := []struct {
		name     string
		domains  []string
		server   string
		timeout  config.Duration
		expected string
	}{
		{
			name:     "missing domains",
			timeout:  config.Duration(5 * time.Second),
			expected: "no domains configured",
		},
		{
			name:     "invalid timeout",
			domains:  []string{"example.com"},
			timeout:  config.Duration(0),
			expected: "timeout has to be greater than zero",
		},
		{
			name:     "server missing scheme",
			domains:  []string{"example.com"},
			server:   "rdap.org",
			timeout:  config.Duration(5 * time.Second),
			expected: "not a valid URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &RDAP{
				Domains: tt.domains,
				Server:  tt.server,
				Timeout: tt.timeout,
				Log:     testutil.Logger{},
			}
			require.ErrorContains(t, plugin.Init(), tt.expected)
		})
	}
}

func TestGather(t *testing.T) {
	srv := newMockServer(t)
	defer srv.Close()

	tests := []struct {
		name     string
		domains  []string
		expected []telegraf.Metric
	}{
		{
			name:    "full response",
			domains: []string{"example.com"},
			expected: []telegraf.Metric{
				metric.New(
					"rdap",
					map[string]string{
						"domain": "example.com",
						"status": "client transfer prohibited,client delete prohibited",
					},
					map[string]interface{}{
						"creation_timestamp":   mustUnix(t, "1995-08-14T04:00:00Z"),
						"dnssec_enabled":       true,
						"expiration_timestamp": mustUnix(t, "2028-08-13T04:00:00Z"),
						"updated_timestamp":    mustUnix(t, "2023-08-14T07:01:44Z"),
						"registrar":            "Example Registrar LLC",
						"registrant":           "Jane Doe",
						"name_servers":         "a.iana-servers.net,b.iana-servers.net",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:    "minimal response",
			domains: []string{"minimal.dev"},
			expected: []telegraf.Metric{
				metric.New(
					"rdap",
					map[string]string{
						"domain": "minimal.dev",
						"status": "unknown",
					},
					map[string]interface{}{
						"creation_timestamp":   int64(0),
						"dnssec_enabled":       false,
						"expiration_timestamp": mustUnix(t, "2030-01-02T00:00:00Z"),
						"updated_timestamp":    int64(0),
						"registrar":            "not set",
						"registrant":           "not set",
						"name_servers":         "",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:    "missing domain emits error metric",
			domains: []string{"missing.com"},
			expected: []telegraf.Metric{
				metric.New(
					"rdap",
					map[string]string{
						"domain": "missing.com",
						"status": "not found",
					},
					map[string]interface{}{
						"error": "RDAP server returned 404, object does not exist.",
					},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &RDAP{
				Domains: tt.domains,
				Server:  srv.URL,
				Timeout: config.Duration(5 * time.Second),
				Log:     testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))
			require.Empty(t, acc.Errors)

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, tt.expected, actual,
				testutil.IgnoreTime(),
				testutil.IgnoreFields("expiry"),
			)
		})
	}
}

func TestGatherInvalidDomain(t *testing.T) {
	plugin := &RDAP{
		Domains: []string{"not a domain"},
		Timeout: config.Duration(5 * time.Second),
		Log:     testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Len(t, acc.Errors, 1)
	require.ErrorContains(t, acc.FirstError(), "invalid domain format")
}

func TestExpiryIsCalculated(t *testing.T) {
	srv := newMockServer(t)
	defer srv.Close()

	plugin := &RDAP{
		Domains: []string{"example.com"},
		Server:  srv.URL,
		Timeout: config.Duration(5 * time.Second),
		Log:     testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))
	require.Empty(t, acc.Errors)

	m, ok := acc.Get("rdap")
	require.True(t, ok)
	expiry, ok := m.Fields["expiry"].(int64)
	require.True(t, ok)
	// example.com expires in 2028, so from any realistic test clock the
	// remaining time is well over a year and positive.
	require.Greater(t, expiry, int64(0))
	require.False(t, strings.HasPrefix(m.Tags["status"], "unknown"))
}
