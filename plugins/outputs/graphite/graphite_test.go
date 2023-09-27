package graphite

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
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
	require.NoError(t, g.Init())

	// Init metrics
	m1 := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"mymeasurement": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	// Prepare point list
	metrics := []telegraf.Metric{m1}

	require.NoError(t, g.Connect())
	err := g.Write(metrics)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNotConnected)
}

func TestGraphiteReconnect(t *testing.T) {
	m := metric.New(
		"mymeasurement",
		map[string]string{
			"host":       "192.168.0.1",
			"datacenter": "|us-west-2|",
		},
		map[string]interface{}{"myfield": float64(0.123)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	g := Graphite{
		Servers:             []string{"localhost:12042"},
		Log:                 testutil.Logger{},
		GraphiteStrictRegex: `[^a-zA-Z0-9-:._=|\p{L}]`,
	}
	require.NoError(t, g.Init())

	t.Log("Writing metric, without any server up, expected to fail")
	require.NoError(t, g.Connect())
	require.Error(t, g.Write([]telegraf.Metric{m}))

	var wg sync.WaitGroup
	wg.Add(1)
	t.Log("Starting server")
	tcpServer, err := net.Listen("tcp", "127.0.0.1:12042")
	require.NoError(t, err)

	t.Log("Writing metric after server came up, we expect automatic reconnect on write without calling Connect() again")
	require.NoError(t, g.Write([]telegraf.Metric{m}))

	go func() {
		defer wg.Done()
		conn, _ := (tcpServer).Accept()
		reader := bufio.NewReader(conn)
		tp := textproto.NewReader(reader)
		data1, _ := tp.ReadLine()
		require.Equal(t, "192_168_0_1.|us-west-2|.mymeasurement.myfield 0.123 1289430000", data1)
		require.NoError(t, conn.Close())
		require.NoError(t, tcpServer.Close())
	}()

	wg.Wait()
	require.NoError(t, g.Close())
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
	require.NoError(t, g.Init())

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
	err := g.Close()
	require.NoError(t, err)
}

func TestGraphiteStrictRegex(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	t.Log("Starting server")
	tcpServer, err := net.Listen("tcp", "127.0.0.1:12042")
	require.NoError(t, err)
	go func() {
		defer wg.Done()
		conn, _ := (tcpServer).Accept()
		reader := bufio.NewReader(conn)
		tp := textproto.NewReader(reader)
		data1, _ := tp.ReadLine()
		require.Equal(t, "192_168_0_1.|us-west-2|.mymeasurement.myfield 0.123 1289430000", data1)
		require.NoError(t, conn.Close())
		require.NoError(t, tcpServer.Close())
	}()

	m := metric.New(
		"mymeasurement",
		map[string]string{
			"host":       "192.168.0.1",
			"datacenter": "|us-west-2|",
		},
		map[string]interface{}{"myfield": float64(0.123)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	g := Graphite{
		Servers:             []string{"localhost:12042"},
		Log:                 testutil.Logger{},
		GraphiteStrictRegex: `[^a-zA-Z0-9-:._=|\p{L}]`,
	}
	require.NoError(t, g.Init())
	require.NoError(t, g.Connect())
	require.NoError(t, g.Write([]telegraf.Metric{m}))

	wg.Wait()
	require.NoError(t, g.Close())
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
	require.NoError(t, g.Init())

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
	err := g.Close()
	require.NoError(t, err)
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
	require.NoError(t, g.Init())

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
	err := g.Close()
	require.NoError(t, err)
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
	require.NoError(t, g.Init())

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
	err := g.Close()
	require.NoError(t, err)
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
	require.NoError(t, g.Init())

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
	err := g.Close()
	require.NoError(t, err)
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
	require.NoError(t, g.Init())

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
	err := g.Close()
	require.NoError(t, err)
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
	require.NoError(t, g.Init())

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
	err := g.Close()
	require.NoError(t, err)
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	container := testutil.Container{
		Image:        "graphiteapp/graphite-statsd",
		ExposedPorts: []string{"8080", "2003", "2004"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("8080"),
			wait.ForListeningPort("2003"),
			wait.ForListeningPort("2004"),
			wait.ForLog("run: statsd:"),
		),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	// Init plugin
	plugin := Graphite{
		Servers:  []string{container.Address + ":" + container.Ports["2003"]},
		Template: "measurement.tags.field",
		Timeout:  config.Duration(2 * time.Second),
		Log:      testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	metrics := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{"source": "foo"},
			map[string]interface{}{"value": 42.0},
			time.Now(),
		),
		metric.New(
			"test",
			map[string]string{"source": "bar"},
			map[string]interface{}{"value": 23.0},
			time.Now(),
		),
	}

	// Verify that we can successfully write data
	require.NoError(t, plugin.Write(metrics))

	// Wait for the data to settle and check if we got the metrics
	url := fmt.Sprintf("http://%s:%s/metrics/index.json", container.Address, container.Ports["8080"])
	require.Eventually(t, func() bool {
		var actual []string
		if err := query(url, &actual); err != nil {
			t.Logf("encountered error %v", err)
			return false
		}
		var foundFoo, foundBar bool
		for _, m := range actual {
			switch m {
			case "test.bar":
				foundBar = true
			case "test.foo":
				foundFoo = true
			default:
				continue
			}
			if foundBar && foundFoo {
				return true
			}
		}
		return false
	}, 10*time.Second, 100*time.Millisecond)
}

func query(url string, data interface{}) error {
	//nolint:gosec // Parameters are fixed in the above call
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("response:", resp)
		return err
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("raw:", string(raw))
		return err
	}
	resp.Body.Close()

	return json.Unmarshal(raw, &data)
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
	tcpServer, err := net.Listen("tcp", "127.0.0.1:12003")
	require.NoError(t, err)
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
	tcpServer, err := net.Listen("tcp", "127.0.0.1:12003")
	require.NoError(t, err)
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
	tcpServer, err := net.Listen("tcp", "127.0.0.1:12003")
	require.NoError(t, err)
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
	tcpServer, err := net.Listen("tcp", "127.0.0.1:12003")
	require.NoError(t, err)
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
	tcpServer, err := net.Listen("tcp", "127.0.0.1:12003")
	require.NoError(t, err)
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
	tcpServer, err := net.Listen("tcp", "127.0.0.1:12003")
	require.NoError(t, err)
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
	tcpServer, err := net.Listen("tcp", "127.0.0.1:12003")
	require.NoError(t, err)
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
