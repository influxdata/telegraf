package tcp_listener

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func newTestTcpListener() (*TcpListener, chan []byte) {
	in := make(chan []byte, 1500)
	listener := &TcpListener{
		ServiceAddress:         ":8194",
		AllowedPendingMessages: 10000,
		MaxTCPConnections:      250,
		in:                     in,
		done:                   make(chan struct{}),
	}
	return listener, in
}

// benchmark how long it takes to accept & process 100,000 metrics:
func BenchmarkTCP(b *testing.B) {
	listener := TcpListener{
		ServiceAddress:         ":8198",
		AllowedPendingMessages: 100000,
		MaxTCPConnections:      250,
	}
	listener.parser, _ = parsers.NewInfluxParser()
	acc := &testutil.Accumulator{Discard: true}

	// send multiple messages to socket
	for n := 0; n < b.N; n++ {
		err := listener.Start(acc)
		if err != nil {
			panic(err)
		}

		time.Sleep(time.Millisecond * 25)
		conn, err := net.Dial("tcp", "127.0.0.1:8198")
		if err != nil {
			panic(err)
		}
		for i := 0; i < 100000; i++ {
			fmt.Fprintf(conn, testMsg)
		}
		// wait for 100,000 metrics to get added to accumulator
		time.Sleep(time.Millisecond)
		listener.Stop()
	}
}

func TestHighTrafficTCP(t *testing.T) {
	listener := TcpListener{
		ServiceAddress:         ":8199",
		AllowedPendingMessages: 100000,
		MaxTCPConnections:      250,
	}
	listener.parser, _ = parsers.NewInfluxParser()
	acc := &testutil.Accumulator{}

	// send multiple messages to socket
	err := listener.Start(acc)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 25)
	conn, err := net.Dial("tcp", "127.0.0.1:8199")
	require.NoError(t, err)
	for i := 0; i < 100000; i++ {
		fmt.Fprintf(conn, testMsg)
	}
	time.Sleep(time.Millisecond)
	listener.Stop()

	assert.Equal(t, 100000, len(acc.Metrics))
}

func TestConnectTCP(t *testing.T) {
	listener := TcpListener{
		ServiceAddress:         ":8194",
		AllowedPendingMessages: 10000,
		MaxTCPConnections:      250,
	}
	listener.parser, _ = parsers.NewInfluxParser()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	time.Sleep(time.Millisecond * 25)
	conn, err := net.Dial("tcp", "127.0.0.1:8194")
	require.NoError(t, err)

	// send single message to socket
	fmt.Fprintf(conn, testMsg)
	time.Sleep(time.Millisecond * 15)
	acc.AssertContainsTaggedFields(t, "cpu_load_short",
		map[string]interface{}{"value": float64(12)},
		map[string]string{"host": "server01"},
	)

	// send multiple messages to socket
	fmt.Fprintf(conn, testMsgs)
	time.Sleep(time.Millisecond * 15)
	hostTags := []string{"server02", "server03",
		"server04", "server05", "server06"}
	for _, hostTag := range hostTags {
		acc.AssertContainsTaggedFields(t, "cpu_load_short",
			map[string]interface{}{"value": float64(12)},
			map[string]string{"host": hostTag},
		)
	}
}

// Test that MaxTCPConections is respected
func TestConcurrentConns(t *testing.T) {
	listener := TcpListener{
		ServiceAddress:         ":8195",
		AllowedPendingMessages: 10000,
		MaxTCPConnections:      2,
	}
	listener.parser, _ = parsers.NewInfluxParser()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	time.Sleep(time.Millisecond * 25)
	_, err := net.Dial("tcp", "127.0.0.1:8195")
	assert.NoError(t, err)
	_, err = net.Dial("tcp", "127.0.0.1:8195")
	assert.NoError(t, err)

	// Connection over the limit:
	conn, err := net.Dial("tcp", "127.0.0.1:8195")
	assert.NoError(t, err)
	net.Dial("tcp", "127.0.0.1:8195")
	buf := make([]byte, 1500)
	n, err := conn.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t,
		"Telegraf maximum concurrent TCP connections (2) reached, closing.\n"+
			"You may want to increase max_tcp_connections in"+
			" the Telegraf tcp listener configuration.\n",
		string(buf[:n]))

	_, err = conn.Write([]byte(testMsg))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 10)
	assert.Zero(t, acc.NFields())
}

