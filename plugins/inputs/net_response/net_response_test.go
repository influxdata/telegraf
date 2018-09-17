package net_response

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSample(t *testing.T) {
	c := &NetResponse{}
	output := c.SampleConfig()
	if output != sampleConfig {
		t.Error("Sample config doesn't match")
	}
}

func TestDescription(t *testing.T) {
	c := &NetResponse{}
	output := c.Description()
	if output != description {
		t.Error("Description output is not correct")
	}
}
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

func TestNoPort(t *testing.T) {
	var acc testutil.Accumulator
	c := NetResponse{
		Protocol: "tcp",
		Address:  ":",
	}
	err1 := c.Gather(&acc)
	require.Error(t, err1)
	assert.Equal(t, "Bad port", err1.Error())
}

func TestAddressOnly(t *testing.T) {
	var acc testutil.Accumulator
	c := NetResponse{
		Protocol: "tcp",
		Address:  "127.0.0.1",
	}
	err1 := c.Gather(&acc)
	require.Error(t, err1)
	assert.Equal(t, "address 127.0.0.1: missing port in address", err1.Error())
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
	err1 := tc.Gather(&acc)
	require.Error(t, err1)
	assert.Equal(t, "Send string cannot be empty", err1.Error())
	err2 := uc.Gather(&acc)
	require.Error(t, err2)
	assert.Equal(t, "Expected string cannot be empty", err2.Error())
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
	require.NoError(t, err1)
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
	err1 := c.Gather(&acc)
	// Override response time
	for _, p := range acc.Metrics {
		p.Fields["response_time"] = 1.0
	}
	// Error
	require.NoError(t, err1)
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
