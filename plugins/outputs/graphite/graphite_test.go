package graphite

import (
	"bufio"
	"net"
	"net/textproto"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf"

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
	pt1, _ := telegraf.NewMetric(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"mymeasurement": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	// Prepare point list
	var metrics []telegraf.Metric
	metrics = append(metrics, pt1)
	// Error
	err1 := g.Connect()
	require.NoError(t, err1)
	err2 := g.Write(metrics)
	require.Error(t, err2)
	assert.Equal(t, "Could not write to any Graphite server in cluster\n", err2.Error())
}

func TestGraphiteOK(t *testing.T) {
	var wg sync.WaitGroup
	// Init plugin
	g := Graphite{
		Prefix: "my.prefix",
	}
	// Init metrics
	pt1, _ := telegraf.NewMetric(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"mymeasurement": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	pt2, _ := telegraf.NewMetric(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	pt3, _ := telegraf.NewMetric(
		"my_measurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	// Prepare point list
	var metrics []telegraf.Metric
	metrics = append(metrics, pt1)
	metrics = append(metrics, pt2)
	metrics = append(metrics, pt3)
	// Start TCP server
	wg.Add(1)
	go TCPServer(t, &wg)
	wg.Wait()
	// Connect
	wg.Add(1)
	err1 := g.Connect()
	wg.Wait()
	require.NoError(t, err1)
	// Send Data
	err2 := g.Write(metrics)
	require.NoError(t, err2)
	wg.Add(1)
	// Waiting TCPserver
	wg.Wait()
	g.Close()
}

func TCPServer(t *testing.T, wg *sync.WaitGroup) {
	tcpServer, _ := net.Listen("tcp", "127.0.0.1:2003")
	wg.Done()
	conn, _ := tcpServer.Accept()
	wg.Done()
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)
	data1, _ := tp.ReadLine()
	assert.Equal(t, "my.prefix.192_168_0_1.mymeasurement 3.14 1289430000", data1)
	data2, _ := tp.ReadLine()
	assert.Equal(t, "my.prefix.192_168_0_1.mymeasurement.value 3.14 1289430000", data2)
	data3, _ := tp.ReadLine()
	assert.Equal(t, "my.prefix.192_168_0_1.my_measurement.value 3.14 1289430000", data3)
	conn.Close()
	wg.Done()
}
