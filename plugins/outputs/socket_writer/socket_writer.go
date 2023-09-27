//go:generate ../../../tools/readme_config_includer/generator
package socket_writer

import (
	"crypto/tls"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

//go:embed sample.conf
var sampleConfig string

type SocketWriter struct {
	ContentEncoding string `toml:"content_encoding"`
	Address         string
	KeepAlivePeriod *config.Duration
	tlsint.ClientConfig
	Log telegraf.Logger `toml:"-"`

	serializers.Serializer

	encoder internal.ContentEncoder

	net.Conn
}

func (*SocketWriter) SampleConfig() string {
	return sampleConfig
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
		sw.Log.Debugf("Unable to configure keep alive (%s): %s", sw.Address, err)
	}
	//set encoder
	sw.encoder, err = internal.NewContentEncoder(sw.ContentEncoding)
	if err != nil {
		return err
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
	if *sw.KeepAlivePeriod == 0 {
		return tcpc.SetKeepAlive(false)
	}
	if err := tcpc.SetKeepAlive(true); err != nil {
		return err
	}
	return tcpc.SetKeepAlivePeriod(time.Duration(*sw.KeepAlivePeriod))
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
			sw.Log.Debugf("Could not serialize metric: %v", err)
			continue
		}

		bs, err = sw.encoder.Encode(bs)
		if err != nil {
			sw.Log.Debugf("Could not encode metric: %v", err)
			continue
		}

		if _, err := sw.Conn.Write(bs); err != nil {
			//TODO log & keep going with remaining strings
			var netErr net.Error
			if errors.As(err, &netErr) {
				// permanent error. close the connection
				sw.Close()
				sw.Conn = nil
				return fmt.Errorf("closing connection: %w", netErr)
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

func init() {
	outputs.Add("socket_writer", func() telegraf.Output {
		return &SocketWriter{}
	})
}
