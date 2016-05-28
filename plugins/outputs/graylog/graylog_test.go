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
)

func TestWrite(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	go UDPServer(t, &wg)

	i := Graylog{
		Servers: []string{"127.0.0.1:12201"},
	}
	i.Connect()

	metrics := testutil.MockMetrics()
	metrics = append(metrics, testutil.TestMetric(int64(1234567890)))

	i.Write(metrics)
}

type GelfObject map[string]interface{}

func UDPServer(t *testing.T, wg *sync.WaitGroup) {
	serverAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12201")
	udpServer, _ := net.ListenUDP("udp", serverAddr)
	defer wg.Done()

	bufR := make([]byte, 1024)
	n, _, _ := udpServer.ReadFromUDP(bufR)

	b := bytes.NewReader(bufR[0:n])
	r, _ := zlib.NewReader(b)

	bufW := bytes.NewBuffer(nil)
	io.Copy(bufW, r)

	var obj GelfObject
	json.Unmarshal(bufW.Bytes()[12:], &obj)
	assert.Equal(t, bufW.Bytes()[0], 0x1e)
	assert.Equal(t, bufW.Bytes()[0], 0x0f)
	assert.Equal(t, obj["value"], 1)
	r.Close()
}
