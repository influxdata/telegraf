package udp_listener

import (
	"fmt"
	"io/ioutil"
	"log"
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

func newTestUdpListener() (*UdpListener, chan []byte) {
	in := make(chan []byte, 1500)
	listener := &UdpListener{
		ServiceAddress:         ":8125",
		AllowedPendingMessages: 10000,
		in:   in,
		done: make(chan struct{}),
	}
	return listener, in
}

func TestHighTrafficUDP(t *testing.T) {
	listener := UdpListener{
		ServiceAddress:         ":8126",
		AllowedPendingMessages: 100000,
	}
	listener.parser, _ = parsers.NewInfluxParser()
	acc := &testutil.Accumulator{}

	// send multiple messages to socket
	err := listener.Start(acc)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 25)
	conn, err := net.Dial("udp", "127.0.0.1:8126")
	require.NoError(t, err)
	for i := 0; i < 20000; i++ {
		// arbitrary, just to give the OS buffer some slack handling the
		// packet storm.
		time.Sleep(time.Microsecond)
		fmt.Fprintf(conn, testMsgs)
	}
	time.Sleep(time.Millisecond)
	listener.Stop()

	// this is not an exact science, since UDP packets can easily get lost or
	// dropped, but assume that the OS will be able to
	// handle at least 90% of the sent UDP packets.
	assert.InDelta(t, 100000, len(acc.Metrics), 10000)
}

func TestConnectUDP(t *testing.T) {
	listener := UdpListener{
		ServiceAddress:         ":8127",
		AllowedPendingMessages: 10000,
	}
	listener.parser, _ = parsers.NewInfluxParser()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	time.Sleep(time.Millisecond * 25)
	conn, err := net.Dial("udp", "127.0.0.1:8127")
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

func TestRunParser(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	var testmsg = []byte("cpu_load_short,host=server01 value=12.0 1422568543702900257\n")

	listener, in := newTestUdpListener()
	acc := testutil.Accumulator{}
	listener.acc = &acc
	defer close(listener.done)

	listener.parser, _ = parsers.NewInfluxParser()
	listener.wg.Add(1)
	go listener.udpParser()

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
	log.SetOutput(ioutil.Discard)
	var testmsg = []byte("cpu_load_short")

	listener, in := newTestUdpListener()
	acc := testutil.Accumulator{}
	listener.acc = &acc
	defer close(listener.done)

	listener.parser, _ = parsers.NewInfluxParser()
	listener.wg.Add(1)
	go listener.udpParser()

	in <- testmsg
	time.Sleep(time.Millisecond * 25)

	if a := acc.NFields(); a != 0 {
		t.Errorf("got %v, expected %v", a, 0)
	}
}

func TestRunParserGraphiteMsg(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	var testmsg = []byte("cpu.load.graphite 12 1454780029")

	listener, in := newTestUdpListener()
	acc := testutil.Accumulator{}
	listener.acc = &acc
	defer close(listener.done)

	listener.parser, _ = parsers.NewGraphiteParser("_", []string{}, nil)
	listener.wg.Add(1)
	go listener.udpParser()

	in <- testmsg
	time.Sleep(time.Millisecond * 25)
	listener.Gather(&acc)

	acc.AssertContainsFields(t, "cpu_load_graphite",
		map[string]interface{}{"value": float64(12)})
}

func TestRunParserJSONMsg(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	var testmsg = []byte("{\"a\": 5, \"b\": {\"c\": 6}}\n")

	listener, in := newTestUdpListener()
	acc := testutil.Accumulator{}
	listener.acc = &acc
	defer close(listener.done)

	listener.parser, _ = parsers.NewJSONParser("udp_json_test", []string{}, nil)
	listener.wg.Add(1)
	go listener.udpParser()

	in <- testmsg
	time.Sleep(time.Millisecond * 25)
	listener.Gather(&acc)

	acc.AssertContainsFields(t, "udp_json_test",
		map[string]interface{}{
			"a":   float64(5),
			"b_c": float64(6),
		})
}
