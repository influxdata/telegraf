package stan

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"log"
	"strings"
)

type Stan struct {
	Servers     []string `toml:"servers"`
	Secure      bool     `toml:"secure"`
	Username    string   `toml:"username"`
	Password    string   `toml:"password"`
	Credentials string   `toml:"credentials"`
	Subject     string   `toml:"subject"`
	ClusterID   string   `toml:"cluster_id"`
	ClientID    string   `toml:"client_id"`

	tls.ClientConfig

	natsConn   *nats.Conn
	stanConn   stan.Conn
	serializer serializers.Serializer
}

var sampleConfig = `
  ## URLs of NATS Streaming servers
  servers = ["nats://localhost:4222"]

  ## Optional credentials
  # username = ""
  # password = ""

  ## Optional NATS 2.0 and NATS NGS compatible user credentials
  # credentials = "/etc/telegraf/nats.creds"

  ## NATS Streaming subject for producer messages
  subject = "telegraf"

  ## NATS Streaming cluster id
  cluster_id = "test-cluster"

  ## NATS Streaming client id
  client_id = "telegraf"

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

func (s *Stan) SetSerializer(serializer serializers.Serializer) {
	s.serializer = serializer
}

func (s *Stan) Connect() error {
	var err error

	natsOpts := []nats.Option{
		nats.MaxReconnects(-1),
	}

	if s.Username != "" {
		natsOpts = append(natsOpts, nats.UserInfo(s.Username, s.Password))
	}

	if s.Secure {
		tlsConfig, err := s.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}

		natsOpts = append(natsOpts, nats.Secure(tlsConfig))
	}

	s.natsConn, err = nats.Connect(strings.Join(s.Servers, ","), natsOpts...)
	if err != nil {
		return err
	}

	stanOpts := []stan.Option{
		stan.NatsConn(s.natsConn),
	}

	s.stanConn, err = stan.Connect(s.ClusterID, s.ClientID, stanOpts...)
	if err != nil {
		return err
	}

	return nil
}

func (s *Stan) Close() error {
	if err := s.stanConn.Close(); err != nil {
		return err
	}

	s.natsConn.Close()

	return nil
}

func (s *Stan) SampleConfig() string {
	return sampleConfig
}

func (s *Stan) Description() string {
	return "Send telegraf measurements to NATS Streaming"
}

func (s *Stan) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		buf, err := s.serializer.Serialize(metric)
		if err != nil {
			log.Printf("[outputs.stan] could not serialize telegraf metric: %v", err)
			continue
		}

		err = s.stanConn.Publish(s.Subject, buf)
		if err != nil {
			return fmt.Errorf("failed to send nats streaming message: %s", err)
		}
	}

	return nil
}

func init() {
	outputs.Add("stan", func() telegraf.Output {
		return &Stan{}
	})
}
