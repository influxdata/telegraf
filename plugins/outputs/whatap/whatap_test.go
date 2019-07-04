package whatap

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"testing"

	//"time"
	//"github.com/influxdata/telegraf"
	//"github.com/influxdata/telegraf/testutil"

	//"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	License = "x2tggtnopk2t9-z39dt59pe1pmjc-xipbnkb0ph6bn"
	Server  = "121.166.140.134"
)

func TestConnect(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	w := newWhatap()
	addr := listener.Addr().String()
	fmt.Println(addr)

	arr := strings.Split(addr, ":")
	w.Server = arr[0]
	w.Port, err = strconv.Atoi(arr[1])
	require.NoError(t, err)

	err = w.Connect()
	require.NoError(t, err)

	_, err = listener.Accept()
	require.NoError(t, err)
}

func TestAutoOname(t *testing.T) {
	log.Println("WhaTap Test", "TestConnect")
}

func TestWrite(t *testing.T) {
	log.Println("WhaTap Test", "TestConnect")

}
