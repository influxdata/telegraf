package graphite

import (
	"bufio"
	"net"
	"net/textproto"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphiteError(t *testing.T) {
	// Init plugin
	g := Graphite{
		Servers: []string{"127.0.0.1:2003", "127.0.0.1:12003"},
		Prefix:  "my.prefix",
	}
	// Init metrics
	m1, _ := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"mymeasurement": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	// Prepare point list
	var metrics []telegraf.Metric
	metrics = append(metrics, m1)
	// Error
	err1 := g.Connect()
	require.NoError(t, err1)
	err2 := g.Write(metrics)
	require.Error(t, err2)
	assert.Equal(t, "Could not write to any Graphite server in cluster\n", err2.Error())
}

func TestGraphiteOK(t *testing.T) {
	var wg sync.WaitGroup
	// Start TCP server
	wg.Add(1)
	go TCPServer(t, &wg)
	// Give the fake graphite TCP server some time to start:
	time.Sleep(time.Millisecond * 100)

	// Init plugin
	g := Graphite{
		Prefix: "my.prefix",
	}
	// Init metrics
	m1, _ := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"myfield": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m2, _ := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m3, _ := metric.New(
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
	tcpServer, _ := net.Listen("tcp", "127.0.0.1:2003")
	defer wg.Done()
	conn, _ := tcpServer.Accept()
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)
	data1, _ := tp.ReadLine()
	assert.Equal(t, "my.prefix.192_168_0_1.mymeasurement.myfield 3.14 1289430000", data1)
	data2, _ := tp.ReadLine()
	assert.Equal(t, "my.prefix.192_168_0_1.mymeasurement 3.14 1289430000", data2)
	data3, _ := tp.ReadLine()
	assert.Equal(t, "my.prefix.192_168_0_1.my_measurement 3.14 1289430000", data3)
	conn.Close()
}
