package graylog

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"io"
	"net"
	"sync"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteDefault(t *testing.T) {
	scenarioUDP(t, "127.0.0.1:12201")
}

func TestWriteUDP(t *testing.T) {
	scenarioUDP(t, "udp://127.0.0.1:12201")
}

func TestWriteTCP(t *testing.T) {
	scenarioTCP(t, "tcp://127.0.0.1:12201")
}

func scenarioUDP(t *testing.T, server string) {
	var wg sync.WaitGroup
	var wg2 sync.WaitGroup
	wg.Add(1)
	wg2.Add(1)
	go UDPServer(t, &wg, &wg2)
	wg2.Wait()

	i := Graylog{
		Servers: []string{server},
	}
	i.Connect()

	metrics := testutil.MockMetrics()

	// UDP scenario:
	// 4 messages are send

	err := i.Write(metrics)
	require.NoError(t, err)
	err = i.Write(metrics)
	require.NoError(t, err)
	err = i.Write(metrics)
	require.NoError(t, err)
	err = i.Write(metrics)
	require.NoError(t, err)

	wg.Wait()
	i.Close()
}

func scenarioTCP(t *testing.T, server string) {
	var wg sync.WaitGroup
	var wg2 sync.WaitGroup
	var wg3 sync.WaitGroup
	wg.Add(1)
	wg2.Add(1)
	wg3.Add(1)
	go TCPServer(t, &wg, &wg2, &wg3)
	wg2.Wait()

	i := Graylog{
		Servers: []string{server},
	}
	i.Connect()

	metrics := testutil.MockMetrics()

	// TCP scenario:
	// 4 messages are send
	// -> connection gets broken after the 2nd message (server closes connection)
	// -> the 3rd write ends with error
	// -> in the 4th write connection is restored and write is successful

	err := i.Write(metrics)
	require.NoError(t, err)
	err = i.Write(metrics)
	require.NoError(t, err)
	wg3.Wait()
	err = i.Write(metrics)
	require.Error(t, err)
	err = i.Write(metrics)
	require.NoError(t, err)

	wg.Wait()
	i.Close()
}

type GelfObject map[string]interface{}

func UDPServer(t *testing.T, wg *sync.WaitGroup, wg2 *sync.WaitGroup) {
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:12201")
	require.NoError(t, err)
	udpServer, err := net.ListenUDP("udp", serverAddr)
	require.NoError(t, err)
	defer udpServer.Close()
	defer wg.Done()

	bufR := make([]byte, 1024)
	wg2.Done()

	recv := func() {
		n, _, err := udpServer.ReadFromUDP(bufR)
		require.NoError(t, err)

		b := bytes.NewReader(bufR[0:n])
		r, _ := zlib.NewReader(b)

		bufW := bytes.NewBuffer(nil)
		io.Copy(bufW, r)
		r.Close()

		var obj GelfObject
		json.Unmarshal(bufW.Bytes(), &obj)
		assert.Equal(t, obj["_value"], float64(1))
	}

	// in UDP scenario all 4 messages are received

	recv()
	recv()
	recv()
	recv()
}

func TCPServer(t *testing.T, wg *sync.WaitGroup, wg2 *sync.WaitGroup, wg3 *sync.WaitGroup) {
	serverAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:12201")
	require.NoError(t, err)
	tcpServer, err := net.ListenTCP("tcp", serverAddr)
	require.NoError(t, err)
	defer tcpServer.Close()
	defer wg.Done()

	bufR := make([]byte, 1)
	bufW := bytes.NewBuffer(nil)
	wg2.Done()

	accept := func() *net.TCPConn {
		conn, err := tcpServer.AcceptTCP()
		conn.SetLinger(0)
		require.NoError(t, err)
		return conn
	}
	conn := accept()
	defer conn.Close()

	recv := func() {
		bufW.Reset()
		for {
			n, err := conn.Read(bufR)
			require.NoError(t, err)
			if n > 0 {
				if bufR[0] == 0 { // message delimiter found
					break
				} else {
					bufW.Write(bufR)
				}
			}
		}

		var obj GelfObject
		json.Unmarshal(bufW.Bytes(), &obj)
		assert.Equal(t, obj["_value"], float64(1))
	}

	// in TCP scenario only 3 messages are received (1st, 2dn and 4th) due to connection break after the 2nd

	recv()
	recv()
	conn.Close()
	wg3.Done()
	conn = accept()
	defer conn.Close()
	recv()
}
