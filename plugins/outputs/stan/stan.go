package stan

import (
	"fmt"

	nats "github.com/nats-io/go-nats"
	stan "github.com/nats-io/go-nats-streaming"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type Stan struct {
	// NATS Streaming Cluster ID
	ClusterID string `toml:"cluster_id"`

	// NATS Streaming Client ID
	ClientID string `toml:"client_id"`

	// Servers is the NATS server pool to connect to
	Servers []string
	// Credentials
	Username string
	Password string

	// STAN subject to publish metrics to
	Subject string

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool `toml:"insecure_skip_verify"`

	conn       stan.Conn
	serializer serializers.Serializer
}

var sampleConfig = `
  ## NATS Streaming Cluster ID
  cluster_id = "test-cluster"
  ## Client ID
  client_id = "telegraf-client-id"
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
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

func (n *Stan) SetSerializer(serializer serializers.Serializer) {
	n.serializer = serializer
}

func (n *Stan) Connect() error {
	var err error

	natsOpts := nats.DefaultOptions

	// override max reconnection tries
	natsOpts.MaxReconnect = -1

	// override servers, if any were specified
	natsOpts.Servers = n.Servers

	// override authentication, if any was specified
	if n.Username != "" {
		natsOpts.User = n.Username
		natsOpts.Password = n.Password
	}

	// override TLS, if it was specified
	tlsConfig, err := internal.GetTLSConfig(n.SSLCert, n.SSLKey, n.SSLCA, n.InsecureSkipVerify)
	if err != nil {
		return err
	}
	if tlsConfig != nil {
		// set NATS connection TLS options
		natsOpts.Secure = true
		natsOpts.TLSConfig = tlsConfig
	}

	// try and connect to the NATS server
	natsConn, err := natsOpts.Connect()
	if err != nil {
		return err
	}

	// try and connect to the NATS Streaming cluster
	n.conn, err = stan.Connect(n.ClusterID, n.ClientID, stan.NatsConn(natsConn))

	return err
}

func (n *Stan) Close() error {
	return n.conn.Close()
}

func (n *Stan) SampleConfig() string {
	return sampleConfig
}

func (n *Stan) Description() string {
	return "Send telegraf measurements to NATS Streaming"
}

func (n *Stan) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		buf, err := n.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		err = n.conn.Publish(n.Subject, buf)
		if err != nil {
			return fmt.Errorf("FAILED to send NATS Streaming message: %s", err)
		}
	}

	return nil
}

func init() {
	outputs.Add("stan", func() telegraf.Output {
		return &Stan{}
	})
}
