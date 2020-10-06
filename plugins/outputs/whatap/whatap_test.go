package whatap

import (
	"fmt"
	//"log"
	"net"
	"testing"

	//"time"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	License = "x2tggtnopk2t9-z39dt59pe1pmjc-xipbnkb0ph6bn"
	Server  = "121.166.140.134"
)

func TestWhatapConnect(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	w := newWhatap()
	addr := listener.Addr().String()
	fmt.Println(addr)

	w.Servers = append(w.Servers, fmt.Sprintf("%s://%s", "tcp", addr))
	require.NoError(t, err)

	err = w.Connect()
	require.NoError(t, err)

	_, err = listener.Accept()
	require.NoError(t, err)
}

func TestWhatapWriteErr(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	w := newWhatap()
	addr := listener.Addr().String()
	fmt.Println(addr)

	w.Servers = append(w.Servers, fmt.Sprintf("%s://%s", "tcp", addr))
	require.NoError(t, err)

	err = w.Connect()
	require.NoError(t, err)

	lconn, err := listener.Accept()
	require.NoError(t, err)
	lconn.(*net.TCPConn).SetWriteBuffer(256)

	metrics := []telegraf.Metric{testutil.TestMetric(1, "testerr")}

	// close the socket to generate an error
	lconn.Close()
	w.Session.Client.Close()
	err = w.Write(metrics)
	require.Error(t, err)
	assert.Nil(t, w.Session.Client)
}
