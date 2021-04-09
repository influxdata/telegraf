package socket_listener

import (
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/wlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var pki = testutil.NewPKI("../../../testutil/pki")

// testEmptyLog is a helper function to ensure no data is written to log.
// Should be called at the start of the test, and returns a function which should run at the end.
func testEmptyLog(t *testing.T) func() {
	buf := bytes.NewBuffer(nil)
	log.SetOutput(wlog.NewWriter(buf))

	level := wlog.WARN
	wlog.SetLevel(level)

	return func() {
		log.SetOutput(os.Stderr)

		for {
			line, err := buf.ReadBytes('\n')
			if err != nil {
				assert.Equal(t, io.EOF, err)
				break
			}
			assert.Empty(t, string(line), "log not empty")
		}
	}
}

func TestSocketListener_tcp_tls(t *testing.T) {
	defer testEmptyLog(t)()

	sl := newSocketListener()
	sl.Log = testutil.Logger{}
	sl.ServiceAddress = "tcp://127.0.0.1:0"
	sl.ServerConfig = *pki.TLSServerConfig()

	acc := &testutil.Accumulator{}
	err := sl.Start(acc)
	require.NoError(t, err)
	defer sl.Stop()

	tlsCfg, err := pki.TLSClientConfig().TLSConfig()
	require.NoError(t, err)

	secureClient, err := tls.Dial("tcp", sl.Closer.(net.Listener).Addr().String(), tlsCfg)
	require.NoError(t, err)

	testSocketListener(t, sl, secureClient)
}

func TestSocketListener_unix_tls(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "sl.TestSocketListener_unix_tls.sock")

	sl := newSocketListener()
	sl.Log = testutil.Logger{}
	sl.ServiceAddress = "unix://" + sock
	sl.ServerConfig = *pki.TLSServerConfig()

	acc := &testutil.Accumulator{}
	err = sl.Start(acc)
	require.NoError(t, err)
	defer sl.Stop()

	tlsCfg, err := pki.TLSClientConfig().TLSConfig()
	tlsCfg.InsecureSkipVerify = true
	require.NoError(t, err)

	secureClient, err := tls.Dial("unix", sock, tlsCfg)
	require.NoError(t, err)

	testSocketListener(t, sl, secureClient)
}

func TestSocketListener_tcp(t *testing.T) {
	defer testEmptyLog(t)()

	sl := newSocketListener()
	sl.Log = testutil.Logger{}
	sl.ServiceAddress = "tcp://127.0.0.1:0"
	sl.ReadBufferSize = config.Size(1024)

	acc := &testutil.Accumulator{}
	err := sl.Start(acc)
	require.NoError(t, err)
	defer sl.Stop()

	client, err := net.Dial("tcp", sl.Closer.(net.Listener).Addr().String())
	require.NoError(t, err)

	testSocketListener(t, sl, client)
}

func TestSocketListener_udp(t *testing.T) {
	defer testEmptyLog(t)()

	sl := newSocketListener()
	sl.Log = testutil.Logger{}
	sl.ServiceAddress = "udp://127.0.0.1:0"
	sl.ReadBufferSize = config.Size(1024)

	acc := &testutil.Accumulator{}
	err := sl.Start(acc)
	require.NoError(t, err)
	defer sl.Stop()

	client, err := net.Dial("udp", sl.Closer.(net.PacketConn).LocalAddr().String())
	require.NoError(t, err)

	testSocketListener(t, sl, client)
}

func TestSocketListener_unix(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "sl.TestSocketListener_unix.sock")

	defer testEmptyLog(t)()

	f, _ := os.Create(sock)
	require.NoError(t, f.Close())
	sl := newSocketListener()
	sl.Log = testutil.Logger{}
	sl.ServiceAddress = "unix://" + sock
	sl.ReadBufferSize = config.Size(1024)

	acc := &testutil.Accumulator{}
	err = sl.Start(acc)
	require.NoError(t, err)
	defer sl.Stop()

	client, err := net.Dial("unix", sock)
	require.NoError(t, err)

	testSocketListener(t, sl, client)
}

