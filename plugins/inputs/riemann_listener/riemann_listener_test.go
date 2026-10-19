package riemann_listener

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"testing"
	"time"

	riemanngo "github.com/riemann/riemann-go-client"
	riemann_proto "github.com/riemann/riemann-go-client/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

func TestSocketListener_tcp(t *testing.T) {
	log.Println("Entering")

	sl := newRiemannSocketListener()
	sl.Log = testutil.Logger{}
	sl.ServiceAddress = "tcp://127.0.0.1:5555"
	sl.ReadBufferSize = config.Size(1024)

	acc := &testutil.Accumulator{}
	err := sl.Start(acc)
	require.NoError(t, err)
	defer sl.Stop()

	// Check for stats of specific service
	c := riemanngo.NewTCPClient("127.0.0.1:5555", 5*time.Second)
	require.NoError(t, c.Connect())
	require.NoError(t, err)
	defer c.Close()

	result, err := riemanngo.SendEvent(c, &riemanngo.Event{
		Service: "hello",
	})
	require.NoError(t, err)
	require.True(t, result.GetOk())

	// Check for stats no service specified
	c = riemanngo.NewTCPClient("127.0.0.1:5555", 5*time.Second)
	require.NoError(t, c.Connect())
	defer c.Close()

	result, err = riemanngo.SendEvent(c, &riemanngo.Event{})
	require.False(t, result.GetOk())
	require.Equal(t, "No Service Name", result.GetError())
	require.NoError(t, err)
}

func TestMessageTooLarge(t *testing.T) {
	// Find a free port
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := probe.Addr().String()
	require.NoError(t, probe.Close())

	plugin := newRiemannSocketListener()
	plugin.Log = testutil.Logger{}
	plugin.ServiceAddress = "tcp://" + addr
	plugin.MaxMessageSize = config.Size(1024)

	acc := &testutil.Accumulator{}
	require.NoError(t, plugin.Start(acc))
	defer plugin.Stop()

	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer conn.Close()

	// Announce a message above the limit without sending any body
	require.NoError(t, binary.Write(conn, binary.BigEndian, uint32(1025)))

	// Expect the error response and the connection to be closed
	var length uint32
	require.NoError(t, binary.Read(conn, binary.BigEndian, &length))
	buf := make([]byte, length)
	_, err = io.ReadFull(conn, buf)
	require.NoError(t, err)
	var response riemann_proto.Msg
	require.NoError(t, proto.Unmarshal(buf, &response))
	require.False(t, response.GetOk())
	require.Equal(t, "Message too large", response.GetError())
	_, err = conn.Read(buf[:1])
	require.ErrorIs(t, err, io.EOF)
	require.Empty(t, acc.GetTelegrafMetrics())
}
