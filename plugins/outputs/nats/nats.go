package nats

import (
	"fmt"
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	nats_client "github.com/nats-io/go-nats"
)

type NATS struct {
	Servers  []string `toml:"servers"`
	Secure   bool     `toml:"secure"`
	Username string   `toml:"username"`
	Password string   `toml:"password"`
	Subject  string   `toml:"subject"`
	tls.ClientConfig

	conn       *nats_client.Conn
	serializer serializers.Serializer
}

var sampleConfig = `
  ## URLs of NATS servers
  servers = ["nats://localhost:4222"]
  ## Optional credentials
  # username = ""
  # password = ""
  ## NATS subject for producer messages
  subject = "telegraf"

  ## Use Transport Layer Security
  # secure = false

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

func (n *NATS) SetSerializer(serializer serializers.Serializer) {
	n.serializer = serializer
}

func (n *NATS) Connect() error {
	var err error

	// set default NATS connection options
	opts := nats_client.DefaultOptions

	// override max reconnection tries
	opts.MaxReconnect = -1

	// override servers, if any were specified
	opts.Servers = n.Servers

	// override authentication, if any was specified
	if n.Username != "" {
		opts.User = n.Username
		opts.Password = n.Password
	}

	if n.Secure {
		tlsConfig, err := n.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}

		opts.Secure = true
		opts.TLSConfig = tlsConfig
	}

	// try and connect
	n.conn, err = opts.Connect()

	return err
}

func (n *NATS) Close() error {
	n.conn.Close()
	return nil
}

func (n *NATS) SampleConfig() string {
	return sampleConfig
}

func (n *NATS) Description() string {
	return "Send telegraf measurements to NATS"
}

func (n *NATS) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		buf, err := n.serializer.Serialize(metric)
		if err != nil {
			log.Printf("D! [outputs.nats] Could not serialize metric: %v", err)
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