// Test that MaxTCPConections is respected when max==1
func TestConcurrentConns1(t *testing.T) {
	listener := TcpListener{
		ServiceAddress:         ":8196",
		AllowedPendingMessages: 10000,
		MaxTCPConnections:      1,
	}
	listener.parser, _ = parsers.NewInfluxParser()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	time.Sleep(time.Millisecond * 25)
	_, err := net.Dial("tcp", "127.0.0.1:8196")
	assert.NoError(t, err)

	// Connection over the limit:
	conn, err := net.Dial("tcp", "127.0.0.1:8196")
	assert.NoError(t, err)
	net.Dial("tcp", "127.0.0.1:8196")
	buf := make([]byte, 1500)
	n, err := conn.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t,
		"Telegraf maximum concurrent TCP connections (1) reached, closing.\n"+
			"You may want to increase max_tcp_connections in"+
			" the Telegraf tcp listener configuration.\n",
		string(buf[:n]))

	_, err = conn.Write([]byte(testMsg))
	assert.NoError(t, err)
	time.Sleep(time.Millisecond * 10)
	assert.Zero(t, acc.NFields())
}

// Test that MaxTCPConections is respected
func TestCloseConcurrentConns(t *testing.T) {
	listener := TcpListener{
		ServiceAddress:         ":8195",
		AllowedPendingMessages: 10000,
		MaxTCPConnections:      2,
	}
	listener.parser, _ = parsers.NewInfluxParser()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))

	time.Sleep(time.Millisecond * 25)
	_, err := net.Dial("tcp", "127.0.0.1:8195")
	assert.NoError(t, err)
	_, err = net.Dial("tcp", "127.0.0.1:8195")
	assert.NoError(t, err)

	listener.Stop()
}

func TestRunParser(t *testing.T) {
	var testmsg = []byte(testMsg)

	listener, in := newTestTcpListener()
	acc := testutil.Accumulator{}
	listener.acc = &acc
	defer close(listener.done)

	listener.parser, _ = parsers.NewInfluxParser()
	listener.wg.Add(1)
	go listener.tcpParser()

	in <- testmsg
	time.Sleep(time.Millisecond * 25)
	listener.Gather(&acc)

	if a := acc.NFields(); a != 1 {
		t.Errorf("got %v, expected %v", a, 1)
	}

	acc.AssertContainsTaggedFields(t, "cpu_load_short",
		map[string]interface{}{"value": float64(12)},
		map[string]string{"host": "server01"},
	)
}

func TestRunParserInvalidMsg(t *testing.T) {
	var testmsg = []byte("cpu_load_short")

	listener, in := newTestTcpListener()
	acc := testutil.Accumulator{}
	listener.acc = &acc
	defer close(listener.done)

	listener.parser, _ = parsers.NewInfluxParser()
	listener.wg.Add(1)
	go listener.tcpParser()

	in <- testmsg
	time.Sleep(time.Millisecond * 25)

	if a := acc.NFields(); a != 0 {
		t.Errorf("got %v, expected %v", a, 0)
	}
}

func TestRunParserGraphiteMsg(t *testing.T) {
	var testmsg = []byte("cpu.load.graphite 12 1454780029")

	listener, in := newTestTcpListener()
	acc := testutil.Accumulator{}
	listener.acc = &acc
	defer close(listener.done)

	listener.parser, _ = parsers.NewGraphiteParser("_", []string{}, nil)
	listener.wg.Add(1)
	go listener.tcpParser()

	in <- testmsg
	time.Sleep(time.Millisecond * 25)
	listener.Gather(&acc)

	acc.AssertContainsFields(t, "cpu_load_graphite",
		map[string]interface{}{"value": float64(12)})
}

func TestRunParserJSONMsg(t *testing.T) {
	var testmsg = []byte("{\"a\": 5, \"b\": {\"c\": 6}}\n")

	listener, in := newTestTcpListener()
	acc := testutil.Accumulator{}
	listener.acc = &acc
	defer close(listener.done)

	listener.parser, _ = parsers.NewJSONParser("udp_json_test", []string{}, nil)
	listener.wg.Add(1)
	go listener.tcpParser()

	in <- testmsg
	time.Sleep(time.Millisecond * 25)
	listener.Gather(&acc)

	acc.AssertContainsFields(t, "udp_json_test",
		map[string]interface{}{
			"a":   float64(5),
			"b_c": float64(6),
		})
}
