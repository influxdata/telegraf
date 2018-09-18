package socket_writer

import (
	"fmt"
	"log"
	"net"
	"strings"

	"crypto/tls"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type SocketWriter struct {
	Address         string
	KeepAlivePeriod *internal.Duration
	tlsint.ClientConfig

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

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Period between keep alive probes.
  ## Only applies to TCP sockets.
  ## 0 disables keep alive probes.
  ## Defaults to the OS configuration.
  # keep_alive_period = "5m"

  ## Data format to generate.
  ## Each data format has its own unique set of configuration options, read
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

	tlsCfg, err := sw.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	var c net.Conn
	if tlsCfg == nil {
		c, err = net.Dial(spl[0], spl[1])
	} else {
		c, err = tls.Dial(spl[0], spl[1], tlsCfg)
	}
	if err != nil {
		return err
	}

	if err := sw.setKeepAlive(c); err != nil {
		log.Printf("unable to configure keep alive (%s): %s", sw.Address, err)
	}

	sw.Conn = c
	return nil
}

func (sw *SocketWriter) setKeepAlive(c net.Conn) error {
	if sw.KeepAlivePeriod == nil {
		return nil
	}
	tcpc, ok := c.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("cannot set keep alive on a %s socket", strings.SplitN(sw.Address, "://", 2)[0])
	}
	if sw.KeepAlivePeriod.Duration == 0 {
		return tcpc.SetKeepAlive(false)
	}
	if err := tcpc.SetKeepAlive(true); err != nil {
		return err
	}
	return tcpc.SetKeepAlivePeriod(sw.KeepAlivePeriod.Duration)
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
				return fmt.Errorf("closing connection: %v", err)
			}
			return err
		}
	}

	return nil
}

// Close closes the connection. Noop if already closed.
func (sw *SocketWriter) Close() error {
	if sw.Conn == nil {
		return nil
	}
	err := sw.Conn.Close()
	sw.Conn = nil
	return err
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
