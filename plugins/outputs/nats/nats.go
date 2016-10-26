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
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
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

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
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

	// override TLS, if it was specified
	tlsConfig, err := internal.GetTLSConfig(
		n.SSLCert, n.SSLKey, n.SSLCA, n.InsecureSkipVerify)
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
