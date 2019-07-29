// +build windows

package ping

import (
	"errors"
	"reflect"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Windows ping format ( should support multilanguage ?)
var winPLPingOutput = `
Badanie 8.8.8.8 z 32 bajtami danych:
Odpowiedz z 8.8.8.8: bajtow=32 czas=49ms TTL=43
Odpowiedz z 8.8.8.8: bajtow=32 czas=46ms TTL=43
Odpowiedz z 8.8.8.8: bajtow=32 czas=48ms TTL=43
Odpowiedz z 8.8.8.8: bajtow=32 czas=57ms TTL=43

Statystyka badania ping dla 8.8.8.8:
    Pakiety: Wyslane = 4, Odebrane = 4, Utracone = 0
             (0% straty),
Szacunkowy czas bladzenia pakietww w millisekundach:
    Minimum = 46 ms, Maksimum = 57 ms, Czas sredni = 50 ms
`

// Windows ping format ( should support multilanguage ?)
var winENPingOutput = `
Pinging 8.8.8.8 with 32 bytes of data:
Reply from 8.8.8.8: bytes=32 time=52ms TTL=43
Reply from 8.8.8.8: bytes=32 time=50ms TTL=43
Reply from 8.8.8.8: bytes=32 time=50ms TTL=43
Reply from 8.8.8.8: bytes=32 time=51ms TTL=43

Ping statistics for 8.8.8.8:
    Packets: Sent = 4, Received = 4, Lost = 0 (0% loss),
Approximate round trip times in milli-seconds:
    Minimum = 50ms, Maximum = 52ms, Average = 50ms
`

func TestHost(t *testing.T) {
	trans, recReply, recPacket, avg, min, max, err := processPingOutput(winPLPingOutput)
	assert.NoError(t, err)
	assert.Equal(t, 4, trans, "4 packets were transmitted")
	assert.Equal(t, 4, recReply, "4 packets were reply")
	assert.Equal(t, 4, recPacket, "4 packets were received")
	assert.Equal(t, 50, avg, "Average 50")
	assert.Equal(t, 46, min, "Min 46")
	assert.Equal(t, 57, max, "max 57")

	trans, recReply, recPacket, avg, min, max, err = processPingOutput(winENPingOutput)
	assert.NoError(t, err)
	assert.Equal(t, 4, trans, "4 packets were transmitted")
	assert.Equal(t, 4, recReply, "4 packets were reply")
	assert.Equal(t, 4, recPacket, "4 packets were received")
	assert.Equal(t, 50, avg, "Average 50")
	assert.Equal(t, 50, min, "Min 50")
	assert.Equal(t, 52, max, "Max 52")
}

func mockHostPinger(binary string, timeout float64, args ...string) (string, error) {
	return winENPingOutput, nil
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
		"packets_transmitted": 4,
		"packets_received":    4,
		"reply_received":      4,
		"percent_packet_loss": 0.0,
		"percent_reply_loss":  0.0,
		"average_response_ms": 50.0,
		"minimum_response_ms": 50.0,
		"maximum_response_ms": 52.0,
		"result_code":         0,
	}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)

	tags = map[string]string{"url": "www.reddit.com"}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)
}

var errorPingOutput = `
Badanie nask.pl [195.187.242.157] z 32 bajtami danych:
Upłynął limit czasu żądania.
Upłynął limit czasu żądania.
Upłynął limit czasu żądania.
Upłynął limit czasu żądania.

Statystyka badania ping dla 195.187.242.157:
    Pakiety: Wysłane = 4, Odebrane = 0, Utracone = 4
             (100% straty),
`

func mockErrorHostPinger(binary string, timeout float64, args ...string) (string, error) {
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

	acc.GatherError(p.Gather)
	tags := map[string]string{"url": "www.amazon.com"}
	fields := map[string]interface{}{
		"packets_transmitted": 4,
		"packets_received":    0,
		"reply_received":      0,
		"percent_packet_loss": 100.0,
		"percent_reply_loss":  100.0,
		"result_code":         0,
	}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)
}

func TestArguments(t *testing.T) {
	arguments := []string{"-c", "3"}
	p := Ping{
		Count:     2,
		Timeout:   12.0,
		Arguments: arguments,
	}

	actual := p.args("www.google.com")
	require.True(t, reflect.DeepEqual(actual, arguments), "Expected : %s Actual: %s", arguments, actual)
}

var lossyPingOutput = `
Badanie thecodinglove.com [66.6.44.4] z 9800 bajtami danych:
Upłynął limit czasu żądania.
Odpowiedź z 66.6.44.4: bajtów=9800 czas=114ms TTL=48
Odpowiedź z 66.6.44.4: bajtów=9800 czas=114ms TTL=48
Odpowiedź z 66.6.44.4: bajtów=9800 czas=118ms TTL=48
Odpowiedź z 66.6.44.4: bajtów=9800 czas=114ms TTL=48
Odpowiedź z 66.6.44.4: bajtów=9800 czas=114ms TTL=48
Upłynął limit czasu żądania.
Odpowiedź z 66.6.44.4: bajtów=9800 czas=119ms TTL=48
Odpowiedź z 66.6.44.4: bajtów=9800 czas=116ms TTL=48

Statystyka badania ping dla 66.6.44.4:
    Pakiety: Wysłane = 9, Odebrane = 7, Utracone = 2
             (22% straty),
Szacunkowy czas błądzenia pakietów w millisekundach:
    Minimum = 114 ms, Maksimum = 119 ms, Czas średni = 115 ms
`

