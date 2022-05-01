package net_response

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"

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
	err := c.Gather(&acc)
	require.Error(t, err)
	require.Equal(t, "bad protocol", err.Error())
}

func TestNoPort(t *testing.T) {
	var acc testutil.Accumulator
	c := NetResponse{
		Protocol: "tcp",
		Address:  ":",
	}
	err := c.Gather(&acc)
	require.Error(t, err)
	require.Equal(t, "bad port", err.Error())
}

func TestAddressOnly(t *testing.T) {
	var acc testutil.Accumulator
	c := NetResponse{
		Protocol: "tcp",
		Address:  "127.0.0.1",
	}
	err := c.Gather(&acc)
	require.Error(t, err)
	require.Equal(t, "address 127.0.0.1: missing port in address", err.Error())
}

func TestSendExpectStrings(t *testing.T) {
	var acc testutil.Accumulator
	tc := NetResponse{
		Protocol: "udp",
		Address:  "127.0.0.1:7",
		Send:     "",
		Expect:   "toast",
	}
	uc := NetResponse{
		Protocol: "udp",
		Address:  "127.0.0.1:7",
		Send:     "toast",
		Expect:   "",
	}
	err := tc.Gather(&acc)
	require.Error(t, err)
	require.Equal(t, "send string cannot be empty", err.Error())
	err = uc.Gather(&acc)
	require.Error(t, err)
	require.Equal(t, "expected string cannot be empty", err.Error())
}

func TestTCPError(t *testing.T) {
	var acc testutil.Accumulator
	// Init plugin
	c := NetResponse{
		Protocol: "tcp",
		Address:  ":9999",
		Timeout:  config.Duration(time.Second * 30),
	}
	// Gather
	require.NoError(t, c.Gather(&acc))
	acc.AssertContainsTaggedFields(t,
		"net_response",
		map[string]interface{}{
			"result_code": uint64(2),
			"result_type": "connection_failed",
		},
		map[string]string{
			"server":   "",
			"port":     "9999",
			"protocol": "tcp",
			"result":   "connection_failed",
		},
	)
}

func TestTCPOK1(t *testing.T) {
	var wg sync.WaitGroup
	var acc testutil.Accumulator
	// Init plugin
	c := NetResponse{
		Address:     "127.0.0.1:2004",
		Send:        "test",
		Expect:      "test",
		ReadTimeout: config.Duration(time.Second * 3),
		Timeout:     config.Duration(time.Second),
		Protocol:    "tcp",
	}
	// Start TCP server
	wg.Add(1)
	go TCPServer(t, &wg)
	wg.Wait() // Wait for the server to spin up
	wg.Add(1)
	// Connect
	require.NoError(t, c.Gather(&acc))
	acc.Wait(1)

	// Override response time
	for _, p := range acc.Metrics {
		p.Fields["response_time"] = 1.0
	}
	acc.AssertContainsTaggedFields(t,
		"net_response",
		map[string]interface{}{
			"result_code":   uint64(0),
			"result_type":   "success",
			"string_found":  true,
			"response_time": 1.0,
		},
		map[string]string{
			"result":   "success",
			"server":   "127.0.0.1",
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
		ReadTimeout: config.Duration(time.Second * 3),
		Timeout:     config.Duration(time.Second),
		Protocol:    "tcp",
	}
	// Start TCP server
	wg.Add(1)
	go TCPServer(t, &wg)
	wg.Wait()
	wg.Add(1)

	// Connect
	require.NoError(t, c.Gather(&acc))
	acc.Wait(1)

	// Override response time
	for _, p := range acc.Metrics {
		p.Fields["response_time"] = 1.0
	}
	acc.AssertContainsTaggedFields(t,
		"net_response",
		map[string]interface{}{
			"result_code":   uint64(4),
			"result_type":   "string_mismatch",
			"string_found":  false,
			"response_time": 1.0,
		},
		map[string]string{
			"result":   "string_mismatch",
			"server":   "127.0.0.1",
			"port":     "2004",
			"protocol": "tcp",
		},
	)
	// Waiting TCPserver
	wg.Wait()
}

func TestUDPError(t *testing.T) {
	var acc testutil.Accumulator
	// Init plugin
	c := NetResponse{
		Address:  ":9999",
		Send:     "test",
		Expect:   "test",
		Protocol: "udp",
	}
	// Gather
	require.NoError(t, c.Gather(&acc))
	acc.Wait(1)

	// Override response time
	for _, p := range acc.Metrics {
		p.Fields["response_time"] = 1.0
	}
	// Error
	acc.AssertContainsTaggedFields(t,
		"net_response",
		map[string]interface{}{
			"result_code":   uint64(3),
			"result_type":   "read_failed",
			"response_time": 1.0,
			"string_found":  false,
		},
		map[string]string{
			"result":   "read_failed",
			"server":   "",
			"port":     "9999",
			"protocol": "udp",
		},
	)
}

func TestUDPOK1(t *testing.T) {
	var wg sync.WaitGroup
	var acc testutil.Accumulator
	// Init plugin
	c := NetResponse{
		Address:     "127.0.0.1:2004",
		Send:        "test",
		Expect:      "test",
		ReadTimeout: config.Duration(time.Second * 3),
		Timeout:     config.Duration(time.Second),
		Protocol:    "udp",
	}
	// Start UDP server
	wg.Add(1)
	go UDPServer(t, &wg)
	wg.Wait()
	wg.Add(1)

	// Connect
	require.NoError(t, c.Gather(&acc))
	acc.Wait(1)

	// Override response time
	for _, p := range acc.Metrics {
		p.Fields["response_time"] = 1.0
	}
	acc.AssertContainsTaggedFields(t,
		"net_response",
		map[string]interface{}{
			"result_code":   uint64(0),
			"result_type":   "success",
			"string_found":  true,
			"response_time": 1.0,
		},
		map[string]string{
			"result":   "success",
			"server":   "127.0.0.1",
			"port":     "2004",
			"protocol": "udp",
		},
	)
	// Waiting TCPserver
	wg.Wait()
}

func UDPServer(t *testing.T, wg *sync.WaitGroup) {
	defer wg.Done()
	udpAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:2004")
	conn, _ := net.ListenUDP("udp", udpAddr)
	wg.Done()
	buf := make([]byte, 1024)
	_, remoteaddr, _ := conn.ReadFromUDP(buf)
	_, err := conn.WriteToUDP(buf, remoteaddr)
	require.NoError(t, err)
	require.NoError(t, conn.Close())
}

func TCPServer(t *testing.T, wg *sync.WaitGroup) {
	defer wg.Done()
	tcpAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:2004")
	tcpServer, _ := net.ListenTCP("tcp", tcpAddr)
	wg.Done()
	conn, _ := tcpServer.AcceptTCP()
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	require.NoError(t, err)
	_, err = conn.Write(buf)
	require.NoError(t, err)
	require.NoError(t, conn.CloseWrite())
	require.NoError(t, tcpServer.Close())
}
