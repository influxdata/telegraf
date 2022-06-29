//go:build !windows
// +build !windows

package ping

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/go-ping/ping"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/testutil"
)

// BSD/Darwin ping output
var bsdPingOutput = `
PING www.google.com (216.58.217.36): 56 data bytes
64 bytes from 216.58.217.36: icmp_seq=0 ttl=55 time=15.087 ms
64 bytes from 216.58.217.36: icmp_seq=1 ttl=55 time=21.564 ms
64 bytes from 216.58.217.36: icmp_seq=2 ttl=55 time=27.263 ms
64 bytes from 216.58.217.36: icmp_seq=3 ttl=55 time=18.828 ms
64 bytes from 216.58.217.36: icmp_seq=4 ttl=55 time=18.378 ms

--- www.google.com ping statistics ---
5 packets transmitted, 5 packets received, 0.0% packet loss
round-trip min/avg/max/stddev = 15.087/20.224/27.263/4.076 ms
`

// FreeBSD ping6 output
var freebsdPing6Output = `
PING6(64=40+8+16 bytes) 2001:db8::1 --> 2a00:1450:4001:824::2004
24 bytes from 2a00:1450:4001:824::2004, icmp_seq=0 hlim=117 time=93.870 ms
24 bytes from 2a00:1450:4001:824::2004, icmp_seq=1 hlim=117 time=40.278 ms
24 bytes from 2a00:1450:4001:824::2004, icmp_seq=2 hlim=120 time=59.077 ms
24 bytes from 2a00:1450:4001:824::2004, icmp_seq=3 hlim=117 time=37.102 ms
24 bytes from 2a00:1450:4001:824::2004, icmp_seq=4 hlim=117 time=35.727 ms

--- www.google.com ping6 statistics ---
5 packets transmitted, 5 packets received, 0.0% packet loss
round-trip min/avg/max/std-dev = 35.727/53.211/93.870/22.000 ms
`

// Linux ping output
var linuxPingOutput = `
PING www.google.com (216.58.218.164) 56(84) bytes of data.
64 bytes from host.net (216.58.218.164): icmp_seq=1 ttl=63 time=35.2 ms
64 bytes from host.net (216.58.218.164): icmp_seq=2 ttl=63 time=42.3 ms
64 bytes from host.net (216.58.218.164): icmp_seq=3 ttl=63 time=45.1 ms
64 bytes from host.net (216.58.218.164): icmp_seq=4 ttl=63 time=43.5 ms
64 bytes from host.net (216.58.218.164): icmp_seq=5 ttl=63 time=51.8 ms

--- www.google.com ping statistics ---
5 packets transmitted, 5 received, 0% packet loss, time 4010ms
rtt min/avg/max/mdev = 35.225/43.628/51.806/5.325 ms
`

// BusyBox v1.24.1 (2017-02-28 03:28:13 CET) multi-call binary
var busyBoxPingOutput = `
PING 8.8.8.8 (8.8.8.8): 56 data bytes
64 bytes from 8.8.8.8: seq=0 ttl=56 time=22.559 ms
64 bytes from 8.8.8.8: seq=1 ttl=56 time=15.810 ms
64 bytes from 8.8.8.8: seq=2 ttl=56 time=16.262 ms
64 bytes from 8.8.8.8: seq=3 ttl=56 time=15.815 ms

--- 8.8.8.8 ping statistics ---
4 packets transmitted, 4 packets received, 0% packet loss
round-trip min/avg/max = 15.810/17.611/22.559 ms
`

// Fatal ping output (invalid argument)
var fatalPingOutput = `
ping: -i interval too short: Operation not permitted
`

