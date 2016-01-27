package connection

import (
	"net"
	"regexp"
	"sync"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTCPError(t *testing.T) {
	var acc testutil.Accumulator
	// Init plugin
	tcp1 := Tcp{
		Address: ":9999",
	}
	c := Connection{
		Tcps: []Tcp{tcp1},
	}
	// Error
	err1 := c.Gather(&acc)
	require.Error(t, err1)
	assert.Equal(t, "dial tcp 127.0.0.1:9999: getsockopt: connection refused", err1.Error())
}

func TestTCPOK1(t *testing.T) {
	var wg sync.WaitGroup
	var acc testutil.Accumulator
	// Init plugin
	tcp1 := Tcp{
		Address:     "127.0.0.1:2004",
		Send:        "test",
		Expect:      "test",
		ReadTimeout: 3.0,
		Timeout:     1.0,
	}
	c := Connection{
		Tcps: []Tcp{tcp1},
	}
	// Start TCP server
	wg.Add(1)
	go TCPServer(t, &wg)
	wg.Wait()
	// Connect
	wg.Add(1)
	err1 := c.Gather(&acc)
	wg.Wait()
	// Override response time
	for _, p := range acc.Points {
		p.Fields["response_time"] = 1.0
	}
	require.NoError(t, err1)
	acc.AssertContainsTaggedFields(t,
		"tcp_connection",
		map[string]interface{}{
			"string_found":  true,
			"response_time": 1.0,
		},
		map[string]string{"server": tcp1.Address},
	)
	// Waiting TCPserver
	wg.Wait()
}

func TestTCPOK2(t *testing.T) {
	var wg sync.WaitGroup
	var acc testutil.Accumulator
	// Init plugin
	tcp1 := Tcp{
		Address:     "127.0.0.1:2004",
		Send:        "test",
		Expect:      "test2",
		ReadTimeout: 3.0,
		Timeout:     1.0,
	}
	c := Connection{
		Tcps: []Tcp{tcp1},
	}
	// Start TCP server
	wg.Add(1)
	go TCPServer(t, &wg)
	wg.Wait()
	// Connect
	wg.Add(1)
	err1 := c.Gather(&acc)
	wg.Wait()
	// Override response time
	for _, p := range acc.Points {
		p.Fields["response_time"] = 1.0
	}
	require.NoError(t, err1)
	acc.AssertContainsTaggedFields(t,
		"tcp_connection",
		map[string]interface{}{
			"string_found":  false,
			"response_time": 1.0,
		},
		map[string]string{"server": tcp1.Address},
	)
	// Waiting TCPserver
	wg.Wait()
}

func TestUDPrror(t *testing.T) {
	var acc testutil.Accumulator
	// Init plugin
	udp1 := Udp{
		Address: ":9999",
		Send:    "test",
		Expect:  "test",
	}
	c := Connection{
		Udps: []Udp{udp1},
	}
	// Error
	err1 := c.Gather(&acc)
	require.Error(t, err1)
	assert.Regexp(t, regexp.MustCompile(`read udp 127.0.0.1:[0-9]*->127.0.0.1:9999: recvfrom: connection refused`), err1.Error())
}

func TestUDPOK1(t *testing.T) {
	var wg sync.WaitGroup
	var acc testutil.Accumulator
	// Init plugin
	udp1 := Udp{
		Address:     "127.0.0.1:2004",
		Send:        "test",
		Expect:      "test",
		ReadTimeout: 3.0,
		Timeout:     1.0,
	}
	c := Connection{
		Udps: []Udp{udp1},
	}
	// Start UDP server
	wg.Add(1)
	go UDPServer(t, &wg)
	wg.Wait()
	// Connect
	wg.Add(1)
	err1 := c.Gather(&acc)
	wg.Wait()
	// Override response time
	for _, p := range acc.Points {
		p.Fields["response_time"] = 1.0
	}
	require.NoError(t, err1)
	acc.AssertContainsTaggedFields(t,
		"udp_connection",
		map[string]interface{}{
			"string_found":  true,
			"response_time": 1.0,
		},
		map[string]string{"server": udp1.Address},
	)
	// Waiting TCPserver
	wg.Wait()
}

func UDPServer(t *testing.T, wg *sync.WaitGroup) {
	udpAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:2004")
	conn, _ := net.ListenUDP("udp", udpAddr)
	wg.Done()
	buf := make([]byte, 1024)
	_, remoteaddr, _ := conn.ReadFromUDP(buf)
	conn.WriteToUDP(buf, remoteaddr)
	conn.Close()
	wg.Done()
}

func TCPServer(t *testing.T, wg *sync.WaitGroup) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:2004")
	tcpServer, _ := net.ListenTCP("tcp", tcpAddr)
	wg.Done()
	conn, _ := tcpServer.AcceptTCP()
	buf := make([]byte, 1024)
	conn.Read(buf)
	conn.Write(buf)
	conn.CloseWrite()
	tcpServer.Close()
	wg.Done()
}
