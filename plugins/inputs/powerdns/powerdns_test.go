package powerdns

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

type statServer struct{}

var metrics = "corrupt-packets=0,deferred-cache-inserts=0,deferred-cache-lookup=0," +
	"dnsupdate-answers=0,dnsupdate-changes=0,dnsupdate-queries=0," +
	"dnsupdate-refused=0,packetcache-hit=0,packetcache-miss=1,packetcache-size=0," +
	"query-cache-hit=0,query-cache-miss=6,rd-queries=1,recursing-answers=0," +
	"recursing-questions=0,recursion-unanswered=0,security-status=3," +
	"servfail-packets=0,signatures=0,tcp-answers=0,tcp-queries=0," +
	"timedout-packets=0,udp-answers=1,udp-answers-bytes=50,udp-do-queries=0," +
	"udp-queries=0,udp4-answers=1,udp4-queries=1,udp6-answers=0,udp6-queries=0," +
	"key-cache-size=0,latency=26,meta-cache-size=0,qsize-q=0," +
	"signature-cache-size=0,sys-msec=2889,uptime=86317,user-msec=2167,"

// first metric has no "="
var corruptMetrics = "corrupt-packets--0,deferred-cache-inserts=0,deferred-cache-lookup=0," +
	"dnsupdate-answers=0,dnsupdate-changes=0,dnsupdate-queries=0," +
	"dnsupdate-refused=0,packetcache-hit=0,packetcache-miss=1,packetcache-size=0," +
	"query-cache-hit=0,query-cache-miss=6,rd-queries=1,recursing-answers=0," +
	"recursing-questions=0,recursion-unanswered=0,security-status=3," +
	"servfail-packets=0,signatures=0,tcp-answers=0,tcp-queries=0," +
	"timedout-packets=0,udp-answers=1,udp-answers-bytes=50,udp-do-queries=0," +
	"udp-queries=0,udp4-answers=1,udp4-queries=1,udp6-answers=0,udp6-queries=0," +
	"key-cache-size=0,latency=26,meta-cache-size=0,qsize-q=0," +
	"signature-cache-size=0,sys-msec=2889,uptime=86317,user-msec=2167,"

// integer overflow
var intOverflowMetrics = "corrupt-packets=18446744073709550195,deferred-cache-inserts=0,deferred-cache-lookup=0," +
	"dnsupdate-answers=0,dnsupdate-changes=0,dnsupdate-queries=0," +
	"dnsupdate-refused=0,packetcache-hit=0,packetcache-miss=1,packetcache-size=0," +
	"query-cache-hit=0,query-cache-miss=6,rd-queries=1,recursing-answers=0," +
	"recursing-questions=0,recursion-unanswered=0,security-status=3," +
	"servfail-packets=0,signatures=0,tcp-answers=0,tcp-queries=0," +
	"timedout-packets=0,udp-answers=1,udp-answers-bytes=50,udp-do-queries=0," +
	"udp-queries=0,udp4-answers=1,udp4-queries=1,udp6-answers=0,udp6-queries=0," +
	"key-cache-size=0,latency=26,meta-cache-size=0,qsize-q=0," +
	"signature-cache-size=0,sys-msec=2889,uptime=86317,user-msec=2167,"

func (s statServer) serverSocket(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}

		go func(c net.Conn) {
			buf := make([]byte, 1024)
			n, _ := c.Read(buf)

			data := buf[:n]
			if string(data) == "show * \n" {
				// Ignore the returned error as we need to close the socket anyway
				//nolint:errcheck,revive
				c.Write([]byte(metrics))
				// Ignore the returned error as we cannot do anything about it anyway
				//nolint:errcheck,revive
				c.Close()
			}
		}(conn)
	}
}

