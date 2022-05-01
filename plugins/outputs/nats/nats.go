package nats

import (
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type NATS struct {
	Servers     []string `toml:"servers"`
	Secure      bool     `toml:"secure"`
	Name        string   `toml:"name"`
	Username    string   `toml:"username"`
	Password    string   `toml:"password"`
	Credentials string   `toml:"credentials"`
	Subject     string   `toml:"subject"`

	tls.ClientConfig

	Log telegraf.Logger `toml:"-"`

	conn       *nats.Conn
	serializer serializers.Serializer
}

func (n *NATS) SetSerializer(serializer serializers.Serializer) {
	n.serializer = serializer
}

func (n *NATS) Connect() error {
	var err error

	opts := []nats.Option{
		nats.MaxReconnects(-1),
	}

	// override authentication, if any was specified
	if n.Username != "" && n.Password != "" {
		opts = append(opts, nats.UserInfo(n.Username, n.Password))
	}

	if n.Credentials != "" {
		opts = append(opts, nats.UserCredentials(n.Credentials))
	}

	if n.Name != "" {
		opts = append(opts, nats.Name(n.Name))
	}

	if n.Secure {
		tlsConfig, err := n.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}

		opts = append(opts, nats.Secure(tlsConfig))
	}

	// try and connect
	n.conn, err = nats.Connect(strings.Join(n.Servers, ","), opts...)

	return err
}

func (n *NATS) Close() error {
	n.conn.Close()
	return nil
}

func (n *NATS) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		buf, err := n.serializer.Serialize(metric)
		if err != nil {
			n.Log.Debugf("Could not serialize metric: %v", err)
			continue
		}

		err = n.conn.Publish(n.Subject, buf)
		if err != nil {
			return fmt.Errorf("FAILED to send NATS message: %s", err)
		}
	}
	return nil
}

func init() {
	outputs.Add("nats", func() telegraf.Output {
		return &NATS{}
	})
}
