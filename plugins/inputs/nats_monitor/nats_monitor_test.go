package natsmonitor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleVarzResponse = `
{
  "server_id": "zqPyAZVnfJZkzFu3pXnI9g",
  "version": "0.9.6",
  "go": "go1.7.4",
  "host": "0.0.0.0",
  "auth_required": false,
  "ssl_required": false,
  "tls_required": false,
  "tls_verify": false,
  "addr": "0.0.0.0",
  "max_connections": 65536,
  "ping_interval": 120000000000,
  "ping_max": 2,
  "http_host": "0.0.0.0",
  "http_port": 8222,
  "https_port": 0,
  "auth_timeout": 1,
  "max_control_line": 1024,
  "cluster": {
    "addr": "0.0.0.0",
    "cluster_port": 0,
    "auth_timeout": 1
  },
  "tls_timeout": 0.5,
  "port": 4222,
  "max_payload": 1048576,
  "start": "2017-02-15T12:16:57.119465627Z",
  "now": "2017-02-15T13:59:51.317179567Z",
  "uptime": "1h42m54s",
  "mem": 6758400,
  "cores": 1,
  "cpu": 0,
  "connections": 5,
  "total_connections": 1,
  "routes": 0,
  "remotes": 0,
  "in_msgs": 0,
  "out_msgs": 0,
  "in_bytes": 0,
  "out_bytes": 0,
  "slow_consumers": 0,
  "subscriptions": 2,
  "http_req_stats": {
    "/": 1,
    "/connz": 0,
    "/routez": 0,
    "/subsz": 0,
    "/varz": 56
  }
}
`

func TestMonitorClientGetEndpointUrl(t *testing.T) {

	t.Run("One", func(t *testing.T) {
		testCases := []struct {
			url  string
			path string
			want string
		}{
			{"http://localhost:8222", "/varz", "http://localhost:8222/varz"},
			{"http://localhost:8222/", "/varz", "http://localhost:8222/varz"},
			{"http://localhost:8222", "varz", "http://localhost:8222/varz"},
			{"http://192.168.0.1", "varz", "http://192.168.0.1/varz"},
		}
		for _, tc := range testCases {
			c, err := newTestMonitorClient(tc.url)
			require.NoError(t, err)
			ep := c.getEndpointUrl(tc.path)
			assert.Equal(t, tc.want, ep, "api(%s).endpoint(%s) = %s; want %s", tc.url, tc.path, ep, tc.want)
		}
	})

	t.Run("Many", func(t *testing.T) {
		url := "http://127.0.0.1:8222"
		c, err := newTestMonitorClient(url)
		require.NoError(t, err)

		testCases := []struct {
			path string
			want string
		}{
			{"/varz", "http://127.0.0.1:8222/varz"},
			{"/connz", "http://127.0.0.1:8222/connz"},
			{"/connz", "http://127.0.0.1:8222/connz"}, //not unique
			{"/varz", "http://127.0.0.1:8222/varz"},   //not unique
		}
		numUniqueEndpoints := 2

		for _, tc := range testCases {
			ep := c.getEndpointUrl(tc.path)
			assert.Equal(t, tc.want, ep, "api(%s).endpoint(%s) = %s; want %s", url, tc.path, ep, tc.want)
		}
		assert.Equal(t, numUniqueEndpoints, len(c.endpoints), "number unique endpoints")
	})
}

func BenchmarkMonitorClientGetEndpointUrl(b *testing.B) {
	c, _ := newTestMonitorClient("http://localhost:8222")
	for i := 0; i < b.N; i++ {
		c.getEndpointUrl("/varz")
	}
}

func TestMonitorClient(t *testing.T) {

	ts := newTestMonitorServer()
	defer ts.Close()

	c, err := newTestMonitorClient(ts.URL)
	require.NoError(t, err)

	fields, err := c.varz()
	require.NoError(t, err)

	assert := assert.New(t)
	if assert.NotNil(fields) {
		assert.Equal(int64(2), fields["subscriptions"], "subscriptions")
		assert.Equal(int64(5), fields["connections"], "connections")
	}
}

func TestNatsMonitorGeneratesMetrics(t *testing.T) {

	var acc testutil.Accumulator

	ts := newTestMonitorServer()
	defer ts.Close()

	n := &NatsMonitor{Urls: []string{ts.URL}}
	err := n.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{"url": ts.URL}

	fields := map[string]interface{}{
		"subscriptions": int64(2),
		"cpu":           int64(0),
		"mem":           int64(6758400),
		"connections":   int64(5),
		"in_msgs":       int64(0),
		"out_msgs":      int64(0),
		"in_bytes":      int64(0),
		"out_bytes":     int64(0),
	}
	acc.AssertContainsTaggedFields(t, "nats_varz", fields, tags)
}

func BenchmarkNatsMonitorGeneratesMetrics(b *testing.B) {

	var acc testutil.Accumulator

	ts := newTestMonitorServer()
	defer ts.Close()

	n := &NatsMonitor{Urls: []string{ts.URL}}

	b.Run("oneNode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			n.Gather(&acc)
		}
	})
}

func newTestMonitorClient(serverURL string) (*monitorClient, error) {
	addr, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}
	httpClient := &http.Client{}
	return NewMonitorClient(addr, httpClient), nil
}

func newTestMonitorServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string

		switch r.URL.Path {
		case "/varz":
			rsp = sampleVarzResponse
		default:
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
}
