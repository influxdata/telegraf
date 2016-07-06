package net_response

import (
	"net"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBadProtocol(t *testing.T) {
	var acc testutil.Accumulator
	// Init plugin
	c := NetResponse{
		Protocol: "unknownprotocol",
		Address:  ":9999",
	}
	// Error
	err1 := c.Gather(&acc)
	require.Error(t, err1)
	assert.Equal(t, "Bad protocol", err1.Error())
}

func TestTCPError(t *testing.T) {
	var acc testutil.Accumulator
	// Init plugin
	c := NetResponse{
		Protocol: "tcp",
		Address:  ":9999",
	}
	// Error
	err1 := c.Gather(&acc)
	require.Error(t, err1)
	assert.Contains(t, err1.Error(), "getsockopt: connection refused")
}

func TestTCPOK1(t *testing.T) {
	var wg sync.WaitGroup
	var acc testutil.Accumulator
	// Init plugin
	c := NetResponse{
		Address:     "127.0.0.1:2004",
		Send:        "test",
		Expect:      "test",
		ReadTimeout: internal.Duration{Duration: time.Second * 3},
		Timeout:     internal.Duration{Duration: time.Second},
		Protocol:    "tcp",
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
	for _, p := range acc.Metrics {
		p.Fields["response_time"] = 1.0
	}
	require.NoError(t, err1)
	acc.AssertContainsTaggedFields(t,
		"net_response",
		map[string]interface{}{
			"string_found":  true,
			"response_time": 1.0,
		},
		map[string]string{"server": "127.0.0.1",
			"port":     "2004",
			"protocol": "tcp",
		},
	)
	// Waiting TCPserver
	wg.Wait()
}

func TestTCPOK2(t *testing.T) {
	var wg sync.WaitGroup
	var acc testutil.Accumulator
	// Init plugin
	c := NetResponse{
		Address:     "127.0.0.1:2004",
		Send:        "test",
		Expect:      "test2",
		ReadTimeout: internal.Duration{Duration: time.Second * 3},
		Timeout:     internal.Duration{Duration: time.Second},
		Protocol:    "tcp",
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
	for _, p := range acc.Metrics {
		p.Fields["response_time"] = 1.0
	}
	require.NoError(t, err1)
	acc.AssertContainsTaggedFields(t,
		"net_response",
		map[string]interface{}{
			"string_found":  false,
			"response_time": 1.0,
		},
		map[string]string{"server": "127.0.0.1",
			"port":     "2004",
			"protocol": "tcp",
		},
	)
	// Waiting TCPserver
	wg.Wait()
}

func TestUDPrror(t *testing.T) {
	var acc testutil.Accumulator
	// Init plugin
	c := NetResponse{
		Address:  ":9999",
		Send:     "test",
		Expect:   "test",
		Protocol: "udp",
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
	c := NetResponse{
		Address:     "127.0.0.1:2004",
		Send:        "test",
		Expect:      "test",
		ReadTimeout: internal.Duration{Duration: time.Second * 3},
		Timeout:     internal.Duration{Duration: time.Second},
		Protocol:    "udp",
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
	for _, p := range acc.Metrics {
		p.Fields["response_time"] = 1.0
	}
	require.NoError(t, err1)
	acc.AssertContainsTaggedFields(t,
		"net_response",
		map[string]interface{}{
			"string_found":  true,
			"response_time": 1.0,
		},
		map[string]string{"server": "127.0.0.1",
			"port":     "2004",
			"protocol": "udp",
		},
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
