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
	// Compression Codec Tag
	CompressionCodec int
	// RequiredAcks Tag
	RequiredAcks int
	// MaxRetry Tag
	MaxRetry int

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

  ## CompressionCodec represents the various compression codecs recognized by
  ## Kafka in messages.
  ##  0 : No compression
  ##  1 : Gzip compression
  ##  2 : Snappy compression
  compression_codec = 0

  ##  RequiredAcks is used in Produce Requests to tell the broker how many
  ##  replica acknowledgements it must see before responding
  ##   0 : the producer never waits for an acknowledgement from the broker.
  ##       This option provides the lowest latency but the weakest durability
  ##       guarantees (some data will be lost when a server fails).
  ##   1 : the producer gets an acknowledgement after the leader replica has
  ##       received the data. This option provides better durability as the
  ##       client waits until the server acknowledges the request as successful
  ##       (only messages that were written to the now-dead leader but not yet
  ##       replicated will be lost).
  ##   -1: the producer gets an acknowledgement after all in-sync replicas have
  ##       received the data. This option provides the best durability, we
  ##       guarantee that no messages will be lost as long as at least one in
  ##       sync replica remains.
  required_acks = -1

  ##  The total number of times to retry sending a message
  max_retry = 3

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

func (k *Kafka) SetSerializer(serializer serializers.Serializer) {
	k.serializer = serializer
}

func (k *Kafka) Connect() error {
	config := sarama.NewConfig()

	config.Producer.RequiredAcks = sarama.RequiredAcks(k.RequiredAcks)
	config.Producer.Compression = sarama.CompressionCodec(k.CompressionCodec)
	config.Producer.Retry.Max = k.MaxRetry

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
		return &Kafka{
			MaxRetry:     3,
			RequiredAcks: -1,
		}
	})
}