func TestSocketListener_unixgram(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows, as unixgram sockets are not supported")
	}

	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "sl.TestSocketListener_unixgram.sock")

	defer testEmptyLog(t)()

	_, err = os.Create(sock)
	require.NoError(t, err)
	sl := newSocketListener()
	sl.Log = testutil.Logger{}
	sl.ServiceAddress = "unixgram://" + sock
	sl.ReadBufferSize = config.Size(1024)

	acc := &testutil.Accumulator{}
	err = sl.Start(acc)
	require.NoError(t, err)
	defer sl.Stop()

	client, err := net.Dial("unixgram", sock)
	require.NoError(t, err)

	testSocketListener(t, sl, client)
}

func TestSocketListenerDecode_tcp(t *testing.T) {
	defer testEmptyLog(t)()

	sl := newSocketListener()
	sl.Log = testutil.Logger{}
	sl.ServiceAddress = "tcp://127.0.0.1:0"
	sl.ReadBufferSize = config.Size(1024)
	sl.ContentEncoding = "gzip"

	acc := &testutil.Accumulator{}
	err := sl.Start(acc)
	require.NoError(t, err)
	defer sl.Stop()

	client, err := net.Dial("tcp", sl.Closer.(net.Listener).Addr().String())
	require.NoError(t, err)

	testSocketListener(t, sl, client)
}

func TestSocketListenerDecode_udp(t *testing.T) {
	defer testEmptyLog(t)()

	sl := newSocketListener()
	sl.Log = testutil.Logger{}
	sl.ServiceAddress = "udp://127.0.0.1:0"
	sl.ReadBufferSize = config.Size(1024)
	sl.ContentEncoding = "gzip"

	acc := &testutil.Accumulator{}
	err := sl.Start(acc)
	require.NoError(t, err)
	defer sl.Stop()

	client, err := net.Dial("udp", sl.Closer.(net.PacketConn).LocalAddr().String())
	require.NoError(t, err)

	testSocketListener(t, sl, client)
}

func testSocketListener(t *testing.T, sl *SocketListener, client net.Conn) {
	mstr12 := []byte("test,foo=bar v=1i 123456789\ntest,foo=baz v=2i 123456790\n")
	mstr3 := []byte("test,foo=zab v=3i 123456791\n")

	if sl.ContentEncoding == "gzip" {
		encoder, err := internal.NewContentEncoder(sl.ContentEncoding)
		require.NoError(t, err)
		mstr12, err = encoder.Encode(mstr12)
		require.NoError(t, err)

		encoder, err = internal.NewContentEncoder(sl.ContentEncoding)
		require.NoError(t, err)
		mstr3, err = encoder.Encode(mstr3)
		require.NoError(t, err)
	}

	_, err := client.Write(mstr12)
	require.NoError(t, err)
	_, err = client.Write(mstr3)
	require.NoError(t, err)
	acc := sl.Accumulator.(*testutil.Accumulator)

	acc.Wait(3)
	acc.Lock()
	m1 := acc.Metrics[0]
	m2 := acc.Metrics[1]
	m3 := acc.Metrics[2]
	acc.Unlock()

	assert.Equal(t, "test", m1.Measurement)
	assert.Equal(t, map[string]string{"foo": "bar"}, m1.Tags)
	assert.Equal(t, map[string]interface{}{"v": int64(1)}, m1.Fields)
	assert.True(t, time.Unix(0, 123456789).Equal(m1.Time))

	assert.Equal(t, "test", m2.Measurement)
	assert.Equal(t, map[string]string{"foo": "baz"}, m2.Tags)
	assert.Equal(t, map[string]interface{}{"v": int64(2)}, m2.Fields)
	assert.True(t, time.Unix(0, 123456790).Equal(m2.Time))

	assert.Equal(t, "test", m3.Measurement)
	assert.Equal(t, map[string]string{"foo": "zab"}, m3.Tags)
	assert.Equal(t, map[string]interface{}{"v": int64(3)}, m3.Fields)
	assert.True(t, time.Unix(0, 123456791).Equal(m3.Time))
}
