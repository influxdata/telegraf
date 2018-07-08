// +build !freebsd

package nats_streaming

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var sampleServerz = `
{
  "cluster_id": "test-cluster",
  "server_id": "KX1Y9BA1M7cPjLhZ7rldxm",
  "version": "0.10.0",
  "go": "go1.10.3",
  "state": "STANDALONE",
  "now": "2018-06-19T15:05:18.270880488-06:00",
  "start_time": "2018-06-19T15:04:48.730926175-06:00",
  "uptime": "29s",
  "clients": 20,
  "subscriptions": 10,
  "channels": 1,
  "total_msgs": 191474,
  "total_bytes": 28704590
}`

var sampleChannelsz = `
{
  "cluster_id": "test-cluster",
  "server_id": "J3Odi0wXYKWKFWz5D5uhH9",
  "now": "2017-06-07T15:01:02.166116959+02:00",
  "offset": 0,
  "limit": 1024,
  "count": 1,
  "total": 1,
  "channels": [
    {
      "name": "bar",
      "msgs": 123456,
      "bytes": 1234567890,
      "first_seq": 12340,
      "last_seq": 12345,
      "subscriptions": [
        {
          "client_id": "me",
          "inbox": "_INBOX.jAHSY3hcL5EGFQGYmfayvC",
          "ack_inbox": "_INBOX.J3Odi0wXYKWKFWz5D5uhem",
          "is_durable": false,
          "is_offline": false,
          "max_inflight": 1024,
          "ack_wait": 30,
          "last_sent": 704770,
          "pending_count": 0,
          "is_stalled": false
        }
      ]
    }
  ]
}
`

func TestMetricsCorrect(t *testing.T) {
	var acc testutil.Accumulator

	srv := newTestNatsServer()
	defer srv.Close()

	n := &Nats{Server: srv.URL}
	err := n.Gather(&acc)
	require.NoError(t, err)

	serverFields := map[string]interface{}{
		"clients":       int(20),
		"subscriptions": int(10),
		"channels":      int(1),
		"total_msgs":    int(191474),
		"total_bytes":   uint64(28704590),
		"uptime":        int64(29539954313),
	}
	serverTags := map[string]string{
		"server":     srv.URL,
		"cluster_id": "test-cluster",
		"server_id":  "KX1Y9BA1M7cPjLhZ7rldxm",
	}
	acc.AssertContainsTaggedFields(t, "nats_streaming_server", serverFields, serverTags)

	channelFields := map[string]interface{}{
		"msgs":      int(123456),
		"bytes":     uint64(1234567890),
		"first_seq": uint64(12340),
		"last_seq":  uint64(12345),
	}
	channelTags := map[string]string{
		"server":       srv.URL,
		"cluster_id":   "test-cluster",
		"server_id":    "J3Odi0wXYKWKFWz5D5uhH9",
		"channel_name": "bar",
	}
	acc.AssertContainsTaggedFields(t, "nats_streaming_channel", channelFields, channelTags)

	subFields := map[string]interface{}{
		"is_durable":    bool(false),
		"is_offline":    bool(false),
		"max_inflight":  int(1024),
		"ack_wait":      int(30),
		"last_sent":     uint64(704770),
		"pending_count": int(0),
		"is_stalled":    bool(false),
	}
	subTags := map[string]string{
		"server":       srv.URL,
		"cluster_id":   "test-cluster",
		"server_id":    "J3Odi0wXYKWKFWz5D5uhH9",
		"channel_name": "bar",
		"client_id":    "me",
		"inbox":        "_INBOX.jAHSY3hcL5EGFQGYmfayvC",
		"ack_inbox":    "_INBOX.J3Odi0wXYKWKFWz5D5uhem",
		"durable_name": "",
		"queue_name":   "",
	}
	acc.AssertContainsTaggedFields(t, "nats_streaming_subscription", subFields, subTags)
}

func newTestNatsServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rsp string

		switch r.URL.Path {
		case "/streaming/serverz":
			rsp = sampleServerz
		case "/streaming/channelsz":
			rsp = sampleChannelsz
		default:
			panic("Cannot handle request")
		}

		fmt.Fprintln(w, rsp)
	}))
}
