package tcp_forwarder

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers/influx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTCPForwarderError(t *testing.T) {
	server := "127.0.0.1:8089"
	// Init plugin
	g := TCPForwarder{
		Server:     server,
		serializer: &influx.InfluxSerializer{},
	}
	// Error
	err := g.Connect()
	assert.Equal(
		t,
		fmt.Sprintf("dial tcp %s: getsockopt: connection refused", server),
		err.Error())
}

func TestTCPForwaderOK(t *testing.T) {
	var wg sync.WaitGroup
	// Start TCP server
	wg.Add(1)
	TCPServer(t, &wg)
	// Give the fake TCP server some time to start:
	// Init plugin
	g := TCPForwarder{
		serializer: &influx.InfluxSerializer{},
	}
	// Init metrics
	m1, _ := telegraf.NewMetric(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"myfield": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m2, _ := telegraf.NewMetric(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m3, _ := telegraf.NewMetric(
		"my_measurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	// Prepare point list
	metrics := []telegraf.Metric{m1, m2, m3}
	err1 := g.Connect()
	require.NoError(t, err1)
	// Send Data
	err2 := g.Write(metrics)
	require.NoError(t, err2)
	// Waiting TCPserver
	wg.Wait()
	g.Close()
}

func TCPServer(t *testing.T, wg *sync.WaitGroup) {
	tcpServer, err := net.Listen("tcp", "127.0.0.1:8089")
	if err != nil {
		log.Printf("Couldn't Listen to port 8089: %s\n", err)
		return
	}
	go func() {
		defer wg.Done()
		conn, _ := tcpServer.Accept()
		reader := bufio.NewReader(conn)
		tp := textproto.NewReader(reader)
		data1, _ := tp.ReadLine()
		assert.Equal(t, "mymeasurement,host=192.168.0.1 myfield=3.14 1289430000000000000", data1)
		data2, _ := tp.ReadLine()
		assert.Equal(t, "mymeasurement,host=192.168.0.1 value=3.14 1289430000000000000", data2)
		data3, _ := tp.ReadLine()
		assert.Equal(t, "my_measurement,host=192.168.0.1 value=3.14 1289430000000000000", data3)
		conn.Close()
	}()
}
