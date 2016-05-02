package instrumental

import (
	"bufio"
	"net"
	"net/textproto"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	go TCPServer(t, &wg)
	// Give the fake TCP server some time to start:
	time.Sleep(time.Millisecond * 100)

	i := Instrumental{
		Host:     "127.0.0.1",
		ApiToken: "abc123token",
		Prefix:   "my.prefix",
	}
	i.Connect()

	// Default to gauge
	m1, _ := telegraf.NewMetric(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"myfield": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m2, _ := telegraf.NewMetric(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1", "metric_type": "set"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	// Simulate a connection close and reconnect.
	metrics := []telegraf.Metric{m1, m2}
	i.Write(metrics)
	i.Close()

	// Counter and Histogram are increments
	m3, _ := telegraf.NewMetric(
		"my_histogram",
		map[string]string{"host": "192.168.0.1", "metric_type": "histogram"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	// We will drop metrics that simply won't be accepted by Instrumental
	m4, _ := telegraf.NewMetric(
		"bad_values",
		map[string]string{"host": "192.168.0.1", "metric_type": "counter"},
		map[string]interface{}{"value": "\" 3:30\""},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m5, _ := telegraf.NewMetric(
		"my_counter",
		map[string]string{"host": "192.168.0.1", "metric_type": "counter"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	metrics = []telegraf.Metric{m3, m4, m5}
	i.Write(metrics)

	wg.Wait()
	i.Close()
}

func TCPServer(t *testing.T, wg *sync.WaitGroup) {
	tcpServer, _ := net.Listen("tcp", "127.0.0.1:8000")
	defer wg.Done()
	conn, _ := tcpServer.Accept()
	conn.SetDeadline(time.Now().Add(1 * time.Second))
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)

	hello, _ := tp.ReadLine()
	assert.Equal(t, "hello version go/telegraf/1.0", hello)
	auth, _ := tp.ReadLine()
	assert.Equal(t, "authenticate abc123token", auth)

	conn.Write([]byte("ok\nok\n"))

	data1, _ := tp.ReadLine()
	assert.Equal(t, "gauge my.prefix.192_168_0_1.mymeasurement.myfield 3.14 1289430000", data1)
	data2, _ := tp.ReadLine()
	assert.Equal(t, "gauge my.prefix.192_168_0_1.mymeasurement 3.14 1289430000", data2)

	conn, _ = tcpServer.Accept()
	conn.SetDeadline(time.Now().Add(1 * time.Second))
	reader = bufio.NewReader(conn)
	tp = textproto.NewReader(reader)

	hello, _ = tp.ReadLine()
	assert.Equal(t, "hello version go/telegraf/1.0", hello)
	auth, _ = tp.ReadLine()
	assert.Equal(t, "authenticate abc123token", auth)

	conn.Write([]byte("ok\nok\n"))

	data3, _ := tp.ReadLine()
	assert.Equal(t, "increment my.prefix.192_168_0_1.my_histogram 3.14 1289430000", data3)
	data4, _ := tp.ReadLine()
	assert.Equal(t, "increment my.prefix.192_168_0_1.my_counter 3.14 1289430000", data4)

	conn.Close()
}