// Test that ping command output is processed properly
func TestProcessPingOutput(t *testing.T) {
	stats, err := processPingOutput(bsdPingOutput)
	require.NoError(t, err)
	require.Equal(t, 55, stats.ttl, "ttl value is 55")
	require.Equal(t, 5, stats.trans, "5 packets were transmitted")
	require.Equal(t, 5, stats.recv, "5 packets were received")
	require.InDelta(t, 15.087, stats.min, 0.001)
	require.InDelta(t, 20.224, stats.avg, 0.001)
	require.InDelta(t, 27.263, stats.max, 0.001)
	require.InDelta(t, 4.076, stats.stddev, 0.001)

	stats, err = processPingOutput(freebsdPing6Output)
	require.NoError(t, err)
	require.Equal(t, 117, stats.ttl, "ttl value is 117")
	require.Equal(t, 5, stats.trans, "5 packets were transmitted")
	require.Equal(t, 5, stats.recv, "5 packets were received")
	require.InDelta(t, 35.727, stats.min, 0.001)
	require.InDelta(t, 53.211, stats.avg, 0.001)
	require.InDelta(t, 93.870, stats.max, 0.001)
	require.InDelta(t, 22.000, stats.stddev, 0.001)

	stats, err = processPingOutput(linuxPingOutput)
	require.NoError(t, err)
	require.Equal(t, 63, stats.ttl, "ttl value is 63")
	require.Equal(t, 5, stats.trans, "5 packets were transmitted")
	require.Equal(t, 5, stats.recv, "5 packets were received")
	require.InDelta(t, 35.225, stats.min, 0.001)
	require.InDelta(t, 43.628, stats.avg, 0.001)
	require.InDelta(t, 51.806, stats.max, 0.001)
	require.InDelta(t, 5.325, stats.stddev, 0.001)

	stats, err = processPingOutput(busyBoxPingOutput)
	require.NoError(t, err)
	require.Equal(t, 56, stats.ttl, "ttl value is 56")
	require.Equal(t, 4, stats.trans, "4 packets were transmitted")
	require.Equal(t, 4, stats.recv, "4 packets were received")
	require.InDelta(t, 15.810, stats.min, 0.001)
	require.InDelta(t, 17.611, stats.avg, 0.001)
	require.InDelta(t, 22.559, stats.max, 0.001)
	require.InDelta(t, -1.0, stats.stddev, 0.001)
}

// Linux ping output with varying TTL
var linuxPingOutputWithVaryingTTL = `
PING www.google.com (216.58.218.164) 56(84) bytes of data.
64 bytes from host.net (216.58.218.164): icmp_seq=1 ttl=63 time=35.2 ms
64 bytes from host.net (216.58.218.164): icmp_seq=2 ttl=255 time=42.3 ms
64 bytes from host.net (216.58.218.164): icmp_seq=3 ttl=64 time=45.1 ms
64 bytes from host.net (216.58.218.164): icmp_seq=4 ttl=64 time=43.5 ms
64 bytes from host.net (216.58.218.164): icmp_seq=5 ttl=255 time=51.8 ms

--- www.google.com ping statistics ---
5 packets transmitted, 5 received, 0% packet loss, time 4010ms
rtt min/avg/max/mdev = 35.225/43.628/51.806/5.325 ms
`

// Test that ping command output is processed properly
func TestProcessPingOutputWithVaryingTTL(t *testing.T) {
	stats, err := processPingOutput(linuxPingOutputWithVaryingTTL)
	require.NoError(t, err)
	require.Equal(t, 63, stats.ttl, "ttl value is 63")
	require.Equal(t, 5, stats.trans, "5 packets were transmitted")
	require.Equal(t, 5, stats.recv, "5 packets were transmitted")
	require.InDelta(t, 35.225, stats.min, 0.001)
	require.InDelta(t, 43.628, stats.avg, 0.001)
	require.InDelta(t, 51.806, stats.max, 0.001)
	require.InDelta(t, 5.325, stats.stddev, 0.001)
}

// Test that processPingOutput returns an error when 'ping' fails to run, such
// as when an invalid argument is provided
func TestErrorProcessPingOutput(t *testing.T) {
	_, err := processPingOutput(fatalPingOutput)
	require.Error(t, err, "Error was expected from processPingOutput")
}

