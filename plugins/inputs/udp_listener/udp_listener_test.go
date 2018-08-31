package udp_listener

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"testing"

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
	var err error
	listener.parser, err = parsers.NewInfluxParser()
	require.NoError(t, err)
	acc := &testutil.Accumulator{}

	// send multiple messages to socket
	err = listener.Start(acc)
	require.NoError(t, err)

	conn, err := net.Dial("udp", "127.0.0.1:8126")
	require.NoError(t, err)
	mlen := int64(len(testMsgs))
	var sent int64
	for i := 0; i < 20000; i++ {
		for sent > listener.BytesRecv.Get()+32000 {
			// more than 32kb sitting in OS buffer, let it drain
			runtime.Gosched()
		}
		conn.Write([]byte(testMsgs))
		sent += mlen
	}
	for sent > listener.BytesRecv.Get() {
		runtime.Gosched()
	}
	for len(listener.in) > 0 {
		runtime.Gosched()
	}
	listener.Stop()

	assert.Equal(t, uint64(100000), acc.NMetrics())
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

	conn, err := net.Dial("udp", "127.0.0.1:8127")
	require.NoError(t, err)

	// send single message to socket
	fmt.Fprintf(conn, testMsg)
	acc.Wait(1)
	acc.AssertContainsTaggedFields(t, "cpu_load_short",
		map[string]interface{}{"value": float64(12)},
		map[string]string{"host": "server01"},
	)

	// send multiple messages to socket
	fmt.Fprintf(conn, testMsgs)
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
	listener.Gather(&acc)

	acc.Wait(1)
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

	buf := bytes.NewBuffer(nil)
	log.SetOutput(buf)
	defer log.SetOutput(os.Stderr)
	in <- testmsg

	scnr := bufio.NewScanner(buf)
	for scnr.Scan() {
		if strings.Contains(scnr.Text(), fmt.Sprintf(malformedwarn, 1)) {
			break
		}
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
	listener.Gather(&acc)

	acc.Wait(1)
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

	listener.parser, _ = parsers.NewParser(&parsers.Config{
		DataFormat: "json",
		MetricName: "udp_json_test",
	})
	listener.wg.Add(1)
	go listener.udpParser()

	in <- testmsg
	listener.Gather(&acc)

	acc.Wait(1)
	acc.AssertContainsFields(t, "udp_json_test",
		map[string]interface{}{
			"a":   float64(5),
			"b_c": float64(6),
		})
}
