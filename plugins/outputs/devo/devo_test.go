package devo

import (
	"bufio"
	"net"
	"sync"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDevoWriter_tcp(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	dw := newDevoWriter()
	dw.Address = "tcp://" + listener.Addr().String()

	err = dw.Connect()
	require.NoError(t, err)

	lconn, err := listener.Accept()
	require.NoError(t, err)

	testSocketWriter_stream(t, dw, lconn)
}

func TestDevoWriter_udp(t *testing.T) {
	listener, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	dw := newDevoWriter()
	dw.Address = "udp://" + listener.LocalAddr().String()

	err = dw.Connect()
	require.NoError(t, err)

	testSocketWriter_packet(t, dw, listener)
}

func testSocketWriter_stream(t *testing.T, dw *DevoWriter, lconn net.Conn) {
	metrics := []telegraf.Metric{}
	metrics = append(metrics, testutil.TestMetric(1, "test"))
	mbs1out, _ := dw.Serialize(metrics[0])
	mbs1out, _ = dw.mapper.devoMapper(metrics[0], mbs1out)
	mbs1out, _ = dw.encoder.Encode(mbs1out)
	metrics = append(metrics, testutil.TestMetric(2, "test"))
	mbs2out, _ := dw.Serialize(metrics[1])
	mbs2out, _ = dw.mapper.devoMapper(metrics[1], mbs2out)
	mbs2out, _ = dw.encoder.Encode(mbs2out)

	err := dw.Write(metrics)
	require.NoError(t, err)

	scnr := bufio.NewScanner(lconn)
	require.True(t, scnr.Scan())
	mstr1in := scnr.Text() + "\n"
	require.True(t, scnr.Scan())
	mstr2in := scnr.Text() + "\n"

	assert.Equal(t, string(mbs1out), mstr1in)
	assert.Equal(t, string(mbs2out), mstr2in)
}

func testSocketWriter_packet(t *testing.T, dw *DevoWriter, lconn net.PacketConn) {
	metrics := []telegraf.Metric{}
	metrics = append(metrics, testutil.TestMetric(1, "test"))
	mbs1out, _ := dw.Serialize(metrics[0])
	mbs1out, _ = dw.mapper.devoMapper(metrics[0], mbs1out)
	mbs1out, _ = dw.encoder.Encode(mbs1out)
	mbs1str := string(mbs1out)
	metrics = append(metrics, testutil.TestMetric(2, "test"))
	mbs2out, _ := dw.Serialize(metrics[1])
	mbs2out, _ = dw.mapper.devoMapper(metrics[1], mbs2out)
	mbs2out, _ = dw.encoder.Encode(mbs2out)
	mbs2str := string(mbs2out)

	err := dw.Write(metrics)
	require.NoError(t, err)

	buf := make([]byte, 256)
	var mstrins []string
	for len(mstrins) < 2 {
		n, _, err := lconn.ReadFrom(buf)
		require.NoError(t, err)
		mstrins = append(mstrins, string(buf[:n]))
	}
	require.Len(t, mstrins, 2)

	assert.Equal(t, mbs1str, mstrins[0])
	assert.Equal(t, mbs2str, mstrins[1])
}

func TestSocketWriter_Write_reconnect(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	dw := newDevoWriter()
	dw.Address = "tcp://" + listener.Addr().String()

	err = dw.Connect()
	require.NoError(t, err)
	dw.Conn.(*net.TCPConn).SetReadBuffer(256)

	lconn, err := listener.Accept()
	require.NoError(t, err)
	lconn.(*net.TCPConn).SetWriteBuffer(256)
	lconn.Close()
	dw.Conn = nil

	wg := sync.WaitGroup{}
	wg.Add(1)
	var lerr error
	go func() {
		lconn, lerr = listener.Accept()
		wg.Done()
	}()

	metrics := []telegraf.Metric{testutil.TestMetric(1, "testerr")}
	err = dw.Write(metrics)
	require.NoError(t, err)

	wg.Wait()
	assert.NoError(t, lerr)

	mbsout, _ := dw.Serialize(metrics[0])
	mbsout, _ = dw.mapper.devoMapper(metrics[0], mbsout)
	buf := make([]byte, 256)
	n, err := lconn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, string(mbsout), string(buf[:n]))
}