// Test that default arg lists are created correctly
func TestArgs(t *testing.T) {
	p := Ping{
		Count:        2,
		Interface:    "eth0",
		Timeout:      12.0,
		Deadline:     24,
		PingInterval: 1.2,
	}

	var systemCases = []struct {
		system string
		output []string
	}{
		{"darwin", []string{"-c", "2", "-n", "-s", "16", "-i", "1.2", "-W", "12000", "-t", "24", "-I", "eth0", "www.google.com"}},
		{"linux", []string{"-c", "2", "-n", "-s", "16", "-i", "1.2", "-W", "12", "-w", "24", "-I", "eth0", "www.google.com"}},
		{"anything else", []string{"-c", "2", "-n", "-s", "16", "-i", "1.2", "-W", "12", "-w", "24", "-i", "eth0", "www.google.com"}},
	}
	for i := range systemCases {
		actual := p.args("www.google.com", systemCases[i].system)
		expected := systemCases[i].output
		sort.Strings(actual)
		sort.Strings(expected)
		require.True(t, reflect.DeepEqual(expected, actual),
			"Expected: %s Actual: %s", expected, actual)
	}
}

// Test that default arg lists for ping6 are created correctly
func TestArgs6(t *testing.T) {
	p := Ping{
		Count:        2,
		Interface:    "eth0",
		Timeout:      12.0,
		Deadline:     24,
		PingInterval: 1.2,
		Binary:       "ping6",
	}

	var systemCases = []struct {
		system string
		output []string
	}{
		{"freebsd", []string{"-c", "2", "-n", "-s", "16", "-i", "1.2", "-x", "12000", "-X", "24", "-S", "eth0", "www.google.com"}},
		{"linux", []string{"-c", "2", "-n", "-s", "16", "-i", "1.2", "-W", "12", "-w", "24", "-I", "eth0", "www.google.com"}},
		{"anything else", []string{"-c", "2", "-n", "-s", "16", "-i", "1.2", "-W", "12", "-w", "24", "-i", "eth0", "www.google.com"}},
	}
	for i := range systemCases {
		actual := p.args("www.google.com", systemCases[i].system)
		expected := systemCases[i].output
		sort.Strings(actual)
		sort.Strings(expected)
		require.True(t, reflect.DeepEqual(expected, actual),
			"Expected: %s Actual: %s", expected, actual)
	}
}

func TestArguments(t *testing.T) {
	arguments := []string{"-c", "3"}
	expected := append(arguments, "www.google.com")
	p := Ping{
		Count:        2,
		Interface:    "eth0",
		Timeout:      12.0,
		Deadline:     24,
		PingInterval: 1.2,
		Arguments:    arguments,
	}

	for _, system := range []string{"darwin", "linux", "anything else"} {
		actual := p.args("www.google.com", system)
		require.True(t, reflect.DeepEqual(actual, expected), "Expected: %s Actual: %s", expected, actual)
	}
}

func mockHostPinger(_ string, _ float64, _ ...string) (string, error) {
	return linuxPingOutput, nil
}

// Test that Gather function works on a normal ping
func TestPingGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:     []string{"localhost", "influxdata.com"},
		pingHost: mockHostPinger,
	}

	require.NoError(t, acc.GatherError(p.Gather))
	tags := map[string]string{"url": "localhost"}
	fields := map[string]interface{}{
		"packets_transmitted":   5,
		"packets_received":      5,
		"percent_packet_loss":   0.0,
		"ttl":                   63,
		"minimum_response_ms":   35.225,
		"average_response_ms":   43.628,
		"maximum_response_ms":   51.806,
		"standard_deviation_ms": 5.325,
		"result_code":           0,
	}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)

	tags = map[string]string{"url": "influxdata.com"}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)
}

func TestPingGatherIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode, retrieves systems ping utility")
	}

	var acc testutil.Accumulator
	p, ok := inputs.Inputs["ping"]().(*Ping)
	p.Log = testutil.Logger{}
	require.True(t, ok)
	p.Urls = []string{"localhost", "influxdata.com"}
	require.NoError(t, acc.GatherError(p.Gather))

	require.Equal(t, 0, acc.Metrics[0].Fields["result_code"])
	require.Equal(t, 0, acc.Metrics[1].Fields["result_code"])
}

var lossyPingOutput = `
PING www.google.com (216.58.218.164) 56(84) bytes of data.
64 bytes from host.net (216.58.218.164): icmp_seq=1 ttl=63 time=35.2 ms
64 bytes from host.net (216.58.218.164): icmp_seq=3 ttl=63 time=45.1 ms
64 bytes from host.net (216.58.218.164): icmp_seq=5 ttl=63 time=51.8 ms

--- www.google.com ping statistics ---
5 packets transmitted, 3 received, 40% packet loss, time 4010ms
rtt min/avg/max/mdev = 35.225/44.033/51.806/5.325 ms
`

