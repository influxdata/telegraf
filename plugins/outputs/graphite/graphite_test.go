package graphite

import (
	"bufio"
	"net"
	"net/textproto"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestGraphiteError(t *testing.T) {
	// Init plugin
	g := Graphite{
		Servers: []string{"127.0.0.1:12004", "127.0.0.1:12003"},
		Prefix:  "my.prefix",
		Log:     testutil.Logger{},
	}
	// Init metrics
	m1 := metric.New(
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
	require.Equal(t, "could not write to any Graphite server in cluster", err2.Error())
}

func TestGraphiteOK(t *testing.T) {
	var wg sync.WaitGroup
	// Start TCP server
	wg.Add(1)
	t.Log("Starting server")
	TCPServer1(t, &wg)

	// Init plugin
	g := Graphite{
		Prefix:  "my.prefix",
		Servers: []string{"localhost:12003"},
		Log:     testutil.Logger{},
	}

	// Init metrics
	m1 := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"myfield": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m2 := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m3 := metric.New(
		"my_measurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	// Prepare point list
	metrics := []telegraf.Metric{m1}
	metrics2 := []telegraf.Metric{m2, m3}
	err1 := g.Connect()
	require.NoError(t, err1)
	// Send Data
	t.Log("Send first data")
	err2 := g.Write(metrics)
	require.NoError(t, err2)

	// Waiting TCPserver, should reconnect and resend
	wg.Wait()
	t.Log("Finished Waiting for first data")
	var wg2 sync.WaitGroup
	// Start TCP server
	wg2.Add(1)
	TCPServer2(t, &wg2)
	//Write but expect an error, but reconnect
	err3 := g.Write(metrics2)
	t.Log("Finished writing second data, it should have reconnected automatically")

	require.NoError(t, err3)
	t.Log("Finished writing third data")
	wg2.Wait()
	g.Close()
}

func TestGraphiteOkWithSeparatorDot(t *testing.T) {
	var wg sync.WaitGroup
	// Start TCP server
	wg.Add(1)
	t.Log("Starting server")
	TCPServer1(t, &wg)

	// Init plugin
	g := Graphite{
		Prefix:            "my.prefix",
		GraphiteSeparator: ".",
		Servers:           []string{"localhost:12003"},
		Log:               testutil.Logger{},
	}

	// Init metrics
	m1 := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"myfield": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m2 := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m3 := metric.New(
		"my_measurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	// Prepare point list
	metrics := []telegraf.Metric{m1}
	metrics2 := []telegraf.Metric{m2, m3}
	err1 := g.Connect()
	require.NoError(t, err1)
	// Send Data
	t.Log("Send first data")
	err2 := g.Write(metrics)
	require.NoError(t, err2)

	// Waiting TCPserver, should reconnect and resend
	wg.Wait()
	t.Log("Finished Waiting for first data")
	var wg2 sync.WaitGroup
	// Start TCP server
	wg2.Add(1)
	TCPServer2(t, &wg2)
	//Write but expect an error, but reconnect
	err3 := g.Write(metrics2)
	t.Log("Finished writing second data, it should have reconnected automatically")

	require.NoError(t, err3)
	t.Log("Finished writing third data")
	wg2.Wait()
	g.Close()
}

func TestGraphiteOkWithSeparatorUnderscore(t *testing.T) {
	var wg sync.WaitGroup
	// Start TCP server
	wg.Add(1)
	t.Log("Starting server")
	TCPServer1(t, &wg)

	// Init plugin
	g := Graphite{
		Prefix:            "my.prefix",
		GraphiteSeparator: "_",
		Servers:           []string{"localhost:12003"},
		Log:               testutil.Logger{},
	}

	// Init metrics
	m1 := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"myfield": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m2 := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m3 := metric.New(
		"my_measurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	// Prepare point list
	metrics := []telegraf.Metric{m1}
	metrics2 := []telegraf.Metric{m2, m3}
	err1 := g.Connect()
	require.NoError(t, err1)
	// Send Data
	t.Log("Send first data")
	err2 := g.Write(metrics)
	require.NoError(t, err2)

	// Waiting TCPserver, should reconnect and resend
	wg.Wait()
	t.Log("Finished Waiting for first data")
	var wg2 sync.WaitGroup
	// Start TCP server
	wg2.Add(1)
	TCPServer2(t, &wg2)
	//Write but expect an error, but reconnect
	err3 := g.Write(metrics2)
	t.Log("Finished writing second data, it should have reconnected automatically")

	require.NoError(t, err3)
	t.Log("Finished writing third data")
	wg2.Wait()
	g.Close()
}

func TestGraphiteOKWithMultipleTemplates(t *testing.T) {
	var wg sync.WaitGroup
	// Start TCP server
	wg.Add(1)
	t.Log("Starting server")
	TCPServer1WithMultipleTemplates(t, &wg)

	// Init plugin
	g := Graphite{
		Prefix:   "my.prefix",
		Template: "measurement.host.tags.field",
		Templates: []string{
			"my_* host.measurement.tags.field",
			"measurement.tags.host.field",
		},
		Servers: []string{"localhost:12003"},
		Log:     testutil.Logger{},
	}

	// Init metrics
	m1 := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1", "mytag": "valuetag"},
		map[string]interface{}{"myfield": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m2 := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1", "mytag": "valuetag"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m3 := metric.New(
		"my_measurement",
		map[string]string{"host": "192.168.0.1", "mytag": "valuetag"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	// Prepare point list
	metrics := []telegraf.Metric{m1}
	metrics2 := []telegraf.Metric{m2, m3}
	err1 := g.Connect()
	require.NoError(t, err1)
	// Send Data
	t.Log("Send first data")
	err2 := g.Write(metrics)
	require.NoError(t, err2)

	// Waiting TCPserver, should reconnect and resend
	wg.Wait()
	t.Log("Finished Waiting for first data")
	var wg2 sync.WaitGroup
	// Start TCP server
	wg2.Add(1)
	TCPServer2WithMultipleTemplates(t, &wg2)
	//Write but expect an error, but reconnect
	err3 := g.Write(metrics2)
	t.Log("Finished writing second data, it should have reconnected automatically")

	require.NoError(t, err3)
	t.Log("Finished writing third data")
	wg2.Wait()
	g.Close()
}

func TestGraphiteOkWithTags(t *testing.T) {
	var wg sync.WaitGroup
	// Start TCP server
	wg.Add(1)
	t.Log("Starting server")
	TCPServer1WithTags(t, &wg)

	// Init plugin
	g := Graphite{
		Prefix:             "my.prefix",
		GraphiteTagSupport: true,
		Servers:            []string{"localhost:12003"},
		Log:                testutil.Logger{},
	}

	// Init metrics
	m1 := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"myfield": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m2 := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m3 := metric.New(
		"my_measurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	// Prepare point list
	metrics := []telegraf.Metric{m1}
	metrics2 := []telegraf.Metric{m2, m3}
	err1 := g.Connect()
	require.NoError(t, err1)
	// Send Data
	t.Log("Send first data")
	err2 := g.Write(metrics)
	require.NoError(t, err2)

	// Waiting TCPserver, should reconnect and resend
	wg.Wait()
	t.Log("Finished Waiting for first data")
	var wg2 sync.WaitGroup
	// Start TCP server
	wg2.Add(1)
	TCPServer2WithTags(t, &wg2)
	//Write but expect an error, but reconnect
	err3 := g.Write(metrics2)
	t.Log("Finished writing second data, it should have reconnected automatically")

	require.NoError(t, err3)
	t.Log("Finished writing third data")
	wg2.Wait()
	g.Close()
}

func TestGraphiteOkWithTagsAndSeparatorDot(t *testing.T) {
	var wg sync.WaitGroup
	// Start TCP server
	wg.Add(1)
	t.Log("Starting server")
	TCPServer1WithTags(t, &wg)

	// Init plugin
	g := Graphite{
		Prefix:             "my.prefix",
		GraphiteTagSupport: true,
		GraphiteSeparator:  ".",
		Servers:            []string{"localhost:12003"},
		Log:                testutil.Logger{},
	}

	// Init metrics
	m1 := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"myfield": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m2 := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m3 := metric.New(
		"my_measurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	// Prepare point list
	metrics := []telegraf.Metric{m1}
	metrics2 := []telegraf.Metric{m2, m3}
	err1 := g.Connect()
	require.NoError(t, err1)
	// Send Data
	t.Log("Send first data")
	err2 := g.Write(metrics)
	require.NoError(t, err2)

	// Waiting TCPserver, should reconnect and resend
	wg.Wait()
	t.Log("Finished Waiting for first data")
	var wg2 sync.WaitGroup
	// Start TCP server
	wg2.Add(1)
	TCPServer2WithTags(t, &wg2)
	//Write but expect an error, but reconnect
	err3 := g.Write(metrics2)
	t.Log("Finished writing second data, it should have reconnected automatically")

	require.NoError(t, err3)
	t.Log("Finished writing third data")
	wg2.Wait()
	g.Close()
}

func TestGraphiteOkWithTagsAndSeparatorUnderscore(t *testing.T) {
	var wg sync.WaitGroup
	// Start TCP server
	wg.Add(1)
	t.Log("Starting server")
	TCPServer1WithTagsSeparatorUnderscore(t, &wg)

	// Init plugin
	g := Graphite{
		Prefix:             "my_prefix",
		GraphiteTagSupport: true,
		GraphiteSeparator:  "_",
		Servers:            []string{"localhost:12003"},
		Log:                testutil.Logger{},
	}

	// Init metrics
	m1 := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"myfield": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m2 := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m3 := metric.New(
		"my_measurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	// Prepare point list
	metrics := []telegraf.Metric{m1}
	metrics2 := []telegraf.Metric{m2, m3}
	err1 := g.Connect()
	require.NoError(t, err1)
	// Send Data
	t.Log("Send first data")
	err2 := g.Write(metrics)
	require.NoError(t, err2)

	// Waiting TCPserver, should reconnect and resend
	wg.Wait()
	t.Log("Finished Waiting for first data")
	var wg2 sync.WaitGroup
	// Start TCP server
	wg2.Add(1)
	TCPServer2WithTagsSeparatorUnderscore(t, &wg2)
	//Write but expect an error, but reconnect
	err3 := g.Write(metrics2)
	t.Log("Finished writing second data, it should have reconnected automatically")

	require.NoError(t, err3)
	t.Log("Finished writing third data")
	wg2.Wait()
	g.Close()
}

func TCPServer1(t *testing.T, wg *sync.WaitGroup) {
	tcpServer, err := net.Listen("tcp", "127.0.0.1:12003")
	require.NoError(t, err)
	go func() {
		defer wg.Done()
		conn, _ := (tcpServer).Accept()
		reader := bufio.NewReader(conn)
		tp := textproto.NewReader(reader)
		data1, _ := tp.ReadLine()
		require.Equal(t, "my.prefix.192_168_0_1.mymeasurement.myfield 3.14 1289430000", data1)
		require.NoError(t, conn.Close())
		require.NoError(t, tcpServer.Close())
	}()
}

func TCPServer2(t *testing.T, wg *sync.WaitGroup) {
	tcpServer, _ := net.Listen("tcp", "127.0.0.1:12003")
	go func() {
		defer wg.Done()
		conn2, _ := (tcpServer).Accept()
		reader := bufio.NewReader(conn2)
		tp := textproto.NewReader(reader)
		data2, _ := tp.ReadLine()
		require.Equal(t, "my.prefix.192_168_0_1.mymeasurement 3.14 1289430000", data2)
		data3, _ := tp.ReadLine()
		require.Equal(t, "my.prefix.192_168_0_1.my_measurement 3.14 1289430000", data3)
		require.NoError(t, conn2.Close())
		require.NoError(t, tcpServer.Close())
	}()
}

func TCPServer1WithMultipleTemplates(t *testing.T, wg *sync.WaitGroup) {
	tcpServer, _ := net.Listen("tcp", "127.0.0.1:12003")
	go func() {
		defer wg.Done()
		conn, _ := (tcpServer).Accept()
		reader := bufio.NewReader(conn)
		tp := textproto.NewReader(reader)
		data1, _ := tp.ReadLine()
		require.Equal(t, "my.prefix.mymeasurement.valuetag.192_168_0_1.myfield 3.14 1289430000", data1)
		require.NoError(t, conn.Close())
		require.NoError(t, tcpServer.Close())
	}()
}

func TCPServer2WithMultipleTemplates(t *testing.T, wg *sync.WaitGroup) {
	tcpServer, _ := net.Listen("tcp", "127.0.0.1:12003")
	go func() {
		defer wg.Done()
		conn2, _ := (tcpServer).Accept()
		reader := bufio.NewReader(conn2)
		tp := textproto.NewReader(reader)
		data2, _ := tp.ReadLine()
		require.Equal(t, "my.prefix.mymeasurement.valuetag.192_168_0_1 3.14 1289430000", data2)
		data3, _ := tp.ReadLine()
		require.Equal(t, "my.prefix.192_168_0_1.my_measurement.valuetag 3.14 1289430000", data3)
		require.NoError(t, conn2.Close())
		require.NoError(t, tcpServer.Close())
	}()
}

func TCPServer1WithTags(t *testing.T, wg *sync.WaitGroup) {
	tcpServer, _ := net.Listen("tcp", "127.0.0.1:12003")
	go func() {
		defer wg.Done()
		conn, _ := (tcpServer).Accept()
		reader := bufio.NewReader(conn)
		tp := textproto.NewReader(reader)
		data1, _ := tp.ReadLine()
		require.Equal(t, "my.prefix.mymeasurement.myfield;host=192.168.0.1 3.14 1289430000", data1)
		require.NoError(t, conn.Close())
		require.NoError(t, tcpServer.Close())
	}()
}

func TCPServer2WithTags(t *testing.T, wg *sync.WaitGroup) {
	tcpServer, _ := net.Listen("tcp", "127.0.0.1:12003")
	go func() {
		defer wg.Done()
		conn2, _ := (tcpServer).Accept()
		reader := bufio.NewReader(conn2)
		tp := textproto.NewReader(reader)
		data2, _ := tp.ReadLine()
		require.Equal(t, "my.prefix.mymeasurement;host=192.168.0.1 3.14 1289430000", data2)
		data3, _ := tp.ReadLine()
		require.Equal(t, "my.prefix.my_measurement;host=192.168.0.1 3.14 1289430000", data3)
		require.NoError(t, conn2.Close())
		require.NoError(t, tcpServer.Close())
	}()
}

func TCPServer1WithTagsSeparatorUnderscore(t *testing.T, wg *sync.WaitGroup) {
	tcpServer, _ := net.Listen("tcp", "127.0.0.1:12003")
	go func() {
		defer wg.Done()
		conn, _ := (tcpServer).Accept()
		reader := bufio.NewReader(conn)
		tp := textproto.NewReader(reader)
		data1, _ := tp.ReadLine()
		require.Equal(t, "my_prefix_mymeasurement_myfield;host=192.168.0.1 3.14 1289430000", data1)
		require.NoError(t, conn.Close())
		require.NoError(t, tcpServer.Close())
	}()
}

func TCPServer2WithTagsSeparatorUnderscore(t *testing.T, wg *sync.WaitGroup) {
	tcpServer, _ := net.Listen("tcp", "127.0.0.1:12003")
	go func() {
		defer wg.Done()
		conn2, _ := (tcpServer).Accept()
		reader := bufio.NewReader(conn2)
		tp := textproto.NewReader(reader)
		data2, _ := tp.ReadLine()
		require.Equal(t, "my_prefix_mymeasurement;host=192.168.0.1 3.14 1289430000", data2)
		data3, _ := tp.ReadLine()
		require.Equal(t, "my_prefix_my_measurement;host=192.168.0.1 3.14 1289430000", data3)
		require.NoError(t, conn2.Close())
		require.NoError(t, tcpServer.Close())
	}()
}
