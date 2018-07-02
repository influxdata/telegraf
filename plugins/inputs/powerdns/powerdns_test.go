package powerdns

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"net"
	"reflect"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

var respsizesMetrics = "20	45\n40	514434\n60	2345\n80	4356\n100	345\n150	0\n200	0\n400	0\n600	345\n800	0\n1000	0\n1200	0\n1400	0\n1600	0\n1800	0\n2000	91345\n2200	0\n2400	0\n2600	0\n2800	32\n3000	0\n3200	0\n3400	0\n3600	0\n3800	0\n4000	0\n4200	0\n4400	0\n4600	0\n4800	0\n5000	0\n5200	0\n5400	0\n5600	0\n5800	0\n6000	0\n6200	0\n6400	0\n6600	0\n6800	0\n7000	0\n7200	0\n7400	0\n7600	0\n7800	0\n8000	0\n8200	0\n8400	0\n8600	0\n8800	0\n9000	0\n9200	0\n9400	0\n9600	0\n9800	0\n10000	0\n10200	0\n10400	0\n10600	0\n10800	0\n11000	0\n11200	0\n11400	0\n11600	0\n11800	0\n12000	0\n12200	0\n12400	0\n12600	0\n12800	0\n13000	0\n13200	0\n13400	0\n13600	0\n13800	0\n14000	0\n14200	0\n14400	0\n14600	0\n14800	0\n15000	0\n15200	0\n15400	0\n15600	0\n15800	0\n16000	0\n16200	0\n16400	0\n16600	0\n16800	0\n17000	0\n17200	0\n17400	0\n17600	0\n17800	0\n18000	0\n18200	0\n18400	0\n18600	0\n18800	0\n19000	0\n19200	0\n19400	0\n19600	0\n19800	0\n20000	0\n20200	0\n20400	0\n20600	0\n20800	0\n21000	0\n21200	0\n21400	0\n21600	0\n21800	0\n22000	0\n22200	0\n22400	0\n22600	0\n22800	0\n23000	0\n23200	0\n23400	0\n23600	0\n23800	0\n24000	0\n24200	0\n24400	0\n24600	0\n24800	0\n25000	0\n25200	0\n25400	0\n25600	0\n25800	0\n26000	0\n26200	0\n26400	0\n26600	0\n26800	0\n27000	0\n27200	0\n27400	0\n27600	0\n27800	0\n28000	0\n28200	0\n28400	0\n28600	0\n28800	0\n29000	0\n29200	0\n29400	0\n29600	0\n29800	0\n30000	0\n30200	0\n30400	0\n30600	0\n30800	0\n31000	0\n31200	0\n31400	0\n31600	0\n31800	0\n32000	0\n32200	0\n32400	0\n32600	0\n32800	0\n33000	0\n33200	0\n33400	0\n33600	0\n33800	0\n34000	0\n34200	0\n34400	0\n34600	0\n34800	0\n35000	0\n35200	0\n35400	0\n35600	0\n35800	0\n36000	0\n36200	0\n36400	0\n36600	0\n36800	0\n37000	0\n37200	0\n37400	0\n37600	0\n37800	0\n38000	0\n38200	0\n38400	0\n38600	0\n38800	0\n39000	0\n39200	0\n39400	0\n39600	0\n39800	0\n40000	0\n40200	0\n40400	0\n40600	0\n40800	0\n41000	0\n41200	0\n41400	0\n41600	0\n41800	0\n42000	0\n42200	0\n42400	0\n42600	0\n42800	0\n43000	0\n43200	0\n43400	0\n43600	0\n43800	0\n44000	0\n44200	0\n44400	0\n44600	0\n44800	0\n45000	0\n45200	0\n45400	0\n45600	0\n45800	0\n46000	0\n46200	0\n46400	0\n46600	0\n46800	0\n47000	0\n47200	0\n47400	0\n47600	0\n47800	0\n48000	0\n48200	0\n48400	0\n48600	0\n48800	0\n49000	0\n49200	0\n49400	0\n49600	0\n49800	0\n50000	0\n50200	0\n50400	0\n50600	0\n50800	0\n51000	0\n51200	0\n51400	0\n51600	0\n51800	0\n52000	0\n52200	0\n52400	0\n52600	0\n52800	0\n53000	0\n53200	0\n53400	0\n53600	0\n53800	0\n54000	0\n54200	0\n54400	0\n54600	0\n54800	0\n55000	0\n55200	0\n55400	0\n55600	0\n55800	0\n56000	0\n56200	0\n56400	0\n56600	0\n56800	0\n57000	0\n57200	0\n57400	991\n57600	0\n57800	0\n58000	0\n58200	0\n58400	0\n58600	0\n58800	0\n59000	0\n59200	0\n59400	0\n59600	0\n59800	0\n60000	62\n60200	0\n60400	0\n60600	0\n60800	0\n61000	0\n61200	0\n61400	0\n61600	0\n61800	0\n62000	0\n62200	0\n62400	0\n62600	0\n62800	0\n63000	0\n63200	0\n63400	0\n63600	0\n63800	0\n64000	0\n64200	0\n64400	0\n64600	0\n64800	0\n65535	0"
var respsizesIntOverflowMetrics = "20	45\n40	514434\n60	18446744073709550195\n80	18446744073709550195"

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
			if string(data) == "show *\n" {
				c.Write([]byte(metrics))
				c.Close()
			} else if string(data) == "respsizes\n" {
				c.Write([]byte(respsizesMetrics))
				c.Close()
			}
		}(conn)
	}
}

