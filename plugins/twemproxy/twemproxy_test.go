package twemproxy

import (
	"encoding/json"
	"net"
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleStatsAddr = "127.0.0.1:22222"

const sampleStats = `{
  "total_connections": 276448,
  "uptime": 160657,
  "version": "0.4.1",
  "service": "nutcracker",
  "curr_connections": 1322,
  "source": "server1.website.com",
  "demo": {
    "client_connections": 1305,
    "forward_error": 11684,
    "client_err": 147942,
    "server_ejects": 0,
    "fragments": 0,
    "client_eof": 126813,
    "10.16.29.1:6379": {
      "requests": 43604566,
      "server_eof": 0,
      "out_queue": 0,
      "server_err": 0,
      "out_queue_bytes": 0,
      "in_queue": 0,
      "server_timedout": 24,
      "request_bytes": 2775840400,
      "server_connections": 1,
      "response_bytes": 7663182096,
      "in_queue_bytes": 0,
      "server_ejected_at": 0,
      "responses": 43603900
    },
    "10.16.29.2:6379": {
      "requests": 37870211,
      "server_eof": 0,
      "out_queue": 0,
      "server_err": 0,
      "out_queue_bytes": 0,
      "in_queue": 0,
      "server_timedout": 25,
      "request_bytes": 2412114759,
      "server_connections": 1,
      "response_bytes": 5228980582,
      "in_queue_bytes": 0,
      "server_ejected_at": 0,
      "responses": 37869551
    }
  },
  "timestamp": 1447312436
}`

func mockTwemproxyServer() (net.Listener, error) {
	listener, err := net.Listen("tcp", sampleStatsAddr)
	if err != nil {
		return nil, err
	}
	go func(l net.Listener) {
		for {
			conn, _ := l.Accept()
			conn.Write([]byte(sampleStats))
			conn.Close()
			break
		}
	}(listener)

	return listener, nil
}

func TestGather(t *testing.T) {
	mockServer, err := mockTwemproxyServer()
	if err != nil {
		panic(err)
	}
	defer mockServer.Close()

	twemproxy := &Twemproxy{
		Instances: []TwemproxyInstance{
			TwemproxyInstance{
				StatsAddr: sampleStatsAddr,
				Pools:     []string{"demo"},
			},
		},
	}

	var acc testutil.Accumulator
	err = twemproxy.Instances[0].Gather(&acc)
	require.NoError(t, err)

	var sourceData map[string]interface{}
	if err := json.Unmarshal([]byte(sampleStats), &sourceData); err != nil {
		panic(err)
	}

	metrics := []string{"total_connections", "curr_connections", "timestamp"}
	tags := map[string]string{
		"twemproxy": sampleStatsAddr,
		"source":    sourceData["source"].(string),
	}
	for _, m := range metrics {
		assert.NoError(t, acc.ValidateTaggedValue(m, sourceData[m].(float64), tags))
	}

	poolName := "demo"
	poolMetrics := []string{
		"client_connections", "forward_error", "client_err", "server_ejects",
		"fragments", "client_eof",
	}
	tags["pool"] = poolName
	poolData := sourceData[poolName].(map[string]interface{})
	for _, m := range poolMetrics {
		measurement := poolName + "_" + m
		assert.NoError(t, acc.ValidateTaggedValue(measurement, poolData[m].(float64), tags))
	}
	poolServers := []string{"10.16.29.1:6379", "10.16.29.2:6379"}
	for _, s := range poolServers {
		tags["server"] = s
		serverData := poolData[s].(map[string]interface{})
		for k, v := range serverData {
			measurement := poolName + "_" + k
			assert.NoError(t, acc.ValidateTaggedValue(measurement, v, tags))
		}
	}
}
