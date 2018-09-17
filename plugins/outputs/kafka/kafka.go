package kafka

import (
	"crypto/tls"
	"fmt"
	"log"
	"strings"

	"github.com/influxdata/telegraf"
	tlsint "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	uuid "github.com/satori/go.uuid"

	"github.com/Shopify/sarama"
)

var ValidTopicSuffixMethods = []string{
	"",
	"measurement",
	"tags",
}

type (
	Kafka struct {
		Brokers          []string
		Topic            string
		ClientID         string      `toml:"client_id"`
		TopicSuffix      TopicSuffix `toml:"topic_suffix"`
		RoutingTag       string      `toml:"routing_tag"`
		RoutingKey       string      `toml:"routing_key"`
		CompressionCodec int
		RequiredAcks     int
		MaxRetry         int
		MaxMessageBytes  int `toml:"max_message_bytes"`

		Version string `toml:"version"`

		// Legacy TLS config options
		// TLS client certificate
		Certificate string
		// TLS client key
		Key string
		// TLS certificate authority
		CA string

		tlsint.ClientConfig

		// SASL Username
		SASLUsername string `toml:"sasl_username"`
		// SASL Password
		SASLPassword string `toml:"sasl_password"`

		tlsConfig tls.Config
		producer  sarama.SyncProducer

		serializer serializers.Serializer
	}
	TopicSuffix struct {
		Method    string   `toml:"method"`
		Keys      []string `toml:"keys"`
		Separator string   `toml:"separator"`
	}
)

var sampleConfig = `
  ## URLs of kafka brokers
  brokers = ["localhost:9092"]
  ## Kafka topic for producer messages
  topic = "telegraf"

  ## Optional Client id
  # client_id = "Telegraf"

  ## Set the minimal supported Kafka version.  Setting this enables the use of new
  ## Kafka features and APIs.  Of particular interest, lz4 compression
  ## requires at least version 0.10.0.0.
  ##   ex: version = "1.1.0"
  # version = ""

  ## Optional topic suffix configuration.
  ## If the section is omitted, no suffix is used.
  ## Following topic suffix methods are supported:
  ##   measurement - suffix equals to separator + measurement's name
  ##   tags        - suffix equals to separator + specified tags' values
  ##                 interleaved with separator

  ## Suffix equals to "_" + measurement name
  # [outputs.kafka.topic_suffix]
  #   method = "measurement"
  #   separator = "_"

  ## Suffix equals to "__" + measurement's "foo" tag value.
  ##   If there's no such a tag, suffix equals to an empty string
  # [outputs.kafka.topic_suffix]
  #   method = "tags"
  #   keys = ["foo"]
  #   separator = "__"

  ## Suffix equals to "_" + measurement's "foo" and "bar"
  ##   tag values, separated by "_". If there is no such tags,
  ##   their values treated as empty strings.
  # [outputs.kafka.topic_suffix]
  #   method = "tags"
  #   keys = ["foo", "bar"]
  #   separator = "_"

  ## Telegraf tag to use as a routing key
  ##  ie, if this tag exists, its value will be used as the routing key
  routing_tag = "host"

  ## Static routing key.  Used when no routing_tag is set or as a fallback
  ## when the tag specified in routing tag is not found.  If set to "random",
  ## a random value will be generated for each message.
  ##   ex: routing_key = "random"
  ##       routing_key = "telegraf"
  # routing_key = ""

  ## CompressionCodec represents the various compression codecs recognized by
  ## Kafka in messages.
  ##  0 : No compression
  ##  1 : Gzip compression
  ##  2 : Snappy compression
  ##  3 : LZ4 compression
  # compression_codec = 0

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
  # required_acks = -1

  ## The maximum number of times to retry sending a metric before failing
  ## until the next flush.
  # max_retry = 3

  ## The maximum permitted size of a message. Should be set equal to or
  ## smaller than the broker's 'message.max.bytes'.
  # max_message_bytes = 1000000

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Optional SASL Config
  # sasl_username = "kafka"
  # sasl_password = "secret"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "influx"
`

func ValidateTopicSuffixMethod(method string) error {
	for _, validMethod := range ValidTopicSuffixMethods {
		if method == validMethod {
			return nil
		}
	}
	return fmt.Errorf("Unknown topic suffix method provided: %s", method)
}

func (k *Kafka) GetTopicName(metric telegraf.Metric) string {
	var topicName string
	switch k.TopicSuffix.Method {
	case "measurement":
		topicName = k.Topic + k.TopicSuffix.Separator + metric.Name()
	case "tags":
		var topicNameComponents []string
		topicNameComponents = append(topicNameComponents, k.Topic)
		for _, tag := range k.TopicSuffix.Keys {
			tagValue := metric.Tags()[tag]
			if tagValue != "" {
				topicNameComponents = append(topicNameComponents, tagValue)
			}
		}
		topicName = strings.Join(topicNameComponents, k.TopicSuffix.Separator)
	default:
		topicName = k.Topic
	}
	return topicName
}

func (k *Kafka) SetSerializer(serializer serializers.Serializer) {
	k.serializer = serializer
}

func (k *Kafka) Connect() error {
	err := ValidateTopicSuffixMethod(k.TopicSuffix.Method)
	if err != nil {
		return err
	}
	config := sarama.NewConfig()

	if k.Version != "" {
		version, err := sarama.ParseKafkaVersion(k.Version)
		if err != nil {
			return err
		}
		config.Version = version
	}

	if k.ClientID != "" {
		config.ClientID = k.ClientID
	} else {
		config.ClientID = "Telegraf"
	}

	config.Producer.RequiredAcks = sarama.RequiredAcks(k.RequiredAcks)
	config.Producer.Compression = sarama.CompressionCodec(k.CompressionCodec)
	config.Producer.Retry.Max = k.MaxRetry
	config.Producer.Return.Successes = true

	if k.MaxMessageBytes > 0 {
		config.Producer.MaxMessageBytes = k.MaxMessageBytes
	}

	// Legacy support ssl config
	if k.Certificate != "" {
		k.TLSCert = k.Certificate
		k.TLSCA = k.CA
		k.TLSKey = k.Key
	}

	tlsConfig, err := k.ClientConfig.TLSConfig()
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

func (k *Kafka) routingKey(metric telegraf.Metric) string {
	if k.RoutingTag != "" {
		key, ok := metric.GetTag(k.RoutingTag)
		if ok {
			return key
		}
	}

	if k.RoutingKey == "random" {
		u := uuid.NewV4()
		return u.String()
	}

	return k.RoutingKey
}

func (k *Kafka) Write(metrics []telegraf.Metric) error {
	msgs := make([]*sarama.ProducerMessage, 0, len(metrics))
	for _, metric := range metrics {
		buf, err := k.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		m := &sarama.ProducerMessage{
			Topic: k.GetTopicName(metric),
			Value: sarama.ByteEncoder(buf),
		}
		key := k.routingKey(metric)
		if key != "" {
			m.Key = sarama.StringEncoder(key)
		}
		msgs = append(msgs, m)
	}

	err := k.producer.SendMessages(msgs)
	if err != nil {
		// We could have many errors, return only the first encountered.
		if errs, ok := err.(sarama.ProducerErrors); ok {
			for _, prodErr := range errs {
				if prodErr.Err == sarama.ErrMessageSizeTooLarge {
					log.Printf("E! Error writing to output [kafka]: Message too large, consider increasing `max_message_bytes`; dropping batch")
					return nil
				}
				return prodErr
			}
		}
		return err
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
