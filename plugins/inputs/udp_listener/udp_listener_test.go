package udp_listener

import (
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
)

func newTestUdpListener() (*UdpListener, chan []byte) {
	in := make(chan []byte, 1500)
	listener := &UdpListener{
		ServiceAddress:         ":8125",
		UDPPacketSize:          1500,
		AllowedPendingMessages: 10000,
		in:   in,
		done: make(chan struct{}),
	}
	return listener, in
}

func TestRunParser(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	var testmsg = []byte("cpu_load_short,host=server01 value=12.0 1422568543702900257")

	listener, in := newTestUdpListener()
	acc := testutil.Accumulator{}
	listener.acc = &acc
	defer close(listener.done)

	listener.parser, _ = parsers.NewInfluxParser()
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