func TestPowerdnsGeneratesMetrics(t *testing.T) {
	// We create a fake server to return test data
	randomNumber := int64(5239846799706671610)
	sockname := filepath.Join(os.TempDir(), fmt.Sprintf("pdns%d.controlsocket", randomNumber))
	socket, err := net.Listen("unix", sockname)
	if err != nil {
		t.Fatal("Cannot initialize server on port ")
	}

	defer socket.Close()

	s := statServer{}
	go s.serverSocket(socket)

	p := &Powerdns{
		UnixSockets: []string{sockname},
	}

	var acc testutil.Accumulator
	err = acc.GatherError(p.Gather)
	require.NoError(t, err)

	intMetrics := []string{"corrupt-packets", "deferred-cache-inserts",
		"deferred-cache-lookup", "dnsupdate-answers", "dnsupdate-changes",
		"dnsupdate-queries", "dnsupdate-refused", "packetcache-hit",
		"packetcache-miss", "packetcache-size", "query-cache-hit", "query-cache-miss",
		"rd-queries", "recursing-answers", "recursing-questions",
		"recursion-unanswered", "security-status", "servfail-packets", "signatures",
		"tcp-answers", "tcp-queries", "timedout-packets", "udp-answers",
		"udp-answers-bytes", "udp-do-queries", "udp-queries", "udp4-answers",
		"udp4-queries", "udp6-answers", "udp6-queries", "key-cache-size", "latency",
		"meta-cache-size", "qsize-q", "signature-cache-size", "sys-msec", "uptime", "user-msec"}

	for _, metric := range intMetrics {
		assert.True(t, acc.HasInt64Field("powerdns", metric), metric)
	}
}

func TestPowerdnsParseMetrics(t *testing.T) {
	values := parseResponse(metrics, false)

	tests := []struct {
		key   string
		value int64
	}{
		{"corrupt-packets", 0},
		{"deferred-cache-inserts", 0},
		{"deferred-cache-lookup", 0},
		{"dnsupdate-answers", 0},
		{"dnsupdate-changes", 0},
		{"dnsupdate-queries", 0},
		{"dnsupdate-refused", 0},
		{"packetcache-hit", 0},
		{"packetcache-miss", 1},
		{"packetcache-size", 0},
		{"query-cache-hit", 0},
		{"query-cache-miss", 6},
		{"rd-queries", 1},
		{"recursing-answers", 0},
		{"recursing-questions", 0},
		{"recursion-unanswered", 0},
		{"security-status", 3},
		{"servfail-packets", 0},
		{"signatures", 0},
		{"tcp-answers", 0},
		{"tcp-queries", 0},
		{"timedout-packets", 0},
		{"udp-answers", 1},
		{"udp-answers-bytes", 50},
		{"udp-do-queries", 0},
		{"udp-queries", 0},
		{"udp4-answers", 1},
		{"udp4-queries", 1},
		{"udp6-answers", 0},
		{"udp6-queries", 0},
		{"key-cache-size", 0},
		{"latency", 26},
		{"meta-cache-size", 0},
		{"qsize-q", 0},
		{"signature-cache-size", 0},
		{"sys-msec", 2889},
		{"uptime", 86317},
		{"user-msec", 2167},
	}

	for _, test := range tests {
		value, ok := values[test.key]
		if !ok {
			t.Errorf("Did not find key for metric %s in values", test.key)
			continue
		}
		if value != test.value {
			t.Errorf("Metric: %s, Expected: %d, actual: %d",
				test.key, test.value, value)
		}
	}
}

func TestPowerdnsParseMetricsUnderscore(t *testing.T) {
	values := parseResponse(metrics, true)

	tests := []struct {
		key   string
		value int64
	}{
		{"corrupt_packets", 0},
		{"deferred_cache_inserts", 0},
		{"deferred_cache_lookup", 0},
		{"dnsupdate_answers", 0},
		{"dnsupdate_changes", 0},
		{"dnsupdate_queries", 0},
		{"dnsupdate_refused", 0},
		{"packetcache_hit", 0},
		{"packetcache_miss", 1},
		{"packetcache_size", 0},
		{"query_cache_hit", 0},
		{"query_cache_miss", 6},
		{"rd_queries", 1},
		{"recursing_answers", 0},
		{"recursing_questions", 0},
		{"recursion_unanswered", 0},
		{"security_status", 3},
		{"servfail_packets", 0},
		{"signatures", 0},
		{"tcp_answers", 0},
		{"tcp_queries", 0},
		{"timedout_packets", 0},
		{"udp_answers", 1},
		{"udp_answers_bytes", 50},
		{"udp_do_queries", 0},
		{"udp_queries", 0},
		{"udp4_answers", 1},
		{"udp4_queries", 1},
		{"udp6_answers", 0},
		{"udp6_queries", 0},
		{"key_cache_size", 0},
		{"latency", 26},
		{"meta_cache_size", 0},
		{"qsize_q", 0},
		{"signature_cache_size", 0},
		{"sys_msec", 2889},
		{"uptime", 86317},
		{"user_msec", 2167},
	}

	for _, test := range tests {
		value, ok := values[test.key]
		if !ok {
			t.Errorf("Did not find key for metric %s in values", test.key)
			continue
		}
		if value != test.value {
			t.Errorf("Metric: %s, Expected: %d, actual: %d",
				test.key, test.value, value)
		}
	}
}