func mockLossyHostPinger(_ string, _ float64, _ ...string) (string, error) {
	return lossyPingOutput, nil
}

// Test that Gather works on a ping with lossy packets
func TestLossyPingGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:     []string{"www.google.com"},
		pingHost: mockLossyHostPinger,
	}

	require.NoError(t, acc.GatherError(p.Gather))
	tags := map[string]string{"url": "www.google.com"}
	fields := map[string]interface{}{
		"packets_transmitted":   5,
		"packets_received":      3,
		"percent_packet_loss":   40.0,
		"ttl":                   63,
		"minimum_response_ms":   35.225,
		"average_response_ms":   44.033,
		"maximum_response_ms":   51.806,
		"standard_deviation_ms": 5.325,
		"result_code":           0,
	}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)
}

var errorPingOutput = `
PING www.amazon.com (176.32.98.166): 56 data bytes
Request timeout for icmp_seq 0

--- www.amazon.com ping statistics ---
2 packets transmitted, 0 packets received, 100.0% packet loss
`

func mockErrorHostPinger(_ string, _ float64, _ ...string) (string, error) {
	// This error will not trigger correct error paths
	return errorPingOutput, nil
}

// Test that Gather works on a ping with no transmitted packets, even though the
// command returns an error
func TestBadPingGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:     []string{"www.amazon.com"},
		pingHost: mockErrorHostPinger,
	}

	require.NoError(t, acc.GatherError(p.Gather))
	tags := map[string]string{"url": "www.amazon.com"}
	fields := map[string]interface{}{
		"packets_transmitted": 2,
		"packets_received":    0,
		"percent_packet_loss": 100.0,
		"result_code":         0,
	}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)
}

func mockFatalHostPinger(_ string, _ float64, _ ...string) (string, error) {
	return fatalPingOutput, errors.New("so very bad")
}

// Test that a fatal ping command does not gather any statistics.
func TestFatalPingGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:     []string{"www.amazon.com"},
		pingHost: mockFatalHostPinger,
	}

	err := acc.GatherError(p.Gather)
	require.Error(t, err)
	require.EqualValues(t, err.Error(), "host www.amazon.com: ping: -i interval too short: Operation not permitted, so very bad")
	require.False(t, acc.HasMeasurement("packets_transmitted"),
		"Fatal ping should not have packet measurements")
	require.False(t, acc.HasMeasurement("packets_received"),
		"Fatal ping should not have packet measurements")
	require.False(t, acc.HasMeasurement("percent_packet_loss"),
		"Fatal ping should not have packet measurements")
	require.False(t, acc.HasMeasurement("ttl"),
		"Fatal ping should not have packet measurements")
	require.False(t, acc.HasMeasurement("minimum_response_ms"),
		"Fatal ping should not have packet measurements")
	require.False(t, acc.HasMeasurement("average_response_ms"),
		"Fatal ping should not have packet measurements")
	require.False(t, acc.HasMeasurement("maximum_response_ms"),
		"Fatal ping should not have packet measurements")
}

func TestErrorWithHostNamePingGather(t *testing.T) {
	params := []struct {
		out   string
		error error
	}{
		{"", errors.New("host www.amazon.com: so very bad")},
		{"so bad", errors.New("host www.amazon.com: so bad, so very bad")},
	}

	for _, param := range params {
		var acc testutil.Accumulator
		p := Ping{
			Urls: []string{"www.amazon.com"},
			pingHost: func(binary string, timeout float64, args ...string) (string, error) {
				return param.out, errors.New("so very bad")
			},
		}
		require.Error(t, acc.GatherError(p.Gather))
		require.True(t, len(acc.Errors) > 0)
		require.Contains(t, acc.Errors, param.error)
	}
}

func TestPingBinary(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:   []string{"www.google.com"},
		Binary: "ping6",
		pingHost: func(binary string, timeout float64, args ...string) (string, error) {
			require.True(t, binary == "ping6")
			return "", nil
		},
	}
	err := acc.GatherError(p.Gather)
	require.Error(t, err)
	require.EqualValues(t, err.Error(), "fatal error processing ping output: www.google.com")
}

