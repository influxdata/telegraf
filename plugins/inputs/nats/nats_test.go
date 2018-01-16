package nats

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sampleVarz = `
{
  "server_id": "n2afhLHLl64Gcaj7S7jaNa",
  "version": "1.0.0",
  "go": "go1.8",
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
  "http_port": 1337,
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
  "start": "1861-04-12T10:15:26.841483489-05:00",
  "now": "2011-10-05T15:24:23.722084098-07:00",
  "uptime": "150y5md237h8m57s",
  "mem": 15581184,
  "cores": 48,
  "cpu": 9,
  "connections": 2,
  "total_connections": 109,
  "routes": 0,
  "remotes": 0,
  "in_msgs": 74148556,
  "out_msgs": 68863261,
  "in_bytes": 946267004717,
  "out_bytes": 948110960598,
  "slow_consumers": 0,
  "subscriptions": 1,
  "http_req_stats": {
    "/": 1,
    "/connz": 100847,
    "/routez": 0,
    "/subsz": 1,
    "/varz": 205785
  },
  "config_load_time": "2017-07-24T10:15:26.841483489-05:00"
}
`

func TestMetricsCorrect(t *testing.T) {
	var acc testutil.Accumulator

	srv := newTestNatsServer()
	defer srv.Close()

	n := &Nats{Server: srv.URL}
	err := n.Gather(&acc)
	require.NoError(t, err)

	// we get the measurement, and override it, this is neccessary
	// because we can't "equal" the uptime value reliably, as it is
	// calculated via Time.Now() and the Start value in Varz
	s, f := acc.Get("nats_varz")
	assert.Equal(t, true, f, "nats measurement must be found")

	fields := make(map[string]interface{})
	fields["uptime"] = s.Fields["uptime"]
	fields["in_msgs"] = int64(74148556)
	fields["out_msgs"] = int64(68863261)
	fields["connections"] = int(2)
	fields["total_connections"] = uint64(109)
	fields["in_bytes"] = int64(946267004717)
	fields["out_bytes"] = int64(948110960598)
	fields["cpu_usage"] = float64(9)
	fields["mem"] = int64(15581184)
	fields["subscriptions"] = uint32(1)

	acc.AssertContainsFields(t, "nats_varz", fields)
}

func newTestNatsServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string

		switch r.URL.Path {
		case "/varz":
			rsp = sampleVarz
		default:
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
}
