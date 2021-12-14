package tcp_listener

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
)

const (
	testMsg = "cpu_load_short,host=server01 value=12.0 1422568543702900257\n"

	testMsgs = `
cpu_load_short,host=server02 value=12.0 1422568543702900257
cpu_load_short,host=server03 value=12.0 1422568543702900257
cpu_load_short,host=server04 value=12.0 1422568543702900257
cpu_load_short,host=server05 value=12.0 1422568543702900257
cpu_load_short,host=server06 value=12.0 1422568543702900257
`
)

func newTestTCPListener() (*TCPListener, chan []byte) {
	in := make(chan []byte, 1500)
	listener := &TCPListener{
		Log:                    testutil.Logger{},
		ServiceAddress:         "localhost:8194",
		AllowedPendingMessages: 10000,
		MaxTCPConnections:      250,
		in:                     in,
		done:                   make(chan struct{}),
	}
	return listener, in
}

// benchmark how long it takes to accept & process 100,000 metrics:
func BenchmarkTCP(b *testing.B) {
	listener := TCPListener{
		Log:                    testutil.Logger{},
		ServiceAddress:         "localhost:8198",
		AllowedPendingMessages: 100000,
		MaxTCPConnections:      250,
	}
	listener.parser, _ = parsers.NewInfluxParser()
	acc := &testutil.Accumulator{Discard: true}

	// send multiple messages to socket
	for n := 0; n < b.N; n++ {
		require.NoError(b, listener.Start(acc))

		conn, err := net.Dial("tcp", "127.0.0.1:8198")
		require.NoError(b, err)
		for i := 0; i < 100000; i++ {
			_, err := fmt.Fprint(conn, testMsg)
			require.NoError(b, err)
		}
		require.NoError(b, conn.(*net.TCPConn).CloseWrite())
		// wait for all 100,000 metrics to be processed
		buf := []byte{0}
		// will EOF when completed
		_, err = conn.Read(buf)
		require.NoError(b, err)
		listener.Stop()
	}
}

func TestHighTrafficTCP(t *testing.T) {
	listener := TCPListener{
		Log:                    testutil.Logger{},
		ServiceAddress:         "localhost:8199",
		AllowedPendingMessages: 100000,
		MaxTCPConnections:      250,
	}
	listener.parser, _ = parsers.NewInfluxParser()
	acc := &testutil.Accumulator{}

	// send multiple messages to socket
	require.NoError(t, listener.Start(acc))

	conn, err := net.Dial("tcp", "127.0.0.1:8199")
	require.NoError(t, err)
	for i := 0; i < 100000; i++ {
		_, err := fmt.Fprint(conn, testMsg)
		require.NoError(t, err)
	}
	require.NoError(t, conn.(*net.TCPConn).CloseWrite())
	buf := []byte{0}
	_, err = conn.Read(buf)
	require.Equal(t, err, io.EOF)
	listener.Stop()

	require.Equal(t, 100000, int(acc.NMetrics()))
}

func TestConnectTCP(t *testing.T) {
	listener := TCPListener{
		Log:                    testutil.Logger{},
		ServiceAddress:         "localhost:8194",
		AllowedPendingMessages: 10000,
		MaxTCPConnections:      250,
	}
	listener.parser, _ = parsers.NewInfluxParser()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	conn, err := net.Dial("tcp", "127.0.0.1:8194")
	require.NoError(t, err)

	// send single message to socket
	_, err = fmt.Fprint(conn, testMsg)
	require.NoError(t, err)
	acc.Wait(1)
	acc.AssertContainsTaggedFields(t, "cpu_load_short",
		map[string]interface{}{"value": float64(12)},
		map[string]string{"host": "server01"},
	)

	// send multiple messages to socket
	_, err = fmt.Fprint(conn, testMsgs)
	require.NoError(t, err)
	acc.Wait(6)
	hostTags := []string{"server02", "server03",
		"server04", "server05", "server06"}
	for _, hostTag := range hostTags {
		acc.AssertContainsTaggedFields(t, "cpu_load_short",
			map[string]interface{}{"value": float64(12)},
			map[string]string{"host": hostTag},
		)
	}
}

// Test that MaxTCPConnections is respected
func TestConcurrentConns(t *testing.T) {
	listener := TCPListener{
		Log:                    testutil.Logger{},
		ServiceAddress:         "localhost:8195",
		AllowedPendingMessages: 10000,
		MaxTCPConnections:      2,
	}
	listener.parser, _ = parsers.NewInfluxParser()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	_, err := net.Dial("tcp", "127.0.0.1:8195")
	require.NoError(t, err)
	_, err = net.Dial("tcp", "127.0.0.1:8195")
	require.NoError(t, err)

	// Connection over the limit:
	conn, err := net.Dial("tcp", "127.0.0.1:8195")
	require.NoError(t, err)
	_, err = net.Dial("tcp", "127.0.0.1:8195")
	require.NoError(t, err)
	buf := make([]byte, 1500)
	n, err := conn.Read(buf)
	require.NoError(t, err)
	require.Equal(t,
		"Telegraf maximum concurrent TCP connections (2) reached, closing.\n"+
			"You may want to increase max_tcp_connections in"+
			" the Telegraf tcp listener configuration.\n",
		string(buf[:n]))

	_, err = conn.Read(buf)
	require.Equal(t, io.EOF, err)
}