func mockLossyHostPinger(binary string, timeout float64, args ...string) (string, error) {
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
		"packets_transmitted": 9,
		"packets_received":    7,
		"reply_received":      7,
		"percent_packet_loss": 22.22222222222222,
		"percent_reply_loss":  22.22222222222222,
		"average_response_ms": 115.0,
		"minimum_response_ms": 114.0,
		"maximum_response_ms": 119.0,
		"result_code":         0,
	}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)
}

// Fatal ping output (invalid argument)
var fatalPingOutput = `
Bad option -d.


Usage: ping [-t] [-a] [-n count] [-l size] [-f] [-i TTL] [-v TOS]
            [-r count] [-s count] [[-j host-list] | [-k host-list]]
            [-w timeout] [-R] [-S srcaddr] [-4] [-6] target_name

Options:
    -t             Ping the specified host until stopped.
                   To see statistics and continue - type Control-Break;
                   To stop - type Control-C.
    -a             Resolve addresses to hostnames.
    -n count       Number of echo requests to send.
    -l size        Send buffer size.
    -f             Set Don't Fragment flag in packet (IPv4-only).
    -i TTL         Time To Live.
    -v TOS         Type Of Service (IPv4-only. This setting has been deprecated
                   and has no effect on the type of service field in the IP Header).
    -r count       Record route for count hops (IPv4-only).
    -s count       Timestamp for count hops (IPv4-only).
    -j host-list   Loose source route along host-list (IPv4-only).
    -k host-list   Strict source route along host-list (IPv4-only).
    -w timeout     Timeout in milliseconds to wait for each reply.
    -R             Use routing header to test reverse route also (IPv6-only).
    -S srcaddr     Source address to use.
    -4             Force using IPv4.
    -6             Force using IPv6.

`

func mockFatalHostPinger(binary string, timeout float64, args ...string) (string, error) {
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
	assert.True(t, acc.HasFloatField("ping", "errors"),
		"Fatal ping should have packet measurements")
	assert.False(t, acc.HasInt64Field("ping", "packets_transmitted"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasInt64Field("ping", "packets_received"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasFloatField("ping", "percent_packet_loss"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasFloatField("ping", "percent_reply_loss"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasInt64Field("ping", "average_response_ms"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasInt64Field("ping", "maximum_response_ms"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasInt64Field("ping", "minimum_response_ms"),
		"Fatal ping should not have packet measurements")
}

var UnreachablePingOutput = `
Pinging www.google.pl [8.8.8.8] with 32 bytes of data:
Request timed out.
Request timed out.
Reply from 194.204.175.50: Destination net unreachable.
Request timed out.

Ping statistics for 8.8.8.8:
    Packets: Sent = 4, Received = 1, Lost = 3 (75% loss),
`

func mockUnreachableHostPinger(binary string, timeout float64, args ...string) (string, error) {
	return UnreachablePingOutput, errors.New("So very bad")
}

//Reply from 185.28.251.217: TTL expired in transit.

// in case 'Destination net unreachable' ping app return receive packet which is not what we need
// it's not contain valid metric so treat it as lost one
func TestUnreachablePingGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:     []string{"www.google.com"},
		pingHost: mockUnreachableHostPinger,
	}

	acc.GatherError(p.Gather)

	tags := map[string]string{"url": "www.google.com"}
	fields := map[string]interface{}{
		"packets_transmitted": 4,
		"packets_received":    1,
		"reply_received":      0,
		"percent_packet_loss": 75.0,
		"percent_reply_loss":  100.0,
		"result_code":         0,
	}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)

	assert.False(t, acc.HasFloatField("ping", "errors"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasInt64Field("ping", "average_response_ms"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasInt64Field("ping", "maximum_response_ms"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasInt64Field("ping", "minimum_response_ms"),
		"Fatal ping should not have packet measurements")
}

var TTLExpiredPingOutput = `
Pinging www.google.pl [8.8.8.8] with 32 bytes of data:
Request timed out.
Request timed out.
Reply from 185.28.251.217: TTL expired in transit.
Request timed out.

Ping statistics for 8.8.8.8:
    Packets: Sent = 4, Received = 1, Lost = 3 (75% loss),
`

func mockTTLExpiredPinger(binary string, timeout float64, args ...string) (string, error) {
	return TTLExpiredPingOutput, errors.New("So very bad")
}

// in case 'Destination net unreachable' ping app return receive packet which is not what we need
// it's not contain valid metric so treat it as lost one
func TestTTLExpiredPingGather(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:     []string{"www.google.com"},
		pingHost: mockTTLExpiredPinger,
	}

	acc.GatherError(p.Gather)

	tags := map[string]string{"url": "www.google.com"}
	fields := map[string]interface{}{
		"packets_transmitted": 4,
		"packets_received":    1,
		"reply_received":      0,
		"percent_packet_loss": 75.0,
		"percent_reply_loss":  100.0,
		"result_code":         0,
	}
	acc.AssertContainsTaggedFields(t, "ping", fields, tags)

	assert.False(t, acc.HasFloatField("ping", "errors"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasInt64Field("ping", "average_response_ms"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasInt64Field("ping", "maximum_response_ms"),
		"Fatal ping should not have packet measurements")
	assert.False(t, acc.HasInt64Field("ping", "minimum_response_ms"),
		"Fatal ping should not have packet measurements")
}

func TestPingBinary(t *testing.T) {
	var acc testutil.Accumulator
	p := Ping{
		Urls:   []string{"www.google.com"},
		Binary: "ping6",
		pingHost: func(binary string, timeout float64, args ...string) (string, error) {
			assert.True(t, binary == "ping6")
			return "", nil
		},
	}
	acc.GatherError(p.Gather)
}
