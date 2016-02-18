package kafka

import (
	"crypto/tls"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"

	"github.com/Shopify/sarama"
)

type Kafka struct {
	// Kafka brokers to send metrics to
	Brokers []string
	// Kafka topic
	Topic string
	// Routing Key Tag
	RoutingTag string `toml:"routing_tag"`

	// Legacy SSL config options
	// TLS client certificate
	Certificate string
	// TLS client key
	Key string
	// TLS certificate authority
	CA string

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`

	// Skip SSL verification
	InsecureSkipVerify bool

	tlsConfig tls.Config
	producer  sarama.SyncProducer

	serializer serializers.Serializer
}

var sampleConfig = `
  ## URLs of kafka brokers
  brokers = ["localhost:9092"]
  ## Kafka topic for producer messages
  topic = "telegraf"
  ## Telegraf tag to use as a routing key
  ##  ie, if this tag exists, it's value will be used as the routing key
  routing_tag = "host"

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## Data format to output. This can be "influx" or "graphite"
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

func (k *Kafka) SetSerializer(serializer serializers.Serializer) {
	k.serializer = serializer
}

func (k *Kafka) Connect() error {
	config := sarama.NewConfig()
	// Wait for all in-sync replicas to ack the message
	config.Producer.RequiredAcks = sarama.WaitForAll
	// Retry up to 10 times to produce the message
	config.Producer.Retry.Max = 10

	// Legacy support ssl config
	if k.Certificate != "" {
		k.SSLCert = k.Certificate
		k.SSLCA = k.CA
		k.SSLKey = k.Key
	}

	tlsConfig, err := internal.GetTLSConfig(
		k.SSLCert, k.SSLKey, k.SSLCA, k.InsecureSkipVerify)
	if err != nil {
		return err
	}

	if tlsConfig != nil {
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Enable = true
	}

	producer, err := sarama.NewSyncProducer(k.Brokers, config)
	if err != nil {
		return err
	}
	k.producer = producer
	return nil
}

func (k *Kafka) Close() error {
	return k.producer.Close()
}

func (k *Kafka) SampleConfig() string {
	return sampleConfig
}

func (k *Kafka) Description() string {
	return "Configuration for the Kafka server to send metrics to"
}

func (k *Kafka) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		values, err := k.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		var pubErr error
		for _, value := range values {
			m := &sarama.ProducerMessage{
				Topic: k.Topic,
				Value: sarama.StringEncoder(value),
			}
			if h, ok := metric.Tags()[k.RoutingTag]; ok {
				m.Key = sarama.StringEncoder(h)
			}

			_, _, pubErr = k.producer.SendMessage(m)
		}

		if pubErr != nil {
			return fmt.Errorf("FAILED to send kafka message: %s\n", pubErr)
		}
	}
	return nil
}

func init() {
	outputs.Add("kafka", func() telegraf.Output {
		return &Kafka{}
	})
}