func TestPowerdnsParseCorruptMetrics(t *testing.T) {
	values := parseResponse(corruptMetrics, false)

	tests := []struct {
		key   string
		value int64
	}{
		{"deferred-cache-inserts", 0},
		{"deferred-cache-lookup", 0},
		{"dnsupdate-answers", 0},
		{"dnsupdate-changes", 0},
		{"dnsupdate-queries", 0},
		{"dnsupdate-refused", 0},
		{"packetcache-hit", 0},
		{"packetcache-miss", 1},
		{"packetcache-size", 0},
		{"query-cache-hit", 0},
		{"query-cache-miss", 6},
		{"rd-queries", 1},
		{"recursing-answers", 0},
		{"recursing-questions", 0},
		{"recursion-unanswered", 0},
		{"security-status", 3},
		{"servfail-packets", 0},
		{"signatures", 0},
		{"tcp-answers", 0},
		{"tcp-queries", 0},
		{"timedout-packets", 0},
		{"udp-answers", 1},
		{"udp-answers-bytes", 50},
		{"udp-do-queries", 0},
		{"udp-queries", 0},
		{"udp4-answers", 1},
		{"udp4-queries", 1},
		{"udp6-answers", 0},
		{"udp6-queries", 0},
		{"key-cache-size", 0},
		{"latency", 26},
		{"meta-cache-size", 0},
		{"qsize-q", 0},
		{"signature-cache-size", 0},
		{"sys-msec", 2889},
		{"uptime", 86317},
		{"user-msec", 2167},
	}

	for _, test := range tests {
		value, ok := values[test.key]
		if !ok {
			t.Errorf("Did not find key for metric %s in values", test.key)
			continue
		}
		if value != test.value {
			t.Errorf("Metric: %s, Expected: %d, actual: %d",
				test.key, test.value, value)
		}
	}
}

func TestPowerdnsParseIntOverflowMetrics(t *testing.T) {
	values := parseResponse(intOverflowMetrics, false)

	tests := []struct {
		key   string
		value int64
	}{
		{"deferred-cache-inserts", 0},
		{"deferred-cache-lookup", 0},
		{"dnsupdate-answers", 0},
		{"dnsupdate-changes", 0},
		{"dnsupdate-queries", 0},
		{"dnsupdate-refused", 0},
		{"packetcache-hit", 0},
		{"packetcache-miss", 1},
		{"packetcache-size", 0},
		{"query-cache-hit", 0},
		{"query-cache-miss", 6},
		{"rd-queries", 1},
		{"recursing-answers", 0},
		{"recursing-questions", 0},
		{"recursion-unanswered", 0},
		{"security-status", 3},
		{"servfail-packets", 0},
		{"signatures", 0},
		{"tcp-answers", 0},
		{"tcp-queries", 0},
		{"timedout-packets", 0},
		{"udp-answers", 1},
		{"udp-answers-bytes", 50},
		{"udp-do-queries", 0},
		{"udp-queries", 0},
		{"udp4-answers", 1},
		{"udp4-queries", 1},
		{"udp6-answers", 0},
		{"udp6-queries", 0},
		{"key-cache-size", 0},
		{"latency", 26},
		{"meta-cache-size", 0},
		{"qsize-q", 0},
		{"signature-cache-size", 0},
		{"sys-msec", 2889},
		{"uptime", 86317},
		{"user-msec", 2167},
	}

	for _, test := range tests {
		value, ok := values[test.key]
		if !ok {
			t.Errorf("Did not find key for metric %s in values", test.key)
			continue
		}
		if value != test.value {
			t.Errorf("Metric: %s, Expected: %d, actual: %d",
				test.key, test.value, value)
		}
	}
}
