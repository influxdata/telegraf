package whatap

import (
	"fmt"
	//"log"
	"net"
	"os"
	"testing"

	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"

	whatap_hash "github.com/whatap/go-api/common/util/hash"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newWhatap() *Whatap {
	hostname, _ := os.Hostname()
	return &Whatap{
		Timeout: 60 * time.Second,
		Session: TcpSession{},
		Oname:   hostname,
		Oid:     whatap_hash.HashStr(hostname),
	}
}
func TestWhatapConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
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
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
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
