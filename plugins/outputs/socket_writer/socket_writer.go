package socket_writer

import (
	"fmt"
	"net"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type SocketWriter struct {
	Address string

	serializers.Serializer

	net.Conn
}

func (sw *SocketWriter) Description() string {
	return "Generic socket writer capable of handling multiple socket types."
}

func (sw *SocketWriter) SampleConfig() string {
	return `
  ## URL to connect to
  # address = "tcp://127.0.0.1:8094"
  # address = "tcp://example.com:http"
  # address = "tcp4://127.0.0.1:8094"
  # address = "tcp6://127.0.0.1:8094"
  # address = "tcp6://[2001:db8::1]:8094"
  # address = "udp://127.0.0.1:8094"
  # address = "udp4://127.0.0.1:8094"
  # address = "udp6://127.0.0.1:8094"
  # address = "unix:///tmp/telegraf.sock"
  # address = "unixgram:///tmp/telegraf.sock"

  ## Data format to generate.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  # data_format = "influx"
`
}

func (sw *SocketWriter) SetSerializer(s serializers.Serializer) {
	sw.Serializer = s
}

func (sw *SocketWriter) Connect() error {
	spl := strings.SplitN(sw.Address, "://", 2)
	if len(spl) != 2 {
		return fmt.Errorf("invalid address: %s", sw.Address)
	}

	c, err := net.Dial(spl[0], spl[1])
	if err != nil {
		return err
	}

	sw.Conn = c
	return nil
}

// Write writes the given metrics to the destination.
// If an error is encountered, it is up to the caller to retry the same write again later.
// Not parallel safe.
func (sw *SocketWriter) Write(metrics []telegraf.Metric) error {
	if sw.Conn == nil {
		// previous write failed with permanent error and socket was closed.
		if err := sw.Connect(); err != nil {
			return err
		}
	}

	for _, m := range metrics {
		bs, err := sw.Serialize(m)
		if err != nil {
			//TODO log & keep going with remaining metrics
			return err
		}
		if _, err := sw.Conn.Write(bs); err != nil {
			//TODO log & keep going with remaining strings
			if err, ok := err.(net.Error); !ok || !err.Temporary() {
				// permanent error. close the connection
				sw.Close()
				sw.Conn = nil
			}
			return err
		}
	}

	return nil
}

func newSocketWriter() *SocketWriter {
	s, _ := serializers.NewInfluxSerializer()
	return &SocketWriter{
		Serializer: s,
	}
}

func init() {
	outputs.Add("socket_writer", func() telegraf.Output { return newSocketWriter() })
}
