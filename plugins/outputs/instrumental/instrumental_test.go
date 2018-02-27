package instrumental

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
)

func TestWrite(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	TCPServer(t, &wg)

	i := Instrumental{
		Host:     "127.0.0.1",
		ApiToken: "abc123token",
		Prefix:   "my.prefix",
	}

	// Default to gauge
	m1, _ := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"myfield": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m2, _ := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1", "metric_type": "set"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	metrics := []telegraf.Metric{m1, m2}
	i.Write(metrics)

	// Counter and Histogram are increments
	m3, _ := metric.New(
		"my_histogram",
		map[string]string{"host": "192.168.0.1", "metric_type": "histogram"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	// We will modify metric names that won't be accepted by Instrumental
	m4, _ := metric.New(
		"bad_metric_name",
		map[string]string{"host": "192.168.0.1:8888::123", "metric_type": "counter"},
		map[string]interface{}{"value": 1},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	// We will drop metric values that won't be accepted by Instrumental
	m5, _ := metric.New(
		"bad_values",
		map[string]string{"host": "192.168.0.1", "metric_type": "counter"},
		map[string]interface{}{"value": "\" 3:30\""},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	m6, _ := metric.New(
		"my_counter",
		map[string]string{"host": "192.168.0.1", "metric_type": "counter"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	metrics = []telegraf.Metric{m3, m4, m5, m6}
	i.Write(metrics)

	wg.Wait()
}

func TCPServer(t *testing.T, wg *sync.WaitGroup) {
	tcpServer, _ := net.Listen("tcp", "127.0.0.1:8000")
	go func() {
		defer wg.Done()
		conn, _ := tcpServer.Accept()
		conn.SetDeadline(time.Now().Add(1 * time.Second))
		reader := bufio.NewReader(conn)
		tp := textproto.NewReader(reader)

		hello, _ := tp.ReadLine()
		assert.Equal(t, "hello version go/telegraf/1.1", hello)
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
		assert.Equal(t, "hello version go/telegraf/1.1", hello)
		auth, _ = tp.ReadLine()
		assert.Equal(t, "authenticate abc123token", auth)
		conn.Write([]byte("ok\nok\n"))

		data3, _ := tp.ReadLine()
		assert.Equal(t, "increment my.prefix.192_168_0_1.my_histogram 3.14 1289430000", data3)

		data4, _ := tp.ReadLine()
		assert.Equal(t, "increment my.prefix.192_168_0_1_8888_123.bad_metric_name 1 1289430000", data4)

		data5, _ := tp.ReadLine()
		assert.Equal(t, "increment my.prefix.192_168_0_1.my_counter 3.14 1289430000", data5)

		data6, _ := tp.ReadLine()
		assert.Equal(t, "", data6)

		conn.Close()
	}()
}
