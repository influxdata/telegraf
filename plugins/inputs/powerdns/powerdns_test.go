package powerdns

import (
	"net"
	"testing"

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

func TestPowerdnsParseMetrics(t *testing.T) {
	p := &Powerdns{
		Log: testutil.Logger{},
	}

	values := p.parseResponse(metrics)

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

func TestPowerdnsParseCorruptMetrics(t *testing.T) {
	p := &Powerdns{
		Log: testutil.Logger{},
	}

	values := p.parseResponse(corruptMetrics)

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
	p := &Powerdns{
		Log: testutil.Logger{},
	}

	values := p.parseResponse(intOverflowMetrics)

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
