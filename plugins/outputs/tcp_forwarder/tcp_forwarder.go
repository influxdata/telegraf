package tcp_forwarder

import (
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// TCPForwarder structure for configuration and server
type TCPForwarder struct {
	sync.Mutex

	Server     string
	Timeout    internal.Duration
	DataFormat string
	Reconnect  bool
	conn       net.Conn
	serializer serializers.Serializer
}

var sampleConfig = `
  ## TCP servers/endpoints to send metrics to.
  server = "localhost:8089"
  ## timeout for the write connection
  timeout = "5s"
  ## force reconnection before every push
  reconnect = false
  ## Data format to _output_.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

// SetSerializer is the function from output plugin to use a serializer for
// formating data
func (t *TCPForwarder) SetSerializer(serializer serializers.Serializer) {
	t.serializer = serializer
}

// Connect is the default output plugin connection function who make sure it
// can connect to the endpoint
func (t *TCPForwarder) Connect() error {

	if len(t.Server) == 0 {
		t.Server = "localhost:8089"
	}
	if t.Timeout.Duration.Seconds() < 1 {
		t.Timeout.Duration = time.Second
	}

	// try connect
	if err := t.reconnect(); err != nil {
		return err
	}

	return nil
}

func (t *TCPForwarder) reconnect() error {
	if t.Reconnect {
		t.Close()
	}
	if t.Reconnect || t.isClosed() {
		conn, err := net.DialTimeout("tcp", t.Server, t.Timeout.Duration)
		if err == nil {
			fmt.Println("TCP_forwarder, re-connected: " + t.Server)
			t.conn = conn
		} else {
			log.Printf("Error connecting to <%s>: %s", t.Server, err.Error())
			return err
		}
	}
	return nil
}

func (t *TCPForwarder) isClosed() bool {
	var one []byte
	if t.conn == nil {
		return true
	}

	t.conn.SetReadDeadline(time.Now())
	if _, err := t.conn.Read(one); err == io.EOF {
		t.Close()
		return true
	}
	return false
}

// Close is use to close connection to all Tcp endpoints
func (t *TCPForwarder) Close() error {
	t.Lock()
	defer t.Unlock()
	if t.conn != nil {
		t.conn.Close()
		t.conn = nil
	}
	return nil
}

// SampleConfig is the default function who return the default configuration
// for tcp forwarder output
func (t *TCPForwarder) SampleConfig() string {
	return sampleConfig
}

// Description is the default function who return the description of tcp
// forwarder output
func (t *TCPForwarder) Description() string {
	return "Generic TCP forwarder for metrics"
}

// Write is the default function to call to "send" a metric through the Output
func (t *TCPForwarder) Write(metrics []telegraf.Metric) error {
	// reconnect if needed
	if err := t.reconnect(); err != nil {
		return err
	}
	// Prepare data
	t.Lock()
	defer t.Unlock()

	var bp []string
	for _, metric := range metrics {
		sMetrics, err := t.serializer.Serialize(metric)
		if err != nil {
			log.Printf("Error while serializing some metrics: %s", err.Error())
		}
		bp = append(bp, sMetrics...)
	}

	// TODO should we add a join function in serialiser ?
	points := strings.Join(bp, "\n") + "\n"

	t.conn.SetWriteDeadline(time.Now().Add(t.Timeout.Duration))
	if _, e := fmt.Fprintf(t.conn, points); e != nil {
		fmt.Println("ERROR: " + e.Error())
		t.conn.Close()
		t.conn = nil
		return errors.New("Could not write to tcp endpoint\n")
	}
	return nil
}

func init() {
	outputs.Add("tcp_forwarder", func() telegraf.Output {
		return &TCPForwarder{}
	})
}
