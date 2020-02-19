package powerdns_recursor

import (
	"net"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type statServer struct{}

var metrics = "all-outqueries\t3591637\nanswers-slow\t36451\nanswers0-1\t177297\nanswers1-10\t1209328\n" +
	"answers10-100\t1238786\nanswers100-1000\t402917\nauth-zone-queries\t4\nauth4-answers-slow\t44248\n" +
	"auth4-answers0-1\t59169\nauth4-answers1-10\t1747403\nauth4-answers10-100\t1315621\n" +
	"auth4-answers100-1000\t424683\nauth6-answers-slow\t0\nauth6-answers0-1\t0\nauth6-answers1-10\t0\n" +
	"auth6-answers10-100\t0\nauth6-answers100-1000\t0\ncache-entries\t295917\ncache-hits\t148630\n" +
	"cache-misses\t2916149\ncase-mismatches\t0\nchain-resends\t418602\nclient-parse-errors\t0\n" +
	"concurrent-queries\t0\ndlg-only-drops\t0\ndnssec-queries\t151536\ndnssec-result-bogus\t0\n" +
	"dnssec-result-indeterminate\t0\ndnssec-result-insecure\t0\ndnssec-result-nta\t0\n" +
	"dnssec-result-secure\t46\ndnssec-validations\t46\ndont-outqueries\t62\necs-queries\t0\n" +
	"ecs-responses\t0\nedns-ping-matches\t0\nedns-ping-mismatches\t0\nfailed-host-entries\t33\n" +
	"fd-usage\t32\nignored-packets\t0\nipv6-outqueries\t0\nipv6-questions\t0\nmalloc-bytes\t0\n" +
	"max-cache-entries\t1000000\nmax-mthread-stack\t33747\nmax-packetcache-entries\t500000\n" +
	"negcache-entries\t100070\nno-packet-error\t0\nnoedns-outqueries\t72409\nnoerror-answers\t25155259\n" +
	"noping-outqueries\t0\nnsset-invalidations\t2385\nnsspeeds-entries\t3571\nnxdomain-answers\t3307768\n" +
	"outgoing-timeouts\t43876\noutgoing4-timeouts\t43876\noutgoing6-timeouts\t0\nover-capacity-drops\t0\n" +
	"packetcache-entries\t80756\npacketcache-hits\t25698497\npacketcache-misses\t3064625\npolicy-drops\t0\n" +
	"policy-result-custom\t0\npolicy-result-drop\t0\npolicy-result-noaction\t3064779\npolicy-result-nodata\t0\n" +
	"policy-result-nxdomain\t0\npolicy-result-truncate\t0\nqa-latency\t6587\nquery-pipe-full-drops\t0\n" +
	"questions\t28763276\nreal-memory-usage\t280465408\nresource-limits\t0\nsecurity-status\t1\n" +
	"server-parse-errors\t0\nservfail-answers\t300249\nspoof-prevents\t0\nsys-msec\t1296588\n" +
	"tcp-client-overflow\t0\ntcp-clients\t0\ntcp-outqueries\t116\ntcp-questions\t130\nthrottle-entries\t33\n" +
	"throttled-out\t13187\nthrottled-outqueries\t13187\ntoo-old-drops\t2\nudp-in-errors\t4\n" +
	"udp-noport-errors\t2908\nudp-recvbuf-errors\t0\nudp-sndbuf-errors\t0\nunauthorized-tcp\t0\n" +
	"unauthorized-udp\t0\nunexpected-packets\t0\nunreachables\t1695\nuptime\t165725\nuser-msec\t1266384\n" +
	"x-our-latency\t19\nx-ourtime-slow\t632\nx-ourtime0-1\t3060079\nx-ourtime1-2\t3351\nx-ourtime16-32\t197\n" +
	"x-ourtime2-4\t302\nx-ourtime4-8\t194\nx-ourtime8-16\t24\n"

// first metric has no "\t"
var corruptMetrics = "all-outqueries3591637\nanswers-slow\t36451\nanswers0-1\t177297\nanswers1-10\t1209328\n" +
	"answers10-100\t1238786\nanswers100-1000\t402917\nauth-zone-queries\t4\nauth4-answers-slow\t44248\n" +
	"auth4-answers0-1\t59169\nauth4-answers1-10\t1747403\nauth4-answers10-100\t1315621\n" +
	"auth4-answers100-1000\t424683\nauth6-answers-slow\t0\nauth6-answers0-1\t0\nauth6-answers1-10\t0\n" +
	"auth6-answers10-100\t0\nauth6-answers100-1000\t0\ncache-entries\t295917\ncache-hits\t148630\n" +
	"cache-misses\t2916149\ncase-mismatches\t0\nchain-resends\t418602\nclient-parse-errors\t0\n" +
	"concurrent-queries\t0\ndlg-only-drops\t0\ndnssec-queries\t151536\ndnssec-result-bogus\t0\n" +
	"dnssec-result-indeterminate\t0\ndnssec-result-insecure\t0\ndnssec-result-nta\t0\n" +
	"dnssec-result-secure\t46\ndnssec-validations\t46\ndont-outqueries\t62\necs-queries\t0\n" +
	"ecs-responses\t0\nedns-ping-matches\t0\nedns-ping-mismatches\t0\nfailed-host-entries\t33\n" +
	"fd-usage\t32\nignored-packets\t0\nipv6-outqueries\t0\nipv6-questions\t0\nmalloc-bytes\t0\n" +
	"max-cache-entries\t1000000\nmax-mthread-stack\t33747\nmax-packetcache-entries\t500000\n" +
	"negcache-entries\t100070\nno-packet-error\t0\nnoedns-outqueries\t72409\nnoerror-answers\t25155259\n" +
	"noping-outqueries\t0\nnsset-invalidations\t2385\nnsspeeds-entries\t3571\nnxdomain-answers\t3307768\n" +
	"outgoing-timeouts\t43876\noutgoing4-timeouts\t43876\noutgoing6-timeouts\t0\nover-capacity-drops\t0\n" +
	"packetcache-entries\t80756\npacketcache-hits\t25698497\npacketcache-misses\t3064625\npolicy-drops\t0\n" +
	"policy-result-custom\t0\npolicy-result-drop\t0\npolicy-result-noaction\t3064779\npolicy-result-nodata\t0\n" +
	"policy-result-nxdomain\t0\npolicy-result-truncate\t0\nqa-latency\t6587\nquery-pipe-full-drops\t0\n" +
	"questions\t28763276\nreal-memory-usage\t280465408\nresource-limits\t0\nsecurity-status\t1\n" +
	"server-parse-errors\t0\nservfail-answers\t300249\nspoof-prevents\t0\nsys-msec\t1296588\n" +
	"tcp-client-overflow\t0\ntcp-clients\t0\ntcp-outqueries\t116\ntcp-questions\t130\nthrottle-entries\t33\n" +
	"throttled-out\t13187\nthrottled-outqueries\t13187\ntoo-old-drops\t2\nudp-in-errors\t4\n" +
	"udp-noport-errors\t2908\nudp-recvbuf-errors\t0\nudp-sndbuf-errors\t0\nunauthorized-tcp\t0\n" +
	"unauthorized-udp\t0\nunexpected-packets\t0\nunreachables\t1695\nuptime\t165725\nuser-msec\t1266384\n" +
	"x-our-latency\t19\nx-ourtime-slow\t632\nx-ourtime0-1\t3060079\nx-ourtime1-2\t3351\nx-ourtime16-32\t197\n" +
	"x-ourtime2-4\t302\nx-ourtime4-8\t194\nx-ourtime8-16\t24\n"

// integer overflow
var intOverflowMetrics = "all-outqueries\t18446744073709550195\nanswers-slow\t36451\nanswers0-1\t177297\nanswers1-10\t1209328\n" +
	"answers10-100\t1238786\nanswers100-1000\t402917\nauth-zone-queries\t4\nauth4-answers-slow\t44248\n" +
	"auth4-answers0-1\t59169\nauth4-answers1-10\t1747403\nauth4-answers10-100\t1315621\n" +
	"auth4-answers100-1000\t424683\nauth6-answers-slow\t0\nauth6-answers0-1\t0\nauth6-answers1-10\t0\n" +
	"auth6-answers10-100\t0\nauth6-answers100-1000\t0\ncache-entries\t295917\ncache-hits\t148630\n" +
	"cache-misses\t2916149\ncase-mismatches\t0\nchain-resends\t418602\nclient-parse-errors\t0\n" +
	"concurrent-queries\t0\ndlg-only-drops\t0\ndnssec-queries\t151536\ndnssec-result-bogus\t0\n" +
	"dnssec-result-indeterminate\t0\ndnssec-result-insecure\t0\ndnssec-result-nta\t0\n" +
	"dnssec-result-secure\t46\ndnssec-validations\t46\ndont-outqueries\t62\necs-queries\t0\n" +
	"ecs-responses\t0\nedns-ping-matches\t0\nedns-ping-mismatches\t0\nfailed-host-entries\t33\n" +
	"fd-usage\t32\nignored-packets\t0\nipv6-outqueries\t0\nipv6-questions\t0\nmalloc-bytes\t0\n" +
	"max-cache-entries\t1000000\nmax-mthread-stack\t33747\nmax-packetcache-entries\t500000\n" +
	"negcache-entries\t100070\nno-packet-error\t0\nnoedns-outqueries\t72409\nnoerror-answers\t25155259\n" +
	"noping-outqueries\t0\nnsset-invalidations\t2385\nnsspeeds-entries\t3571\nnxdomain-answers\t3307768\n" +
	"outgoing-timeouts\t43876\noutgoing4-timeouts\t43876\noutgoing6-timeouts\t0\nover-capacity-drops\t0\n" +
	"packetcache-entries\t80756\npacketcache-hits\t25698497\npacketcache-misses\t3064625\npolicy-drops\t0\n" +
	"policy-result-custom\t0\npolicy-result-drop\t0\npolicy-result-noaction\t3064779\npolicy-result-nodata\t0\n" +
	"policy-result-nxdomain\t0\npolicy-result-truncate\t0\nqa-latency\t6587\nquery-pipe-full-drops\t0\n" +
	"questions\t28763276\nreal-memory-usage\t280465408\nresource-limits\t0\nsecurity-status\t1\n" +
	"server-parse-errors\t0\nservfail-answers\t300249\nspoof-prevents\t0\nsys-msec\t1296588\n" +
	"tcp-client-overflow\t0\ntcp-clients\t0\ntcp-outqueries\t116\ntcp-questions\t130\nthrottle-entries\t33\n" +
	"throttled-out\t13187\nthrottled-outqueries\t13187\ntoo-old-drops\t2\nudp-in-errors\t4\n" +
	"udp-noport-errors\t2908\nudp-recvbuf-errors\t0\nudp-sndbuf-errors\t0\nunauthorized-tcp\t0\n" +
	"unauthorized-udp\t0\nunexpected-packets\t0\nunreachables\t1695\nuptime\t165725\nuser-msec\t1266384\n" +
	"x-our-latency\t19\nx-ourtime-slow\t632\nx-ourtime0-1\t3060079\nx-ourtime1-2\t3351\nx-ourtime16-32\t197\n" +
	"x-ourtime2-4\t302\nx-ourtime4-8\t194\nx-ourtime8-16\t24\n"

func TestPowerdnsRecursorGeneratesMetrics(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("Skipping test on darwin")
	}
	// We create a fake server to return test data
	controlSocket := "/tmp/pdns5724354148158589552.controlsocket"
	addr, err := net.ResolveUnixAddr("unixgram", controlSocket)
	if err != nil {
		t.Fatal("Cannot parse unix socket")
	}
	socket, err := net.ListenUnixgram("unixgram", addr)
	if err != nil {
		t.Fatal("Cannot initialize server on port")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() {
			socket.Close()
			os.Remove(controlSocket)
			wg.Done()
		}()

		for {
			buf := make([]byte, 1024)
			n, remote, err := socket.ReadFromUnix(buf)
			if err != nil {
				socket.Close()
				return
			}

			data := buf[:n]
			if string(data) == "get-all\n" {
				socket.WriteToUnix([]byte(metrics), remote)
				socket.Close()
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()

	p := &PowerdnsRecursor{
		UnixSockets: []string{controlSocket},
		SocketDir:   "/tmp",
		SocketMode:  "0666",
	}
	err = p.Init()
	require.NoError(t, err)

	var acc testutil.Accumulator

	err = acc.GatherError(p.Gather)
	require.NoError(t, err)

	wg.Wait()

	intMetrics := []string{"all-outqueries", "answers-slow", "answers0-1", "answers1-10",
		"answers10-100", "answers100-1000", "auth-zone-queries", "auth4-answers-slow",
		"auth4-answers0-1", "auth4-answers1-10", "auth4-answers10-100", "auth4-answers100-1000",
		"auth6-answers-slow", "auth6-answers0-1", "auth6-answers1-10", "auth6-answers10-100",
		"auth6-answers100-1000", "cache-entries", "cache-hits", "cache-misses", "case-mismatches",
		"chain-resends", "client-parse-errors", "concurrent-queries", "dlg-only-drops", "dnssec-queries",
		"dnssec-result-bogus", "dnssec-result-indeterminate", "dnssec-result-insecure", "dnssec-result-nta",
		"dnssec-result-secure", "dnssec-validations", "dont-outqueries", "ecs-queries", "ecs-responses",
		"edns-ping-matches", "edns-ping-mismatches", "failed-host-entries", "fd-usage", "ignored-packets",
		"ipv6-outqueries", "ipv6-questions", "malloc-bytes", "max-cache-entries", "max-mthread-stack",
		"max-packetcache-entries", "negcache-entries", "no-packet-error", "noedns-outqueries",
		"noerror-answers", "noping-outqueries", "nsset-invalidations", "nsspeeds-entries",
		"nxdomain-answers", "outgoing-timeouts", "outgoing4-timeouts", "outgoing6-timeouts",
		"over-capacity-drops", "packetcache-entries", "packetcache-hits", "packetcache-misses",
		"policy-drops", "policy-result-custom", "policy-result-drop", "policy-result-noaction",
		"policy-result-nodata", "policy-result-nxdomain", "policy-result-truncate", "qa-latency",
		"query-pipe-full-drops", "questions", "real-memory-usage", "resource-limits", "security-status",
		"server-parse-errors", "servfail-answers", "spoof-prevents", "sys-msec", "tcp-client-overflow",
		"tcp-clients", "tcp-outqueries", "tcp-questions", "throttle-entries", "throttled-out", "throttled-outqueries",
		"too-old-drops", "udp-in-errors", "udp-noport-errors", "udp-recvbuf-errors", "udp-sndbuf-errors",
		"unauthorized-tcp", "unauthorized-udp", "unexpected-packets", "unreachables", "uptime", "user-msec",
		"x-our-latency", "x-ourtime-slow", "x-ourtime0-1", "x-ourtime1-2", "x-ourtime16-32",
		"x-ourtime2-4", "x-ourtime4-8", "x-ourtime8-16"}

	for _, metric := range intMetrics {
		assert.True(t, acc.HasInt64Field("powerdns_recursor", metric), metric)
	}
}

func TestPowerdnsRecursorParseMetrics(t *testing.T) {
	values := parseResponse(metrics)

	tests := []struct {
		key   string
		value int64
	}{
		{"all-outqueries", 3591637},
		{"answers-slow", 36451},
		{"answers0-1", 177297},
		{"answers1-10", 1209328},
		{"answers10-100", 1238786},
		{"answers100-1000", 402917},
		{"auth-zone-queries", 4},
		{"auth4-answers-slow", 44248},
		{"auth4-answers0-1", 59169},
		{"auth4-answers1-10", 1747403},
		{"auth4-answers10-100", 1315621},
		{"auth4-answers100-1000", 424683},
		{"auth6-answers-slow", 0},
		{"auth6-answers0-1", 0},
		{"auth6-answers1-10", 0},
		{"auth6-answers10-100", 0},
		{"auth6-answers100-1000", 0},
		{"cache-entries", 295917},
		{"cache-hits", 148630},
		{"cache-misses", 2916149},
		{"case-mismatches", 0},
		{"chain-resends", 418602},
		{"client-parse-errors", 0},
		{"concurrent-queries", 0},
		{"dlg-only-drops", 0},
		{"dnssec-queries", 151536},
		{"dnssec-result-bogus", 0},
		{"dnssec-result-indeterminate", 0},
		{"dnssec-result-insecure", 0},
		{"dnssec-result-nta", 0},
		{"dnssec-result-secure", 46},
		{"dnssec-validations", 46},
		{"dont-outqueries", 62},
		{"ecs-queries", 0},
		{"ecs-responses", 0},
		{"edns-ping-matches", 0},
		{"edns-ping-mismatches", 0},
		{"failed-host-entries", 33},
		{"fd-usage", 32},
		{"ignored-packets", 0},
		{"ipv6-outqueries", 0},
		{"ipv6-questions", 0},
		{"malloc-bytes", 0},
		{"max-cache-entries", 1000000},
		{"max-mthread-stack", 33747},
		{"max-packetcache-entries", 500000},
		{"negcache-entries", 100070},
		{"no-packet-error", 0},
		{"noedns-outqueries", 72409},
		{"noerror-answers", 25155259},
		{"noping-outqueries", 0},
		{"nsset-invalidations", 2385},
		{"nsspeeds-entries", 3571},
		{"nxdomain-answers", 3307768},
		{"outgoing-timeouts", 43876},
		{"outgoing4-timeouts", 43876},
		{"outgoing6-timeouts", 0},
		{"over-capacity-drops", 0},
		{"packetcache-entries", 80756},
		{"packetcache-hits", 25698497},
		{"packetcache-misses", 3064625},
		{"policy-drops", 0},
		{"policy-result-custom", 0},
		{"policy-result-drop", 0},
		{"policy-result-noaction", 3064779},
		{"policy-result-nodata", 0},
		{"policy-result-nxdomain", 0},
		{"policy-result-truncate", 0},
		{"qa-latency", 6587},
		{"query-pipe-full-drops", 0},
		{"questions", 28763276},
		{"real-memory-usage", 280465408},
		{"resource-limits", 0},
		{"security-status", 1},
		{"server-parse-errors", 0},
		{"servfail-answers", 300249},
		{"spoof-prevents", 0},
		{"sys-msec", 1296588},
		{"tcp-client-overflow", 0},
		{"tcp-clients", 0},
		{"tcp-outqueries", 116},
		{"tcp-questions", 130},
		{"throttle-entries", 33},
		{"throttled-out", 13187},
		{"throttled-outqueries", 13187},
		{"too-old-drops", 2},
		{"udp-in-errors", 4},
		{"udp-noport-errors", 2908},
		{"udp-recvbuf-errors", 0},
		{"udp-sndbuf-errors", 0},
		{"unauthorized-tcp", 0},
		{"unauthorized-udp", 0},
		{"unexpected-packets", 0},
		{"unreachables", 1695},
		{"uptime", 165725},
		{"user-msec", 1266384},
		{"x-our-latency", 19},
		{"x-ourtime-slow", 632},
		{"x-ourtime0-1", 3060079},
		{"x-ourtime1-2", 3351},
		{"x-ourtime16-32", 197},
		{"x-ourtime2-4", 302},
		{"x-ourtime4-8", 194},
		{"x-ourtime8-16", 24},
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

func TestPowerdnsRecursorParseCorruptMetrics(t *testing.T) {
	values := parseResponse(corruptMetrics)

	tests := []struct {
		key   string
		value int64
	}{
		{"answers-slow", 36451},
		{"answers0-1", 177297},
		{"answers1-10", 1209328},
		{"answers10-100", 1238786},
		{"answers100-1000", 402917},
		{"auth-zone-queries", 4},
		{"auth4-answers-slow", 44248},
		{"auth4-answers0-1", 59169},
		{"auth4-answers1-10", 1747403},
		{"auth4-answers10-100", 1315621},
		{"auth4-answers100-1000", 424683},
		{"auth6-answers-slow", 0},
		{"auth6-answers0-1", 0},
		{"auth6-answers1-10", 0},
		{"auth6-answers10-100", 0},
		{"auth6-answers100-1000", 0},
		{"cache-entries", 295917},
		{"cache-hits", 148630},
		{"cache-misses", 2916149},
		{"case-mismatches", 0},
		{"chain-resends", 418602},
		{"client-parse-errors", 0},
		{"concurrent-queries", 0},
		{"dlg-only-drops", 0},
		{"dnssec-queries", 151536},
		{"dnssec-result-bogus", 0},
		{"dnssec-result-indeterminate", 0},
		{"dnssec-result-insecure", 0},
		{"dnssec-result-nta", 0},
		{"dnssec-result-secure", 46},
		{"dnssec-validations", 46},
		{"dont-outqueries", 62},
		{"ecs-queries", 0},
		{"ecs-responses", 0},
		{"edns-ping-matches", 0},
		{"edns-ping-mismatches", 0},
		{"failed-host-entries", 33},
		{"fd-usage", 32},
		{"ignored-packets", 0},
		{"ipv6-outqueries", 0},
		{"ipv6-questions", 0},
		{"malloc-bytes", 0},
		{"max-cache-entries", 1000000},
		{"max-mthread-stack", 33747},
		{"max-packetcache-entries", 500000},
		{"negcache-entries", 100070},
		{"no-packet-error", 0},
		{"noedns-outqueries", 72409},
		{"noerror-answers", 25155259},
		{"noping-outqueries", 0},
		{"nsset-invalidations", 2385},
		{"nsspeeds-entries", 3571},
		{"nxdomain-answers", 3307768},
		{"outgoing-timeouts", 43876},
		{"outgoing4-timeouts", 43876},
		{"outgoing6-timeouts", 0},
		{"over-capacity-drops", 0},
		{"packetcache-entries", 80756},
		{"packetcache-hits", 25698497},
		{"packetcache-misses", 3064625},
		{"policy-drops", 0},
		{"policy-result-custom", 0},
		{"policy-result-drop", 0},
		{"policy-result-noaction", 3064779},
		{"policy-result-nodata", 0},
		{"policy-result-nxdomain", 0},
		{"policy-result-truncate", 0},
		{"qa-latency", 6587},
		{"query-pipe-full-drops", 0},
		{"questions", 28763276},
		{"real-memory-usage", 280465408},
		{"resource-limits", 0},
		{"security-status", 1},
		{"server-parse-errors", 0},
		{"servfail-answers", 300249},
		{"spoof-prevents", 0},
		{"sys-msec", 1296588},
		{"tcp-client-overflow", 0},
		{"tcp-clients", 0},
		{"tcp-outqueries", 116},
		{"tcp-questions", 130},
		{"throttle-entries", 33},
		{"throttled-out", 13187},
		{"throttled-outqueries", 13187},
		{"too-old-drops", 2},
		{"udp-in-errors", 4},
		{"udp-noport-errors", 2908},
		{"udp-recvbuf-errors", 0},
		{"udp-sndbuf-errors", 0},
		{"unauthorized-tcp", 0},
		{"unauthorized-udp", 0},
		{"unexpected-packets", 0},
		{"unreachables", 1695},
		{"uptime", 165725},
		{"user-msec", 1266384},
		{"x-our-latency", 19},
		{"x-ourtime-slow", 632},
		{"x-ourtime0-1", 3060079},
		{"x-ourtime1-2", 3351},
		{"x-ourtime16-32", 197},
		{"x-ourtime2-4", 302},
		{"x-ourtime4-8", 194},
		{"x-ourtime8-16", 24},
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

func TestPowerdnsRecursorParseIntOverflowMetrics(t *testing.T) {
	values := parseResponse(intOverflowMetrics)

	tests := []struct {
		key   string
		value int64
	}{
		{"answers-slow", 36451},
		{"answers0-1", 177297},
		{"answers1-10", 1209328},
		{"answers10-100", 1238786},
		{"answers100-1000", 402917},
		{"auth-zone-queries", 4},
		{"auth4-answers-slow", 44248},
		{"auth4-answers0-1", 59169},
		{"auth4-answers1-10", 1747403},
		{"auth4-answers10-100", 1315621},
		{"auth4-answers100-1000", 424683},
		{"auth6-answers-slow", 0},
		{"auth6-answers0-1", 0},
		{"auth6-answers1-10", 0},
		{"auth6-answers10-100", 0},
		{"auth6-answers100-1000", 0},
		{"cache-entries", 295917},
		{"cache-hits", 148630},
		{"cache-misses", 2916149},
		{"case-mismatches", 0},
		{"chain-resends", 418602},
		{"client-parse-errors", 0},
		{"concurrent-queries", 0},
		{"dlg-only-drops", 0},
		{"dnssec-queries", 151536},
		{"dnssec-result-bogus", 0},
		{"dnssec-result-indeterminate", 0},
		{"dnssec-result-insecure", 0},
		{"dnssec-result-nta", 0},
		{"dnssec-result-secure", 46},
		{"dnssec-validations", 46},
		{"dont-outqueries", 62},
		{"ecs-queries", 0},
		{"ecs-responses", 0},
		{"edns-ping-matches", 0},
		{"edns-ping-mismatches", 0},
		{"failed-host-entries", 33},
		{"fd-usage", 32},
		{"ignored-packets", 0},
		{"ipv6-outqueries", 0},
		{"ipv6-questions", 0},
		{"malloc-bytes", 0},
		{"max-cache-entries", 1000000},
		{"max-mthread-stack", 33747},
		{"max-packetcache-entries", 500000},
		{"negcache-entries", 100070},
		{"no-packet-error", 0},
		{"noedns-outqueries", 72409},
		{"noerror-answers", 25155259},
		{"noping-outqueries", 0},
		{"nsset-invalidations", 2385},
		{"nsspeeds-entries", 3571},
		{"nxdomain-answers", 3307768},
		{"outgoing-timeouts", 43876},
		{"outgoing4-timeouts", 43876},
		{"outgoing6-timeouts", 0},
		{"over-capacity-drops", 0},
		{"packetcache-entries", 80756},
		{"packetcache-hits", 25698497},
		{"packetcache-misses", 3064625},
		{"policy-drops", 0},
		{"policy-result-custom", 0},
		{"policy-result-drop", 0},
		{"policy-result-noaction", 3064779},
		{"policy-result-nodata", 0},
		{"policy-result-nxdomain", 0},
		{"policy-result-truncate", 0},
		{"qa-latency", 6587},
		{"query-pipe-full-drops", 0},
		{"questions", 28763276},
		{"real-memory-usage", 280465408},
		{"resource-limits", 0},
		{"security-status", 1},
		{"server-parse-errors", 0},
		{"servfail-answers", 300249},
		{"spoof-prevents", 0},
		{"sys-msec", 1296588},
		{"tcp-client-overflow", 0},
		{"tcp-clients", 0},
		{"tcp-outqueries", 116},
		{"tcp-questions", 130},
		{"throttle-entries", 33},
		{"throttled-out", 13187},
		{"throttled-outqueries", 13187},
		{"too-old-drops", 2},
		{"udp-in-errors", 4},
		{"udp-noport-errors", 2908},
		{"udp-recvbuf-errors", 0},
		{"udp-sndbuf-errors", 0},
		{"unauthorized-tcp", 0},
		{"unauthorized-udp", 0},
		{"unexpected-packets", 0},
		{"unreachables", 1695},
		{"uptime", 165725},
		{"user-msec", 1266384},
		{"x-our-latency", 19},
		{"x-ourtime-slow", 632},
		{"x-ourtime0-1", 3060079},
		{"x-ourtime1-2", 3351},
		{"x-ourtime16-32", 197},
		{"x-ourtime2-4", 302},
		{"x-ourtime4-8", 194},
		{"x-ourtime8-16", 24},
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
