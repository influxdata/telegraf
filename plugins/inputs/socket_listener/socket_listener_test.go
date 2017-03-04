package socket_listener

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSocketListener_tcp(t *testing.T) {
	sl := newSocketListener()
	sl.ServiceAddress = "tcp://127.0.0.1:0"

	acc := &testutil.Accumulator{}
	err := sl.Start(acc)
	require.NoError(t, err)

	client, err := net.Dial("tcp", sl.Closer.(net.Listener).Addr().String())
	require.NoError(t, err)

	testSocketListener(t, sl, client)
}

func TestSocketListener_udp(t *testing.T) {
	sl := newSocketListener()
	sl.ServiceAddress = "udp://127.0.0.1:0"

	acc := &testutil.Accumulator{}
	err := sl.Start(acc)
	require.NoError(t, err)

	client, err := net.Dial("udp", sl.Closer.(net.PacketConn).LocalAddr().String())
	require.NoError(t, err)

	testSocketListener(t, sl, client)
}

func TestSocketListener_unix(t *testing.T) {
	defer os.Remove("/tmp/telegraf_test.sock")
	sl := newSocketListener()
	sl.ServiceAddress = "unix:///tmp/telegraf_test.sock"

	acc := &testutil.Accumulator{}
	err := sl.Start(acc)
	require.NoError(t, err)

	client, err := net.Dial("unix", "/tmp/telegraf_test.sock")
	require.NoError(t, err)

	testSocketListener(t, sl, client)
}

func TestSocketListener_unixgram(t *testing.T) {
	defer os.Remove("/tmp/telegraf_test.sock")
	sl := newSocketListener()
	sl.ServiceAddress = "unixgram:///tmp/telegraf_test.sock"

	acc := &testutil.Accumulator{}
	err := sl.Start(acc)
	require.NoError(t, err)

	client, err := net.Dial("unixgram", "/tmp/telegraf_test.sock")
	require.NoError(t, err)

	testSocketListener(t, sl, client)
}

func testSocketListener(t *testing.T, sl *SocketListener, client net.Conn) {
	mstr12 := "test,foo=bar v=1i 123456789\ntest,foo=baz v=2i 123456790\n"
	mstr3 := "test,foo=zab v=3i 123456791"
	client.Write([]byte(mstr12))
	client.Write([]byte(mstr3))
	if _, ok := client.(net.Conn); ok {
		// stream connection. needs trailing newline to terminate mstr3
		client.Write([]byte{'\n'})
	}

	acc := sl.Accumulator.(*testutil.Accumulator)

	acc.Lock()
	if len(acc.Metrics) < 1 {
		acc.Wait()
	}
	require.True(t, len(acc.Metrics) >= 1)
	m := acc.Metrics[0]
	acc.Unlock()

	assert.Equal(t, "test", m.Measurement)
	assert.Equal(t, map[string]string{"foo": "bar"}, m.Tags)
	assert.Equal(t, map[string]interface{}{"v": int64(1)}, m.Fields)
	assert.True(t, time.Unix(0, 123456789).Equal(m.Time))

	acc.Lock()
	if len(acc.Metrics) < 2 {
		acc.Wait()
	}
	require.True(t, len(acc.Metrics) >= 2)
	m = acc.Metrics[1]
	acc.Unlock()

	assert.Equal(t, "test", m.Measurement)
	assert.Equal(t, map[string]string{"foo": "baz"}, m.Tags)
	assert.Equal(t, map[string]interface{}{"v": int64(2)}, m.Fields)
	assert.True(t, time.Unix(0, 123456790).Equal(m.Time))

	acc.Lock()
	if len(acc.Metrics) < 3 {
		acc.Wait()
	}
	require.True(t, len(acc.Metrics) >= 3)
	m = acc.Metrics[2]
	acc.Unlock()

	assert.Equal(t, "test", m.Measurement)
	assert.Equal(t, map[string]string{"foo": "zab"}, m.Tags)
	assert.Equal(t, map[string]interface{}{"v": int64(3)}, m.Fields)
	assert.True(t, time.Unix(0, 123456791).Equal(m.Time))
}
