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
	m1, _ := telegraf.NewMetric(
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
	// Init plugin
	g := Graphite{
		Prefix: "my.prefix",
	}
	// Init metrics
	m1, _ := telegraf.NewMetric(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"mymeasurement": float64(3.14)},
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
	var metrics []telegraf.Metric
	metrics = append(metrics, m1)
	metrics = append(metrics, m2)
	metrics = append(metrics, m3)
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

func TestGraphiteMetricName(t *testing.T) {
	m1, _ := telegraf.NewMetric(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m2, _ := telegraf.NewMetric(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1", "afoo": "first", "bfoo": "second"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m3, _ := telegraf.NewMetric(
		"mymeasurement",
		map[string]string{"afoo": "first", "bfoo": "second"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m4, _ := telegraf.NewMetric(
		"custom1",
		map[string]string{"host": "192.168.0.1", "afoo": "first", "bfoo": "second"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m5, _ := telegraf.NewMetric(
		"custom.2",
		map[string]string{"host": "192.168.0.1", "afoo": "first", "bfoo": "second"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	g := Graphite{
		Prefix: "my.prefix",
		MetricsNameBuilder: map[string][]string{
			"custom1":  []string{"host", "afoo", "bfoo"},
			"custom.2": []string{"{{field}}", "bfoo", "{{metric}}"}},
	}
	metricName1 := g.buildMetricName(m1, "value")
	metricName2 := g.buildMetricName(m2, "value")
	metricName3 := g.buildMetricName(m3, "value")
	metricName4 := g.buildMetricName(m4, "value")
	metricName5 := g.buildMetricName(m5, "value")

	assert.Equal(t, "my.prefix.192_168_0_1.mymeasurement.value", metricName1)
	assert.Equal(t, "my.prefix.192_168_0_1.first.second.mymeasurement.value", metricName2)
	assert.Equal(t, "my.prefix.first.second.mymeasurement.value", metricName3)
	assert.Equal(t, "my.prefix.192_168_0_1.first.second", metricName4)
	assert.Equal(t, "my.prefix.value.second.custom_2", metricName5)

}
