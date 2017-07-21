package kafka

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"

	"github.com/Shopify/sarama"
)

const (
	TOPIC_SUFFIX_METHOD_EMPTY       uint8 = iota
	TOPIC_SUFFIX_METHOD_MEASUREMENT
	TOPIC_SUFFIX_METHOD_TAG
	TOPIC_SUFFIX_METHOD_TAGS
)

var TopicSuffixMethodStringToUID = map[string]uint8{
	"":            TOPIC_SUFFIX_METHOD_EMPTY,
	"measurement": TOPIC_SUFFIX_METHOD_MEASUREMENT,
	"tag":         TOPIC_SUFFIX_METHOD_TAG,
	"tags":        TOPIC_SUFFIX_METHOD_TAGS,
}

type (
	Kafka struct {
		// Kafka brokers to send metrics to
		Brokers []string
		// Kafka topic
		Topic string
		// Kafka topic suffix option
		TopicSuffix TopicSuffix `toml:"topic_suffix"`
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

		// SASL Username
		SASLUsername string `toml:"sasl_username"`
		// SASL Password
		SASLPassword string `toml:"sasl_password"`

		tlsConfig tls.Config
		producer  sarama.SyncProducer

		serializer serializers.Serializer

		topicSuffixMethodUID uint8
	}
	TopicSuffix struct {
		Method       string `toml:"method"`
		Key          string `toml:"key"`
		Keys         []string `toml:"keys"`
		KeySeparator string `toml:"key_separator"`
	}
)

var sampleConfig = `
  ## URLs of kafka brokers
  brokers = ["localhost:9092"]
  ## Kafka topic for producer messages
  topic = "telegraf"

  ## Optional topic suffix configuration.
  ## If the section is omitted, no suffix is used.
  ## Following topic suffix methods are supported:
  ##   measurement - suffix equals to measurement's name
  ##   tag         - suffix equals to specified tag's value
  ##   tags        - suffix equals to specified tags' values
  ##                 interleaved with key_separator

  ## Suffix equals to measurement name to topic
  # [outputs.kafka.topic_suffix]
  #   method = "measurement"

  ## Suffix equals to measurement's "foo" tag value.
  ##   If there's no such a tag, suffix equals to an empty string
  # [outputs.kafka.topic_suffix]
  #   method = "tag"
  #   key = "foo"

  ## Suffix equals to measurement's "foo" and "bar"
  ##   tag values, separated by "_". If there is no such tags,
  ##   their values treated as empty strings.
  # [outputs.kafka.topic_suffix]
  #   method = "tags"
  #   keys = ["foo", "bar"]
  #   key_separator = "_"

  ## Telegraf tag to use as a routing key
  ##  ie, if this tag exists, its value will be used as the routing key
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

  ## Optional SASL Config
  # sasl_username = "kafka"
  # sasl_password = "secret"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

func GetTopicSuffixMethodUID(method string) (uint8, error) {
	methodUID, ok := TopicSuffixMethodStringToUID[method]
	if !ok {
		return 0, fmt.Errorf("Unkown topic suffix method provided: %s", method)
	}
	return methodUID, nil
}

func (k *Kafka) GetTopicName(metric telegraf.Metric) string {
	var topicName string
	switch k.topicSuffixMethodUID {
	case TOPIC_SUFFIX_METHOD_MEASUREMENT:
		topicName = k.Topic + metric.Name()
	case TOPIC_SUFFIX_METHOD_TAG:
		topicName = k.Topic + metric.Tags()[k.TopicSuffix.Key]
	case TOPIC_SUFFIX_METHOD_TAGS:
		var tags_values []string
		for _, tag := range k.TopicSuffix.Keys {
			tags_values = append(tags_values, metric.Tags()[tag])
		}
		topicName = k.Topic + strings.Join(tags_values, k.TopicSuffix.KeySeparator)
	default:
		topicName = k.Topic
	}
	return topicName
}

func (k *Kafka) SetSerializer(serializer serializers.Serializer) {
	k.serializer = serializer
}

func (k *Kafka) Connect() error {
	topicSuffixMethod, err := GetTopicSuffixMethodUID(k.TopicSuffix.Method)
	if err != nil {
		return err
	}
	k.topicSuffixMethodUID = topicSuffixMethod

	config := sarama.NewConfig()

	config.Producer.RequiredAcks = sarama.RequiredAcks(k.RequiredAcks)
	config.Producer.Compression = sarama.CompressionCodec(k.CompressionCodec)
	config.Producer.Retry.Max = k.MaxRetry
	config.Producer.Return.Successes = true

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

	if k.SASLUsername != "" && k.SASLPassword != "" {
		config.Net.SASL.User = k.SASLUsername
		config.Net.SASL.Password = k.SASLPassword
		config.Net.SASL.Enable = true
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
		buf, err := k.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		topicName := k.GetTopicName(metric)

		m := &sarama.ProducerMessage{
			Topic: topicName,
			Value: sarama.ByteEncoder(buf),
		}
		if h, ok := metric.Tags()[k.RoutingTag]; ok {
			m.Key = sarama.StringEncoder(h)
		}

		_, _, err = k.producer.SendMessage(m)

		if err != nil {
			return fmt.Errorf("FAILED to send kafka message: %s\n", err)
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