// Test that MaxTCPConnections is respected when max==1
func TestConcurrentConns1(t *testing.T) {
	listener := TCPListener{
		Log:                    testutil.Logger{},
		ServiceAddress:         "localhost:8196",
		AllowedPendingMessages: 10000,
		MaxTCPConnections:      1,
	}
	listener.parser, _ = parsers.NewInfluxParser()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	_, err := net.Dial("tcp", "127.0.0.1:8196")
	require.NoError(t, err)

	// Connection over the limit:
	conn, err := net.Dial("tcp", "127.0.0.1:8196")
	require.NoError(t, err)
	_, err = net.Dial("tcp", "127.0.0.1:8196")
	require.NoError(t, err)
	buf := make([]byte, 1500)
	n, err := conn.Read(buf)
	require.NoError(t, err)
	require.Equal(t,
		"Telegraf maximum concurrent TCP connections (1) reached, closing.\n"+
			"You may want to increase max_tcp_connections in"+
			" the Telegraf tcp listener configuration.\n",
		string(buf[:n]))

	_, err = conn.Read(buf)
	require.Equal(t, io.EOF, err)
}

// Test that MaxTCPConnections is respected
func TestCloseConcurrentConns(t *testing.T) {
	listener := TCPListener{
		Log:                    testutil.Logger{},
		ServiceAddress:         "localhost:8195",
		AllowedPendingMessages: 10000,
		MaxTCPConnections:      2,
	}
	listener.parser, _ = parsers.NewInfluxParser()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))

	_, err := net.Dial("tcp", "127.0.0.1:8195")
	require.NoError(t, err)
	_, err = net.Dial("tcp", "127.0.0.1:8195")
	require.NoError(t, err)

	listener.Stop()
}

func TestRunParser(t *testing.T) {
	var testmsg = []byte(testMsg)

	listener, in := newTestTCPListener()
	acc := testutil.Accumulator{}
	listener.acc = &acc
	defer close(listener.done)

	listener.parser, _ = parsers.NewInfluxParser()
	listener.wg.Add(1)
	go listener.tcpParser()

	in <- testmsg
	require.NoError(t, listener.Gather(&acc))

	acc.Wait(1)
	acc.AssertContainsTaggedFields(t, "cpu_load_short",
		map[string]interface{}{"value": float64(12)},
		map[string]string{"host": "server01"},
	)
}

func TestRunParserInvalidMsg(_ *testing.T) {
	var testmsg = []byte("cpu_load_short")

	listener, in := newTestTCPListener()
	acc := testutil.Accumulator{}
	listener.acc = &acc
	defer close(listener.done)

	listener.parser, _ = parsers.NewInfluxParser()
	listener.wg.Add(1)

	buf := bytes.NewBuffer(nil)
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)

	go listener.tcpParser()
	in <- testmsg

	scnr := bufio.NewScanner(buf)
	for scnr.Scan() {
		if strings.Contains(scnr.Text(), "tcp_listener has received 1 malformed packets thus far.") {
			break
		}
	}
}

func TestRunParserGraphiteMsg(t *testing.T) {
	var testmsg = []byte("cpu.load.graphite 12 1454780029")

	listener, in := newTestTCPListener()
	acc := testutil.Accumulator{}
	listener.acc = &acc
	defer close(listener.done)

	listener.parser, _ = parsers.NewGraphiteParser("_", []string{}, nil)
	listener.wg.Add(1)
	go listener.tcpParser()

	in <- testmsg
	require.NoError(t, listener.Gather(&acc))

	acc.Wait(1)
	acc.AssertContainsFields(t, "cpu_load_graphite",
		map[string]interface{}{"value": float64(12)})
}

func TestRunParserJSONMsg(t *testing.T) {
	var testmsg = []byte("{\"a\": 5, \"b\": {\"c\": 6}}\n")

	listener, in := newTestTCPListener()
	acc := testutil.Accumulator{}
	listener.acc = &acc
	defer close(listener.done)

	listener.parser, _ = parsers.NewParser(&parsers.Config{
		DataFormat: "json",
		MetricName: "udp_json_test",
	})
	listener.wg.Add(1)
	go listener.tcpParser()

	in <- testmsg
	require.NoError(t, listener.Gather(&acc))

	acc.Wait(1)
	acc.AssertContainsFields(t, "udp_json_test",
		map[string]interface{}{
			"a":   float64(5),
			"b_c": float64(6),
		})
}
