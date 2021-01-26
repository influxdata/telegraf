package procstat

import (
	"net"
	"testing"
	"time"

	"github.com/elastic/gosigar/sys/linux"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestAddConnectionEndpoints(t *testing.T) {
	tests := []struct {
		name        string
		pid         PID
		ppid        PID
		listenPorts map[uint32]interface{}
		tcp         map[uint32][]ConnInfo
		publicIPs   []net.IP
		privateIPs  []net.IP
		metrics     []telegraf.Metric
		err         string
	}{
		{
			name: "no connections, no metrics",
		},
		{
			name:        "outside connection",
			pid:         100,
			listenPorts: map[uint32]interface{}{},
			tcp: map[uint32][]ConnInfo{
				100: {
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 34567,
						dstIP:   net.ParseIP("1.1.1.1"),
						dstPort: 80,
						state:   linux.TCP_ESTABLISHED,
					},
				},
			},
			publicIPs:  []net.IP{},
			privateIPs: []net.IP{},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					MetricNameTCPConnections,
					map[string]string{},
					map[string]interface{}{
						TCPConnectionKey: "1.1.1.1:80",
					},
					time.Now(),
				),
			},
		},
		{
			name:        "TCP states except SYN_SENT are used for connections",
			pid:         100,
			listenPorts: map[uint32]interface{}{},
			tcp: map[uint32][]ConnInfo{
				100: {
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 10000,
						dstIP:   net.ParseIP("1.1.1.1"),
						dstPort: 80,
						state:   linux.TCP_ESTABLISHED,
					},
					{ // this is ignore, is a host trying to connect but the other end has not replied
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 10001,
						dstIP:   net.ParseIP("1.1.1.1"),
						dstPort: 81,
						state:   linux.TCP_SYN_SENT,
					},
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 10002,
						dstIP:   net.ParseIP("1.1.1.1"),
						dstPort: 82,
						state:   linux.TCP_SYN_RECV,
					},
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 10003,
						dstIP:   net.ParseIP("1.1.1.1"),
						dstPort: 83,
						state:   linux.TCP_FIN_WAIT1,
					},
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 10004,
						dstIP:   net.ParseIP("1.1.1.1"),
						dstPort: 84,
						state:   linux.TCP_FIN_WAIT2,
					},
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 10005,
						dstIP:   net.ParseIP("1.1.1.1"),
						dstPort: 85,
						state:   linux.TCP_TIME_WAIT,
					},
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 10006,
						dstIP:   net.ParseIP("1.1.1.1"),
						dstPort: 86,
						state:   linux.TCP_CLOSE,
					},
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 10007,
						dstIP:   net.ParseIP("1.1.1.1"),
						dstPort: 87,
						state:   linux.TCP_CLOSE_WAIT,
					},
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 10008,
						dstIP:   net.ParseIP("1.1.1.1"),
						dstPort: 88,
						state:   linux.TCP_LAST_ACK,
					},
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 10009,
						dstIP:   net.ParseIP("1.1.1.1"),
						dstPort: 89,
						state:   linux.TCP_CLOSING,
					},
				},
			},
			publicIPs:  []net.IP{},
			privateIPs: []net.IP{},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					MetricNameTCPConnections,
					map[string]string{},
					map[string]interface{}{
						TCPConnectionKey: "1.1.1.1:80,1.1.1.1:82,1.1.1.1:83,1.1.1.1:84,1.1.1.1:85,1.1.1.1:86,1.1.1.1:87,1.1.1.1:88,1.1.1.1:89",
					},
					time.Now(),
				),
			},
		},
		{
			name:        "IPv4 listener",
			pid:         100,
			listenPorts: map[uint32]interface{}{80: nil},
			tcp: map[uint32][]ConnInfo{
				100: {
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 80,
						state:   linux.TCP_LISTEN,
					},
				},
			},
			publicIPs:  []net.IP{net.ParseIP("192.168.0.2")},
			privateIPs: []net.IP{},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					MetricNameTCPConnections,
					map[string]string{},
					map[string]interface{}{
						TCPListenKey: "192.168.0.2:80",
					},
					time.Now(),
				),
			},
		},
		{
			name:        "process listening in a IP not present in the local IPs will generate metric anyway",
			pid:         100,
			listenPorts: map[uint32]interface{}{80: nil},
			tcp: map[uint32][]ConnInfo{
				100: {
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 80,
						state:   linux.TCP_LISTEN,
					},
				},
			},
			publicIPs:  []net.IP{},
			privateIPs: []net.IP{},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					MetricNameTCPConnections,
					map[string]string{},
					map[string]interface{}{
						TCPListenKey: "192.168.0.2:80",
					},
					time.Now(),
				),
			},
		},
		{
			name:        "process listening in a port not present in the listeners list will generate metric anyway",
			pid:         100,
			listenPorts: map[uint32]interface{}{},
			tcp: map[uint32][]ConnInfo{
				100: {
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 80,
						state:   linux.TCP_LISTEN,
					},
				},
			},
			publicIPs:  []net.IP{},
			privateIPs: []net.IP{},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					MetricNameTCPConnections,
					map[string]string{},
					map[string]interface{}{
						TCPListenKey: "192.168.0.2:80",
					},
					time.Now(),
				),
			},
		},
		{
			name:        "IPv6 listener",
			pid:         100,
			listenPorts: map[uint32]interface{}{80: nil},
			tcp: map[uint32][]ConnInfo{
				100: {
					{
						srcIP:   net.ParseIP("dead::beef"),
						srcPort: 80,
						state:   linux.TCP_LISTEN,
					},
				},
			},
			publicIPs:  []net.IP{},
			privateIPs: []net.IP{},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					MetricNameTCPConnections,
					map[string]string{},
					map[string]interface{}{
						TCPListenKey: "[dead::beef]:80",
					},
					time.Now(),
				),
			},
		},
		{
			name: "private IPv4 listener do not generate metrics",
			pid:  100,
			listenPorts: map[uint32]interface{}{
				80: nil,
			},
			tcp: map[uint32][]ConnInfo{
				100: {
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 80,
						state:   linux.TCP_LISTEN,
					},
				},
			},
			publicIPs:  []net.IP{},
			privateIPs: []net.IP{net.ParseIP("192.168.0.2")},
			metrics:    []telegraf.Metric{},
		},
		{
			name:        "private IPv6 listener do not generate metrics",
			pid:         100,
			listenPorts: map[uint32]interface{}{80: nil},
			tcp: map[uint32][]ConnInfo{
				100: {
					{
						srcIP:   net.ParseIP("dead::beef"),
						srcPort: 80,
						state:   linux.TCP_LISTEN,
					},
				},
			},
			publicIPs:  []net.IP{},
			privateIPs: []net.IP{net.ParseIP("dead::beef")},
			metrics:    []telegraf.Metric{},
		},
		{
			name:        "0.0.0.0 listener listen in all public IPv4s",
			pid:         100,
			listenPorts: map[uint32]interface{}{80: nil},
			tcp: map[uint32][]ConnInfo{
				100: {
					{
						srcIP:   net.ParseIP("0.0.0.0"),
						srcPort: 80,
						state:   linux.TCP_LISTEN,
					},
				},
			},
			publicIPs:  []net.IP{net.ParseIP("192.168.0.2"), net.ParseIP("10.10.0.2"), net.ParseIP("dead::beef")},
			privateIPs: []net.IP{},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					MetricNameTCPConnections,
					map[string]string{},
					map[string]interface{}{
						TCPListenKey: "10.10.0.2:80,192.168.0.2:80",
					},
					time.Now(),
				),
			},
		},
		{
			name:        ":: listener listen in all public IPv4 and IPv6s",
			pid:         100,
			listenPorts: map[uint32]interface{}{80: nil},
			tcp: map[uint32][]ConnInfo{
				100: {
					{
						srcIP:   net.ParseIP("::"),
						srcPort: 80,
						state:   linux.TCP_LISTEN,
					},
				},
			},
			publicIPs:  []net.IP{net.ParseIP("192.168.0.2"), net.ParseIP("10.10.0.2"), net.ParseIP("dead::beef")},
			privateIPs: []net.IP{},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					MetricNameTCPConnections,
					map[string]string{},
					map[string]interface{}{
						TCPListenKey: "10.10.0.2:80,192.168.0.2:80,[dead::beef]:80",
					},
					time.Now(),
				),
			},
		},
		{
			name:        "ignore listeners in loopback IPs",
			pid:         100,
			listenPorts: map[uint32]interface{}{80: nil},
			tcp: map[uint32][]ConnInfo{
				100: {
					{
						srcIP:   net.ParseIP("127.0.0.1"),
						srcPort: 80,
						state:   linux.TCP_LISTEN,
					},
				},
			},
			publicIPs:  []net.IP{},
			privateIPs: []net.IP{},
			metrics:    []telegraf.Metric{},
		},
		{
			name:        "ignore connections from external hosts to local listeners",
			pid:         100,
			listenPorts: map[uint32]interface{}{80: nil},
			tcp: map[uint32][]ConnInfo{
				100: {
					{
						srcIP:   net.ParseIP("127.0.0.1"),
						srcPort: 80,
						dstIP:   net.ParseIP("54.89.89.54"),
						dstPort: 30123,
						state:   linux.TCP_ESTABLISHED,
					},
				},
			},
			publicIPs:  []net.IP{},
			privateIPs: []net.IP{},
			metrics:    []telegraf.Metric{},
		},
		{
			name:        "ignore connections from internal procs to other internal procs using the public IPs",
			pid:         100,
			listenPorts: map[uint32]interface{}{80: nil},
			tcp: map[uint32][]ConnInfo{
				100: {
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 30000,
						dstIP:   net.ParseIP("192.168.0.2"),
						dstPort: 80,
						state:   linux.TCP_ESTABLISHED,
					},
				},
			},
			publicIPs:  []net.IP{},
			privateIPs: []net.IP{net.ParseIP("192.168.0.2")},
			metrics:    []telegraf.Metric{},
		},
		{ // We are testing how behaves addConnectionEnpoints it if received a "pid not found" kind of error
			name:        "proc without network info does not generates an error, nor metrics",
			pid:         100,
			listenPorts: map[uint32]interface{}{},
			tcp:         map[uint32][]ConnInfo{},
			publicIPs:   []net.IP{},
			privateIPs:  []net.IP{},
			metrics:     []telegraf.Metric{},
		},
		{
			name: "process listening in two differents ports using :: with differents public IPs",
		},
		{ // same schema valid for: apache httpd, php-fpm
			name: "service with a parent process and several child, only the parent should report the listeners, parent case (nginx style)",
			pid:  101, // parent
			listenPorts: map[uint32]interface{}{
				80: nil,
			},
			tcp: map[uint32][]ConnInfo{
				100: { // parent
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 80,
						state:   linux.TCP_LISTEN,
					},
				},
				101: { // child
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 80,
						state:   linux.TCP_LISTEN,
					},
				},
			},
			publicIPs:  []net.IP{},
			privateIPs: []net.IP{},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					MetricNameTCPConnections,
					map[string]string{},
					map[string]interface{}{
						TCPListenKey: "192.168.0.2:80",
					},
					time.Now(),
				),
			},
		},
		{
			name: "service with a parent process and several child, only the parent should report the listeners, child case (nginx style)",
			pid:  101, // child
			ppid: 100,
			listenPorts: map[uint32]interface{}{
				80: nil,
			},
			tcp: map[uint32][]ConnInfo{
				100: { // parent
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 80,
						state:   linux.TCP_LISTEN,
					},
				},
				101: { // child
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 80,
						state:   linux.TCP_LISTEN,
					},
				},
			},
			publicIPs:  []net.IP{},
			privateIPs: []net.IP{},
			metrics:    []telegraf.Metric{},
		},
		{
			name: "child process listening in parent process plus other port, generate metric with the extra listener",
			pid:  101, // child
			ppid: 100,
			listenPorts: map[uint32]interface{}{
				80:  nil,
				443: nil,
			},
			tcp: map[uint32][]ConnInfo{
				100: { // parent
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 80,
						state:   linux.TCP_LISTEN,
					},
				},
				101: { // child
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 80,
						state:   linux.TCP_LISTEN,
					},
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 443,
						state:   linux.TCP_LISTEN,
					},
				},
			},
			publicIPs:  []net.IP{},
			privateIPs: []net.IP{},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					MetricNameTCPConnections,
					map[string]string{},
					map[string]interface{}{
						TCPListenKey: "192.168.0.2:443",
					},
					time.Now(),
				),
			},
		},
		{
			name:        "process listening in 0.0.0.0 and also in some IPv4 address, avoid duplication",
			pid:         100,
			listenPorts: map[uint32]interface{}{80: nil},
			tcp: map[uint32][]ConnInfo{
				100: {
					{
						srcIP:   net.ParseIP("0.0.0.0"),
						srcPort: 80,
						state:   linux.TCP_LISTEN,
					},
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 80,
						state:   linux.TCP_LISTEN,
					},
					{
						srcIP:   net.ParseIP("172.17.0.1"),
						srcPort: 80,
						state:   linux.TCP_LISTEN,
					},
				},
			},
			publicIPs:  []net.IP{net.ParseIP("192.168.0.2"), net.ParseIP("10.10.0.2"), net.ParseIP("dead::beef")},
			privateIPs: []net.IP{},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					MetricNameTCPConnections,
					map[string]string{},
					map[string]interface{}{
						TCPListenKey: "10.10.0.2:80,172.17.0.1:80,192.168.0.2:80",
					},
					time.Now(),
				),
			},
		},
		{
			name:        "avoid duplication in outboun connections",
			pid:         100,
			listenPorts: map[uint32]interface{}{},
			tcp: map[uint32][]ConnInfo{
				100: {
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 34567,
						dstIP:   net.ParseIP("1.1.1.1"),
						dstPort: 80,
						state:   linux.TCP_ESTABLISHED,
					},
					{
						srcIP:   net.ParseIP("192.168.0.2"),
						srcPort: 34568,
						dstIP:   net.ParseIP("1.1.1.1"),
						dstPort: 80,
						state:   linux.TCP_ESTABLISHED,
					},
				},
			},
			publicIPs:  []net.IP{},
			privateIPs: []net.IP{},
			metrics: []telegraf.Metric{
				testutil.MustMetric(
					MetricNameTCPConnections,
					map[string]string{},
					map[string]interface{}{
						TCPConnectionKey: "1.1.1.1:80",
					},
					time.Now(),
				),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var acc testutil.Accumulator

			proc := &testProc{
				pid:  test.pid,
				ppid: test.ppid,
			}

			netInfo := NetworkInfo{
				tcp:         test.tcp,
				listenPorts: test.listenPorts,
				publicIPs:   test.publicIPs,
				privateIPs:  test.privateIPs,
			}

			err := addConnectionEnpoints(&acc, proc, netInfo)
			if err != nil {
				assert.EqualError(t, err, test.err)
				//assert.FailNowf(t, "error calling addConnectionEnpoints", err.Error())
			}

			// Function has generated the same number of metrics defined in the test
			assert.Len(t, acc.GetTelegrafMetrics(), len(test.metrics))

			for _, m := range test.metrics {
				for _, value := range m.FieldList() {
					assert.Truef(
						t,
						acc.HasPoint(m.Name(), m.Tags(), value.Key, value.Value),
						"Missing point: %s,%v %s=%s\nMetrics: %v",
						m.Name(),
						m.Tags(),
						value.Key,
						value.Value,
						acc.GetTelegrafMetrics(),
					)
				}
			}
		})
	}
}