func TestMemcachedGeneratesMetrics(t *testing.T) {
	// We create a fake server to return test data
	var randomNumber int64
	binary.Read(rand.Reader, binary.LittleEndian, &randomNumber)
	socket, err := net.Listen("unix", fmt.Sprintf("/tmp/pdns%d.controlsocket", randomNumber))
	if err != nil {
		t.Fatal("Cannot initialize server on port ")
	}

	defer socket.Close()

	s := statServer{}
	go s.serverSocket(socket)

	p := &Powerdns{
		UnixSockets: []string{fmt.Sprintf("/tmp/pdns%d.controlsocket", randomNumber)},
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

func TestPowerdnsparseGeneralStats(t *testing.T) {
	values, err := parseGeneralStats(metrics)
	if err != nil {
		t.Errorf("Function returned error: %v", err)
	}

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

func TestPowerdnsParseRespsizesResponse(t *testing.T) {
	actual, err := parseRespsizesResponse(respsizesMetrics)
	if err != nil {
		t.Errorf("Function returned error: %v", err)
	}
	expect := map[string]interface{}{
		"20":    int64(45),
		"40":    int64(514479),
		"60":    int64(516824),
		"80":    int64(521180),
		"100":   int64(521525),
		"150":   int64(521525),
		"200":   int64(521525),
		"400":   int64(521525),
		"600":   int64(521870),
		"800":   int64(521870),
		"1000":  int64(521870),
		"1200":  int64(521870),
		"1400":  int64(521870),
		"1600":  int64(521870),
		"1800":  int64(521870),
		"2000":  int64(613215),
		"2200":  int64(613215),
		"2400":  int64(613215),
		"2600":  int64(613215),
		"2800":  int64(613247),
		"3000":  int64(613247),
		"3200":  int64(613247),
		"3400":  int64(613247),
		"3600":  int64(613247),
		"3800":  int64(613247),
		"4000":  int64(613247),
		"4200":  int64(613247),
		"4400":  int64(613247),
		"4600":  int64(613247),
		"4800":  int64(613247),
		"5000":  int64(613247),
		"5200":  int64(613247),
		"5400":  int64(613247),
		"5600":  int64(613247),
		"5800":  int64(613247),
		"6000":  int64(613247),
		"6200":  int64(613247),
		"6400":  int64(613247),
		"6600":  int64(613247),
		"6800":  int64(613247),
		"7000":  int64(613247),
		"7200":  int64(613247),
		"7400":  int64(613247),
		"7600":  int64(613247),
		"7800":  int64(613247),
		"8000":  int64(613247),
		"8200":  int64(613247),
		"8400":  int64(613247),
		"8600":  int64(613247),
		"8800":  int64(613247),
		"9000":  int64(613247),
		"9200":  int64(613247),
		"9400":  int64(613247),
		"9600":  int64(613247),
		"9800":  int64(613247),
		"10000": int64(613247),
		"10200": int64(613247),
		"10400": int64(613247),
		"10600": int64(613247),
		"10800": int64(613247),
		"11000": int64(613247),
		"11200": int64(613247),
		"11400": int64(613247),
		"11600": int64(613247),
		"11800": int64(613247),
		"12000": int64(613247),
		"12200": int64(613247),
		"12400": int64(613247),
		"12600": int64(613247),
		"12800": int64(613247),
		"13000": int64(613247),
		"13200": int64(613247),
		"13400": int64(613247),
		"13600": int64(613247),
		"13800": int64(613247),
		"14000": int64(613247),
		"14200": int64(613247),
		"14400": int64(613247),
		"14600": int64(613247),
		"14800": int64(613247),
		"15000": int64(613247),
		"15200": int64(613247),
		"15400": int64(613247),
		"15600": int64(613247),
		"15800": int64(613247),
		"16000": int64(613247),
		"16200": int64(613247),
		"16400": int64(613247),
		"16600": int64(613247),
		"16800": int64(613247),
		"17000": int64(613247),
		"17200": int64(613247),
		"17400": int64(613247),
		"17600": int64(613247),
		"17800": int64(613247),
		"18000": int64(613247),
		"18200": int64(613247),
		"18400": int64(613247),
		"18600": int64(613247),
		"18800": int64(613247),
		"19000": int64(613247),
		"19200": int64(613247),
		"19400": int64(613247),
		"19600": int64(613247),
		"19800": int64(613247),
		"20000": int64(613247),
		"20200": int64(613247),
		"20400": int64(613247),
		"20600": int64(613247),
		"20800": int64(613247),
		"21000": int64(613247),
		"21200": int64(613247),
		"21400": int64(613247),
		"21600": int64(613247),
		"21800": int64(613247),
		"22000": int64(613247),
		"22200": int64(613247),
		"22400": int64(613247),
		"22600": int64(613247),
		"22800": int64(613247),
		"23000": int64(613247),
		"23200": int64(613247),
		"23400": int64(613247),
		"23600": int64(613247),
		"23800": int64(613247),
		"24000": int64(613247),
		"24200": int64(613247),
		"24400": int64(613247),
		"24600": int64(613247),
		"24800": int64(613247),
		"25000": int64(613247),
		"25200": int64(613247),
		"25400": int64(613247),
		"25600": int64(613247),
		"25800": int64(613247),
		"26000": int64(613247),
		"26200": int64(613247),
		"26400": int64(613247),
		"26600": int64(613247),
		"26800": int64(613247),
		"27000": int64(613247),
		"27200": int64(613247),
		"27400": int64(613247),
		"27600": int64(613247),
		"27800": int64(613247),
		"28000": int64(613247),
		"28200": int64(613247),
		"28400": int64(613247),
		"28600": int64(613247),
		"28800": int64(613247),
		"29000": int64(613247),
		"29200": int64(613247),
		"29400": int64(613247),
		"29600": int64(613247),
		"29800": int64(613247),
		"30000": int64(613247),
		"30200": int64(613247),
		"30400": int64(613247),
		"30600": int64(613247),
		"30800": int64(613247),
		"31000": int64(613247),
		"31200": int64(613247),
		"31400": int64(613247),
		"31600": int64(613247),
		"31800": int64(613247),
		"32000": int64(613247),
		"32200": int64(613247),
		"32400": int64(613247),
		"32600": int64(613247),
		"32800": int64(613247),
		"33000": int64(613247),
		"33200": int64(613247),
		"33400": int64(613247),
		"33600": int64(613247),
		"33800": int64(613247),
		"34000": int64(613247),
		"34200": int64(613247),
		"34400": int64(613247),
		"34600": int64(613247),
		"34800": int64(613247),
		"35000": int64(613247),
		"35200": int64(613247),
		"35400": int64(613247),
		"35600": int64(613247),
		"35800": int64(613247),
		"36000": int64(613247),
		"36200": int64(613247),
		"36400": int64(613247),
		"36600": int64(613247),
		"36800": int64(613247),
		"37000": int64(613247),
		"37200": int64(613247),
		"37400": int64(613247),
		"37600": int64(613247),
		"37800": int64(613247),
		"38000": int64(613247),
		"38200": int64(613247),
		"38400": int64(613247),
		"38600": int64(613247),
		"38800": int64(613247),
		"39000": int64(613247),
		"39200": int64(613247),
		"39400": int64(613247),
		"39600": int64(613247),
		"39800": int64(613247),
		"40000": int64(613247),
		"40200": int64(613247),
		"40400": int64(613247),
		"40600": int64(613247),
		"40800": int64(613247),
		"41000": int64(613247),
		"41200": int64(613247),
		"41400": int64(613247),
		"41600": int64(613247),
		"41800": int64(613247),
		"42000": int64(613247),
		"42200": int64(613247),
		"42400": int64(613247),
		"42600": int64(613247),
		"42800": int64(613247),
		"43000": int64(613247),
		"43200": int64(613247),
		"43400": int64(613247),
		"43600": int64(613247),
		"43800": int64(613247),
		"44000": int64(613247),
		"44200": int64(613247),
		"44400": int64(613247),
		"44600": int64(613247),
		"44800": int64(613247),
		"45000": int64(613247),
		"45200": int64(613247),
		"45400": int64(613247),
		"45600": int64(613247),
		"45800": int64(613247),
		"46000": int64(613247),
		"46200": int64(613247),
		"46400": int64(613247),
		"46600": int64(613247),
		"46800": int64(613247),
		"47000": int64(613247),
		"47200": int64(613247),
		"47400": int64(613247),
		"47600": int64(613247),
		"47800": int64(613247),
		"48000": int64(613247),
		"48200": int64(613247),
		"48400": int64(613247),
		"48600": int64(613247),
		"48800": int64(613247),
		"49000": int64(613247),
		"49200": int64(613247),
		"49400": int64(613247),
		"49600": int64(613247),
		"49800": int64(613247),
		"50000": int64(613247),
		"50200": int64(613247),
		"50400": int64(613247),
		"50600": int64(613247),
		"50800": int64(613247),
		"51000": int64(613247),
		"51200": int64(613247),
		"51400": int64(613247),
		"51600": int64(613247),
		"51800": int64(613247),
		"52000": int64(613247),
		"52200": int64(613247),
		"52400": int64(613247),
		"52600": int64(613247),
		"52800": int64(613247),
		"53000": int64(613247),
		"53200": int64(613247),
		"53400": int64(613247),
		"53600": int64(613247),
		"53800": int64(613247),
		"54000": int64(613247),
		"54200": int64(613247),
		"54400": int64(613247),
		"54600": int64(613247),
		"54800": int64(613247),
		"55000": int64(613247),
		"55200": int64(613247),
		"55400": int64(613247),
		"55600": int64(613247),
		"55800": int64(613247),
		"56000": int64(613247),
		"56200": int64(613247),
		"56400": int64(613247),
		"56600": int64(613247),
		"56800": int64(613247),
		"57000": int64(613247),
		"57200": int64(613247),
		"57400": int64(614238),
		"57600": int64(614238),
		"57800": int64(614238),
		"58000": int64(614238),
		"58200": int64(614238),
		"58400": int64(614238),
		"58600": int64(614238),
		"58800": int64(614238),
		"59000": int64(614238),
		"59200": int64(614238),
		"59400": int64(614238),
		"59600": int64(614238),
		"59800": int64(614238),
		"60000": int64(614300),
		"60200": int64(614300),
		"60400": int64(614300),
		"60600": int64(614300),
		"60800": int64(614300),
		"61000": int64(614300),
		"61200": int64(614300),
		"61400": int64(614300),
		"61600": int64(614300),
		"61800": int64(614300),
		"62000": int64(614300),
		"62200": int64(614300),
		"62400": int64(614300),
		"62600": int64(614300),
		"62800": int64(614300),
		"63000": int64(614300),
		"63200": int64(614300),
		"63400": int64(614300),
		"63600": int64(614300),
		"63800": int64(614300),
		"64000": int64(614300),
		"64200": int64(614300),
		"64400": int64(614300),
		"64600": int64(614300),
		"64800": int64(614300),
		"65535": int64(614300),
		"count": int64(614300),
		"sum":   int64(264691940),
	}
	if !reflect.DeepEqual(actual, expect) {
		t.Errorf("Latency histogram mismatch, actual output does not match expected output")
	}
}

func TestPowerdnsParseRespsizesIntOverflowResponse(t *testing.T) {
	values, err := parseRespsizesResponse(respsizesIntOverflowMetrics)
	if err == nil {
		t.Errorf("Expected integer overflow error, function returned no error")
	}
	if values != nil {
		t.Errorf("Values returned should be nil for integer overflow error")
	}
}

func TestPowerdnsParseCorruptMetrics(t *testing.T) {
	values, err := parseGeneralStats(corruptMetrics)
	if err != nil {
		t.Errorf("Function returned error: %v", err)
	}

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
	values, err := parseGeneralStats(intOverflowMetrics)
	if err == nil {
		t.Errorf("Expected integer overflow error, function returned no error")
	}
	if values != nil {
		t.Errorf("Values returned should be nil for integer overflow error")
	}
}
