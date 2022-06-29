package socket_writer

import (
	"bufio"
	"net"
	"runtime"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func TestSocketWriter_tcp(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	sw := newSocketWriter()
	sw.Address = "tcp://" + listener.Addr().String()

	err = sw.Connect()
	require.NoError(t, err)

	lconn, err := listener.Accept()
	require.NoError(t, err)

	testSocketWriterStream(t, sw, lconn)
}

func TestSocketWriter_udp(t *testing.T) {
	listener, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	sw := newSocketWriter()
	sw.Address = "udp://" + listener.LocalAddr().String()

	err = sw.Connect()
	require.NoError(t, err)

	testSocketWriterPacket(t, sw, listener)
}

func TestSocketWriter_unix(t *testing.T) {
	sock := testutil.TempSocket(t)

	listener, err := net.Listen("unix", sock)
	require.NoError(t, err)

	sw := newSocketWriter()
	sw.Address = "unix://" + sock

	err = sw.Connect()
	require.NoError(t, err)

	lconn, err := listener.Accept()
	require.NoError(t, err)

	testSocketWriterStream(t, sw, lconn)
}

func TestSocketWriter_unixgram(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows, as unixgram sockets are not supported")
	}

	sock := testutil.TempSocket(t)

	listener, err := net.ListenPacket("unixgram", sock)
	require.NoError(t, err)

	sw := newSocketWriter()
	sw.Address = "unixgram://" + sock

	err = sw.Connect()
	require.NoError(t, err)

	testSocketWriterPacket(t, sw, listener)
}

func testSocketWriterStream(t *testing.T, sw *SocketWriter, lconn net.Conn) {
	metrics := []telegraf.Metric{}
	metrics = append(metrics, testutil.TestMetric(1, "test"))
	mbs1out, _ := sw.Serialize(metrics[0])
	mbs1out, _ = sw.encoder.Encode(mbs1out)
	metrics = append(metrics, testutil.TestMetric(2, "test"))
	mbs2out, _ := sw.Serialize(metrics[1])
	mbs2out, _ = sw.encoder.Encode(mbs2out)

	err := sw.Write(metrics)
	require.NoError(t, err)

	scnr := bufio.NewScanner(lconn)
	require.True(t, scnr.Scan())
	mstr1in := scnr.Text() + "\n"
	require.True(t, scnr.Scan())
	mstr2in := scnr.Text() + "\n"

	require.Equal(t, string(mbs1out), mstr1in)
	require.Equal(t, string(mbs2out), mstr2in)
}

func testSocketWriterPacket(t *testing.T, sw *SocketWriter, lconn net.PacketConn) {
	metrics := []telegraf.Metric{}
	metrics = append(metrics, testutil.TestMetric(1, "test"))
	mbs1out, _ := sw.Serialize(metrics[0])
	mbs1out, _ = sw.encoder.Encode(mbs1out)
	mbs1str := string(mbs1out)
	metrics = append(metrics, testutil.TestMetric(2, "test"))
	mbs2out, _ := sw.Serialize(metrics[1])
	mbs2out, _ = sw.encoder.Encode(mbs2out)
	mbs2str := string(mbs2out)

	err := sw.Write(metrics)
	require.NoError(t, err)

	buf := make([]byte, 256)
	var mstrins []string
	for len(mstrins) < 2 {
		n, _, err := lconn.ReadFrom(buf)
		require.NoError(t, err)
		mstrins = append(mstrins, string(buf[:n]))
	}
	require.Len(t, mstrins, 2)

	require.Equal(t, mbs1str, mstrins[0])
	require.Equal(t, mbs2str, mstrins[1])
}

func TestSocketWriter_Write_err(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	sw := newSocketWriter()
	sw.Address = "tcp://" + listener.Addr().String()

	err = sw.Connect()
	require.NoError(t, err)
	err = sw.Conn.(*net.TCPConn).SetReadBuffer(256)
	require.NoError(t, err)

	lconn, err := listener.Accept()
	require.NoError(t, err)
	err = lconn.(*net.TCPConn).SetWriteBuffer(256)
	require.NoError(t, err)

	metrics := []telegraf.Metric{testutil.TestMetric(1, "testerr")}

	// close the socket to generate an error
	err = lconn.Close()
	require.NoError(t, err)

	err = sw.Conn.Close()
	require.NoError(t, err)

	err = sw.Write(metrics)
	require.Error(t, err)
	require.Nil(t, sw.Conn)
}

func TestSocketWriter_Write_reconnect(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	sw := newSocketWriter()
	sw.Address = "tcp://" + listener.Addr().String()

	err = sw.Connect()
	require.NoError(t, err)
	err = sw.Conn.(*net.TCPConn).SetReadBuffer(256)
	require.NoError(t, err)

	lconn, err := listener.Accept()
	require.NoError(t, err)
	err = lconn.(*net.TCPConn).SetWriteBuffer(256)
	require.NoError(t, err)

	err = lconn.Close()
	require.NoError(t, err)
	sw.Conn = nil

	wg := sync.WaitGroup{}
	wg.Add(1)
	var lerr error
	go func() {
		lconn, lerr = listener.Accept()
		wg.Done()
	}()

	metrics := []telegraf.Metric{testutil.TestMetric(1, "testerr")}
	err = sw.Write(metrics)
	require.NoError(t, err)

	wg.Wait()
	require.NoError(t, lerr)

	mbsout, _ := sw.Serialize(metrics[0])
	buf := make([]byte, 256)
	n, err := lconn.Read(buf)
	require.NoError(t, err)
	require.Equal(t, string(mbsout), string(buf[:n]))
}

func TestSocketWriter_udp_gzip(t *testing.T) {
	listener, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	sw := newSocketWriter()
	sw.Address = "udp://" + listener.LocalAddr().String()
	sw.ContentEncoding = "gzip"

	err = sw.Connect()
	require.NoError(t, err)

	testSocketWriterPacket(t, sw, listener)
}
