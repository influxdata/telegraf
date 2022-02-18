package whatap

import (
	"net"
	"testing"

	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func newWhatap(addr string) *Whatap {
	servers := make([]string, 0)
	servers = append(servers, "tcp://"+addr)
	w := &Whatap{
		Timeout: config.Duration(60 * time.Second),
		Log:     testutil.Logger{},
		Servers: servers,
	}
	return w
}
func TestWhatapConnect(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	w := newWhatap(listener.Addr().String())
	err = w.Init()
	require.NoError(t, err)

	err = w.Connect()
	require.NoError(t, err)

	_, err = listener.Accept()
	require.NoError(t, err)
}

func TestWhatapWriteErr(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	w := newWhatap(listener.Addr().String())
	err = w.Init()
	require.NoError(t, err)

	err = w.Connect()
	require.NoError(t, err)

	lconn, err := listener.Accept()
	require.NoError(t, err)
	err = lconn.(*net.TCPConn).SetWriteBuffer(256)
	require.NoError(t, err)

	metrics := []telegraf.Metric{testutil.TestMetric(1, "testerr")}

	err = lconn.Close()
	require.NoError(t, err)

	err = w.Close()
	require.NoError(t, err)

	err = w.Write(metrics)
	require.NoError(t, err)
}
