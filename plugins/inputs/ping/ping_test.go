// +build !windows

package ping

import (
	"errors"
	"reflect"
	"runtime"
	"sort"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
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

// Fatal ping output (invalid argument)
var fatalPingOutput = `
ping: -i interval too short: Operation not permitted
`

// Test that ping command output is processed properly
func TestProcessPingOutput(t *testing.T) {
	trans, rec, avg, err := processPingOutput(bsdPingOutput)
	assert.NoError(t, err)
	assert.Equal(t, 5, trans, "5 packets were transmitted")
	assert.Equal(t, 5, rec, "5 packets were transmitted")
	assert.InDelta(t, 20.224, avg, 0.001)

	trans, rec, avg, err = processPingOutput(linuxPingOutput)
	assert.NoError(t, err)
	assert.Equal(t, 5, trans, "5 packets were transmitted")
	assert.Equal(t, 5, rec, "5 packets were transmitted")
	assert.InDelta(t, 43.628, avg, 0.001)
}

// Test that processPingOutput returns an error when 'ping' fails to run, such
// as when an invalid argument is provided
func TestErrorProcessPingOutput(t *testing.T) {
	_, _, _, err := processPingOutput(fatalPingOutput)
	assert.Error(t, err, "Error was expected from processPingOutput")
}

// Test that arg lists and created correctly
func TestArgs(t *testing.T) {
	p := Ping{
		Count: 2,
	}

	// Actual and Expected arg lists must be sorted for reflect.DeepEqual

	actual := p.args("www.google.com")
	expected := []string{"-c", "2", "-n", "-s", "16", "www.google.com"}
	sort.Strings(actual)
	sort.Strings(expected)
	assert.True(t, reflect.DeepEqual(expected, actual),
		"Expected: %s Actual: %s", expected, actual)

	p.Interface = "eth0"
	actual = p.args("www.google.com")
	expected = []string{"-c", "2", "-n", "-s", "16", "-I", "eth0",
		"www.google.com"}
	sort.Strings(actual)
	sort.Strings(expected)
	assert.True(t, reflect.DeepEqual(expected, actual),
		"Expected: %s Actual: %s", expected, actual)

	p.Timeout = 12.0
	actual = p.args("www.google.com")
	switch runtime.GOOS {
	case "darwin", "freebsd":
		expected = []string{"-c", "2", "-n", "-s", "16", "-I", "eth0", "-t",
			"12.0", "www.google.com"}
	default:
		expected = []string{"-c", "2", "-n", "-s", "16", "-I", "eth0", "-W",
			"12.0", "www.google.com"}
	}

	sort.Strings(actual)
	sort.Strings(expected)
	assert.True(t, reflect.DeepEqual(expected, actual),
		"Expected: %s Actual: %s", expected, actual)

	p.PingInterval = 1.2
	actual = p.args("www.google.com")
	switch runtime.GOOS {
	case "darwin", "freebsd":
		expected = []string{"-c", "2", "-n", "-s", "16", "-I", "eth0", "-t",
			"12.0", "-i", "1.2", "www.google.com"}
	default:
		expected = []string{"-c", "2", "-n", "-s", "16", "-I", "eth0", "-W",
			"12.0", "-i", "1.2", "www.google.com"}
	}
	sort.Strings(actual)
	sort.Strings(expected)
	assert.True(t, reflect.DeepEqual(expected, actual),
		"Expected: %s Actual: %s", expected, actual)
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

	p.Gather(&acc)
	tags := map[string]string{"url": "www.google.com"}
	fields := map[string]interface{}{
		"packets_transmitted": 5,
		"packets_received":    5,
		"percent_packet_loss": 0.0,
		"average_response_ms": 43.628,
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

	p.Gather(&acc)
	tags := map[string]string{"url": "www.google.com"}
	fields := map[string]interface{}{
		"packets_transmitted": 5,
		"packets_received":    3,
		"percent_packet_loss": 40.0,
		"average_response_ms": 44.033,
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
	return errorPingOutput, errors.New("No packets received")
}

// Test that Gather works on a ping with no transmitted packets, even though the
// command returns an error
func TestBadPingGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:     []string{"www.amazon.com"},
		pingHost: mockErrorHostPinger,
	}

	p.Gather(&acc)
	tags := map[string]string{"url": "www.amazon.com"}
	fields := map[string]interface{}{
		"packets_transmitted": 2,
		"packets_received":    0,
		"percent_packet_loss": 100.0,
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

	p.Gather(&acc)
	assert.False(t, acc.HasMeasurement("packets_transmitted"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasMeasurement("packets_received"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasMeasurement("percent_packet_loss"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasMeasurement("average_response_ms"),
		"Fatal ping should not have packet measurements")
}