// Test that Gather function works using native ping
func TestPingGatherNative(t *testing.T) {
	type test struct {
		P *Ping
	}

	fakePingFunc := func(destination string) (*pingStats, error) {
		s := &pingStats{
			Statistics: ping.Statistics{
				PacketsSent: 5,
				PacketsRecv: 5,
				Rtts: []time.Duration{
					3 * time.Millisecond,
					4 * time.Millisecond,
					1 * time.Millisecond,
					5 * time.Millisecond,
					2 * time.Millisecond,
				},
			},
			ttl: 1,
		}

		return s, nil
	}

	tests := []test{
		{
			P: &Ping{
				Urls:           []string{"localhost", "127.0.0.2"},
				Method:         "native",
				Count:          5,
				Percentiles:    []int{50, 95, 99},
				nativePingFunc: fakePingFunc,
			},
		},
		{
			P: &Ping{
				Urls:           []string{"localhost", "127.0.0.2"},
				Method:         "native",
				Count:          5,
				PingInterval:   1,
				Percentiles:    []int{50, 95, 99},
				nativePingFunc: fakePingFunc,
			},
		},
	}

	for _, tc := range tests {
		var acc testutil.Accumulator
		require.NoError(t, tc.P.Init())
		require.NoError(t, acc.GatherError(tc.P.Gather))
		require.True(t, acc.HasPoint("ping", map[string]string{"url": "localhost"}, "packets_transmitted", 5))
		require.True(t, acc.HasPoint("ping", map[string]string{"url": "localhost"}, "packets_received", 5))
		require.True(t, acc.HasField("ping", "percentile50_ms"))
		require.Equal(t, float64(3), acc.Metrics[0].Fields["percentile50_ms"])
		require.True(t, acc.HasField("ping", "percentile95_ms"))
		require.Equal(t, float64(4.799999), acc.Metrics[0].Fields["percentile95_ms"])
		require.True(t, acc.HasField("ping", "percentile99_ms"))
		require.Equal(t, float64(4.96), acc.Metrics[0].Fields["percentile99_ms"])
		require.True(t, acc.HasField("ping", "percent_packet_loss"))
		require.True(t, acc.HasField("ping", "minimum_response_ms"))
		require.True(t, acc.HasField("ping", "average_response_ms"))
		require.True(t, acc.HasField("ping", "maximum_response_ms"))
		require.True(t, acc.HasField("ping", "standard_deviation_ms"))
	}
}

func TestNoPacketsSent(t *testing.T) {
	p := &Ping{
		Log:         testutil.Logger{},
		Urls:        []string{"localhost", "127.0.0.2"},
		Method:      "native",
		Count:       5,
		Percentiles: []int{50, 95, 99},
		nativePingFunc: func(destination string) (*pingStats, error) {
			s := &pingStats{
				Statistics: ping.Statistics{
					PacketsSent: 0,
					PacketsRecv: 0,
				},
			}

			return s, nil
		},
	}

	var testAcc testutil.Accumulator
	require.NoError(t, p.Init())

	p.pingToURLNative("localhost", &testAcc)
	require.Zero(t, testAcc.Errors)
	require.True(t, testAcc.HasField("ping", "result_code"))
	require.Equal(t, 2, testAcc.Metrics[0].Fields["result_code"])
}

// Test failed DNS resolutions
func TestDNSLookupError(t *testing.T) {
	p := &Ping{
		Count:  1,
		Log:    testutil.Logger{},
		Urls:   []string{"localhost"},
		Method: "native",
		IPv6:   false,
		nativePingFunc: func(destination string) (*pingStats, error) {
			return nil, fmt.Errorf("unknown")
		},
	}

	var testAcc testutil.Accumulator
	require.NoError(t, p.Init())

	p.pingToURLNative("localhost", &testAcc)
	require.Zero(t, testAcc.Errors)
	require.True(t, testAcc.HasField("ping", "result_code"))
	require.Equal(t, 1, testAcc.Metrics[0].Fields["result_code"])
}
