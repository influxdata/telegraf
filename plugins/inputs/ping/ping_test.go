// +build !windows

package ping

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	trans, rec, min, avg, max, stddev, err := processPingOutput(bsdPingOutput)
	assert.NoError(t, err)
	assert.Equal(t, 5, trans, "5 packets were transmitted")
	assert.Equal(t, 5, rec, "5 packets were transmitted")
	assert.InDelta(t, 15.087, min, 0.001)
	assert.InDelta(t, 20.224, avg, 0.001)
	assert.InDelta(t, 27.263, max, 0.001)
	assert.InDelta(t, 4.076, stddev, 0.001)

	trans, rec, min, avg, max, stddev, err = processPingOutput(linuxPingOutput)
	assert.NoError(t, err)
	assert.Equal(t, 5, trans, "5 packets were transmitted")
	assert.Equal(t, 5, rec, "5 packets were transmitted")
	assert.InDelta(t, 35.225, min, 0.001)
	assert.InDelta(t, 43.628, avg, 0.001)
	assert.InDelta(t, 51.806, max, 0.001)
	assert.InDelta(t, 5.325, stddev, 0.001)

	trans, rec, min, avg, max, stddev, err = processPingOutput(busyBoxPingOutput)
	assert.NoError(t, err)
	assert.Equal(t, 4, trans, "4 packets were transmitted")
	assert.Equal(t, 4, rec, "4 packets were transmitted")
	assert.InDelta(t, 15.810, min, 0.001)
	assert.InDelta(t, 17.611, avg, 0.001)
	assert.InDelta(t, 22.559, max, 0.001)
	assert.InDelta(t, -1.0, stddev, 0.001)
}

// Test that processPingOutput returns an error when 'ping' fails to run, such
// as when an invalid argument is provided
func TestErrorProcessPingOutput(t *testing.T) {
	_, _, _, _, _, _, err := processPingOutput(fatalPingOutput)
	assert.Error(t, err, "Error was expected from processPingOutput")
}

// Test that arg lists and created correctly
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
		{"darwin", []string{"-c", "2", "-n", "-s", "16", "-i", "1.2", "-W", "12000", "-t", "24", "-S", "eth0", "www.google.com"}},
		{"linux", []string{"-c", "2", "-n", "-s", "16", "-i", "1.2", "-W", "12", "-w", "24", "-I", "eth0", "www.google.com"}},
		{"anything else", []string{"-c", "2", "-n", "-s", "16", "-i", "1.2", "-W", "12", "-w", "24", "-I", "eth0", "www.google.com"}},
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

func mockHostPinger(timeout float64, args ...string) (string, error) {
	return linuxPingOutput, nil
}

// Test that Gather function works on a normal ping
func TestPingGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:     []string{"www.google.com", "www.reddit.com"},
		pingHost: mockHostPinger,
	}

	acc.GatherError(p.Gather)
	tags := map[string]string{"url": "www.google.com"}
	fields := map[string]interface{}{
		"packets_transmitted":   5,
		"packets_received":      5,
		"percent_packet_loss":   0.0,
		"minimum_response_ms":   35.225,
		"average_response_ms":   43.628,
		"maximum_response_ms":   51.806,
		"standard_deviation_ms": 5.325,
		"result_code":           0,
	}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)

	tags = map[string]string{"url": "www.reddit.com"}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)
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

func mockLossyHostPinger(timeout float64, args ...string) (string, error) {
	return lossyPingOutput, nil
}

// Test that Gather works on a ping with lossy packets
func TestLossyPingGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:     []string{"www.google.com"},
		pingHost: mockLossyHostPinger,
	}

	acc.GatherError(p.Gather)
	tags := map[string]string{"url": "www.google.com"}
	fields := map[string]interface{}{
		"packets_transmitted":   5,
		"packets_received":      3,
		"percent_packet_loss":   40.0,
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

func mockErrorHostPinger(timeout float64, args ...string) (string, error) {
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

	acc.GatherError(p.Gather)
	tags := map[string]string{"url": "www.amazon.com"}
	fields := map[string]interface{}{
		"packets_transmitted": 2,
		"packets_received":    0,
		"percent_packet_loss": 100.0,
		"result_code":         0,
	}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)
}

func mockFatalHostPinger(timeout float64, args ...string) (string, error) {
	return fatalPingOutput, errors.New("So very bad")
}

// Test that a fatal ping command does not gather any statistics.
func TestFatalPingGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:     []string{"www.amazon.com"},
		pingHost: mockFatalHostPinger,
	}

	acc.GatherError(p.Gather)
	assert.False(t, acc.HasMeasurement("packets_transmitted"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasMeasurement("packets_received"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasMeasurement("percent_packet_loss"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasMeasurement("minimum_response_ms"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasMeasurement("average_response_ms"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasMeasurement("maximum_response_ms"),
		"Fatal ping should not have packet measurements")
}

func TestErrorWithHostNamePingGather(t *testing.T) {
	params := []struct {
		out   string
		error error
	}{
		{"", errors.New("host www.amazon.com: So very bad")},
		{"so bad", errors.New("host www.amazon.com: so bad, So very bad")},
	}

	for _, param := range params {
		var acc testutil.Accumulator
		p := Ping{
			Urls: []string{"www.amazon.com"},
			pingHost: func(timeout float64, args ...string) (string, error) {
				return param.out, errors.New("So very bad")
			},
		}
		acc.GatherError(p.Gather)
		assert.True(t, len(acc.Errors) > 0)
		assert.Contains(t, acc.Errors, param.error)
	}
}
