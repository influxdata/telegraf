package twemproxy

import (
	"encoding/json"
	"net"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const sampleAddr = "127.0.0.1:22222"

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
	listener, err := net.Listen("tcp", sampleAddr)
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
		Addr:  sampleAddr,
		Pools: []string{"demo"},
	}

	var acc testutil.Accumulator
	acc.SetDebug(true)
	err = twemproxy.Gather(&acc)
	require.NoError(t, err)

	var sourceData map[string]interface{}
	if err := json.Unmarshal([]byte(sampleStats), &sourceData); err != nil {
		panic(err)
	}

	fields := map[string]interface{}{
		"total_connections": float64(276448),
		"curr_connections":  float64(1322),
		"timestamp":         float64(1.447312436e+09),
	}
	tags := map[string]string{
		"twemproxy": sampleAddr,
		"source":    sourceData["source"].(string),
	}
	acc.AssertContainsTaggedFields(t, "twemproxy", fields, tags)

	poolName := "demo"
	poolFields := map[string]interface{}{
		"client_connections": float64(1305),
		"client_eof":         float64(126813),
		"client_err":         float64(147942),
		"forward_error":      float64(11684),
		"fragments":          float64(0),
		"server_ejects":      float64(0),
	}
	tags["pool"] = poolName
	acc.AssertContainsTaggedFields(t, "twemproxy_pool", poolFields, tags)

	poolServerTags1 := map[string]string{
		"pool":      "demo",
		"server":    "10.16.29.2:6379",
		"source":    "server1.website.com",
		"twemproxy": "127.0.0.1:22222",
	}
	poolServerFields1 := map[string]interface{}{
		"in_queue":           float64(0),
		"in_queue_bytes":     float64(0),
		"out_queue":          float64(0),
		"out_queue_bytes":    float64(0),
		"request_bytes":      float64(2.412114759e+09),
		"requests":           float64(3.7870211e+07),
		"response_bytes":     float64(5.228980582e+09),
		"responses":          float64(3.7869551e+07),
		"server_connections": float64(1),
		"server_ejected_at":  float64(0),
		"server_eof":         float64(0),
		"server_err":         float64(0),
		"server_timedout":    float64(25),
	}
	acc.AssertContainsTaggedFields(t, "twemproxy_pool_server",
		poolServerFields1, poolServerTags1)

	poolServerTags2 := map[string]string{
		"pool":      "demo",
		"server":    "10.16.29.1:6379",
		"source":    "server1.website.com",
		"twemproxy": "127.0.0.1:22222",
	}
	poolServerFields2 := map[string]interface{}{
		"in_queue":           float64(0),
		"in_queue_bytes":     float64(0),
		"out_queue":          float64(0),
		"out_queue_bytes":    float64(0),
		"request_bytes":      float64(2.7758404e+09),
		"requests":           float64(4.3604566e+07),
		"response_bytes":     float64(7.663182096e+09),
		"responses":          float64(4.36039e+07),
		"server_connections": float64(1),
		"server_ejected_at":  float64(0),
		"server_eof":         float64(0),
		"server_err":         float64(0),
		"server_timedout":    float64(24),
	}
	acc.AssertContainsTaggedFields(t, "twemproxy_pool_server",
		poolServerFields2, poolServerTags2)
}
