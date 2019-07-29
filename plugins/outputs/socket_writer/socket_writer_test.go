package socket_writer

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	testSocketWriter_stream(t, sw, lconn)
}

func TestSocketWriter_udp(t *testing.T) {
	listener, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)

	sw := newSocketWriter()
	sw.Address = "udp://" + listener.LocalAddr().String()

	err = sw.Connect()
	require.NoError(t, err)

	testSocketWriter_packet(t, sw, listener)
}

func TestSocketWriter_unix(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "sw.TestSocketWriter_unix.sock")

	listener, err := net.Listen("unix", sock)
	require.NoError(t, err)

	sw := newSocketWriter()
	sw.Address = "unix://" + sock

	err = sw.Connect()
	require.NoError(t, err)

	lconn, err := listener.Accept()
	require.NoError(t, err)

	testSocketWriter_stream(t, sw, lconn)
}

func TestSocketWriter_unixgram(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "sw.TSW_unixgram.sock")

	listener, err := net.ListenPacket("unixgram", sock)
	require.NoError(t, err)

	sw := newSocketWriter()
	sw.Address = "unixgram://" + sock

	err = sw.Connect()
	require.NoError(t, err)

	testSocketWriter_packet(t, sw, listener)
}

func testSocketWriter_stream(t *testing.T, sw *SocketWriter, lconn net.Conn) {
	metrics := []telegraf.Metric{}
	metrics = append(metrics, testutil.TestMetric(1, "test"))
	mbs1out, _ := sw.Serialize(metrics[0])
	metrics = append(metrics, testutil.TestMetric(2, "test"))
	mbs2out, _ := sw.Serialize(metrics[1])

	err := sw.Write(metrics)
	require.NoError(t, err)

	scnr := bufio.NewScanner(lconn)
	require.True(t, scnr.Scan())
	mstr1in := scnr.Text() + "\n"
	require.True(t, scnr.Scan())
	mstr2in := scnr.Text() + "\n"

	assert.Equal(t, string(mbs1out), mstr1in)
	assert.Equal(t, string(mbs2out), mstr2in)
}

func testSocketWriter_packet(t *testing.T, sw *SocketWriter, lconn net.PacketConn) {
	metrics := []telegraf.Metric{}
	metrics = append(metrics, testutil.TestMetric(1, "test"))
	mbs1out, _ := sw.Serialize(metrics[0])
	metrics = append(metrics, testutil.TestMetric(2, "test"))
	mbs2out, _ := sw.Serialize(metrics[1])

	err := sw.Write(metrics)
	require.NoError(t, err)

	buf := make([]byte, 256)
	var mstrins []string
	for len(mstrins) < 2 {
		n, _, err := lconn.ReadFrom(buf)
		require.NoError(t, err)
		for _, bs := range bytes.Split(buf[:n], []byte{'\n'}) {
			if len(bs) == 0 {
				continue
			}
			mstrins = append(mstrins, string(bs)+"\n")
		}
	}
	require.Len(t, mstrins, 2)

	assert.Equal(t, string(mbs1out), mstrins[0])
	assert.Equal(t, string(mbs2out), mstrins[1])
}

func TestSocketWriter_Write_err(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	sw := newSocketWriter()
	sw.Address = "tcp://" + listener.Addr().String()

	err = sw.Connect()
	require.NoError(t, err)
	sw.Conn.(*net.TCPConn).SetReadBuffer(256)

	lconn, err := listener.Accept()
	require.NoError(t, err)
	lconn.(*net.TCPConn).SetWriteBuffer(256)

	metrics := []telegraf.Metric{testutil.TestMetric(1, "testerr")}

	// close the socket to generate an error
	lconn.Close()
	sw.Conn.Close()
	err = sw.Write(metrics)
	require.Error(t, err)
	assert.Nil(t, sw.Conn)
}

func TestSocketWriter_Write_reconnect(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	sw := newSocketWriter()
	sw.Address = "tcp://" + listener.Addr().String()

	err = sw.Connect()
	require.NoError(t, err)
	sw.Conn.(*net.TCPConn).SetReadBuffer(256)

	lconn, err := listener.Accept()
	require.NoError(t, err)
	lconn.(*net.TCPConn).SetWriteBuffer(256)
	lconn.Close()
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
	assert.NoError(t, lerr)

	mbsout, _ := sw.Serialize(metrics[0])
	buf := make([]byte, 256)
	n, err := lconn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, string(mbsout), string(buf[:n]))
}
