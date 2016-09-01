package nats

import (
	"fmt"

	nats_client "github.com/nats-io/nats"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type NATS struct {
	// Servers is the NATS server pool to connect to
	Servers []string
	// Credentials
	Username string
	Password string
	// NATS subject to publish metrics to
	Subject string

	// Path to CA file
	CAFile string `toml:"tls_ca"`

	// Skip SSL verification
	InsecureSkipVerify bool

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
  ## Optional TLS Config
  ## CA certificate used to self-sign NATS server(s) TLS certificate(s)
  # tls_ca = "/etc/telegraf/ca.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Data format to output.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

func (n *NATS) SetSerializer(serializer serializers.Serializer) {
	n.serializer = serializer
}

func (n *NATS) Connect() error {
	var err error
	// set NATS connection options
	opts := nats_client.DefaultOptions
	opts.Servers = n.Servers
	if n.Username != "" {
		opts.User = n.Username
		opts.Password = n.Password
	}

	// is TLS enabled?
	tlsConfig, err := internal.GetTLSConfig(
		"", "", n.CAFile, n.InsecureSkipVerify)
	if err != nil {
		return err
	}
	if tlsConfig != nil {
		// set NATS connection TLS options
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
		values, err := n.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		var pubErr error
		for _, value := range values {
			err = n.conn.Publish(n.Subject, []byte(value))
			if err != nil {
				pubErr = err
			}
		}

		if pubErr != nil {
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
