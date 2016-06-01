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
	var wg2 sync.WaitGroup
	wg.Add(1)
	wg2.Add(1)
	go UDPServer(t, &wg, &wg2)
	wg2.Wait()

	i := Graylog{
		Servers: []string{"127.0.0.1:12201"},
	}
	i.Connect()

	metrics := testutil.MockMetrics()

	i.Write(metrics)

	wg.Wait()
	i.Close()
}

type GelfObject map[string]interface{}

func UDPServer(t *testing.T, wg *sync.WaitGroup, wg2 *sync.WaitGroup) {
	serverAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12201")
	udpServer, _ := net.ListenUDP("udp", serverAddr)
	defer wg.Done()

	bufR := make([]byte, 1024)
	wg2.Done()
	n, _, _ := udpServer.ReadFromUDP(bufR)

	b := bytes.NewReader(bufR[0:n])
	r, _ := zlib.NewReader(b)

	bufW := bytes.NewBuffer(nil)
	io.Copy(bufW, r)
	r.Close()

	var obj GelfObject
	json.Unmarshal(bufW.Bytes(), &obj)
	assert.Equal(t, obj["_value"], float64(1))
}
