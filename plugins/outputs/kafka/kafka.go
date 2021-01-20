package kafka

import (
	"crypto/tls"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/gofrs/uuid"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/kafka"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

var ValidTopicSuffixMethods = []string{
	"",
	"measurement",
	"tags",
}

var zeroTime = time.Unix(0, 0)

type Kafka struct {
	Brokers         []string    `toml:"brokers"`
	Topic           string      `toml:"topic"`
	TopicTag        string      `toml:"topic_tag"`
	ExcludeTopicTag bool        `toml:"exclude_topic_tag"`
	TopicSuffix     TopicSuffix `toml:"topic_suffix"`
	RoutingTag      string      `toml:"routing_tag"`
	RoutingKey      string      `toml:"routing_key"`

	// Legacy TLS config options
	// TLS client certificate
	Certificate string
	// TLS client key
	Key string
	// TLS certificate authority
	CA string

	kafka.WriteConfig

	Log telegraf.Logger `toml:"-"`

	tlsConfig tls.Config

	producerFunc func(addrs []string, config *sarama.Config) (sarama.SyncProducer, error)
	producer     sarama.SyncProducer

	serializer serializers.Serializer
}

type TopicSuffix struct {
	Method    string   `toml:"method"`
	Keys      []string `toml:"keys"`
	Separator string   `toml:"separator"`
}

// DebugLogger logs messages from sarama at the debug level.
type DebugLogger struct {
}

func (*DebugLogger) Print(v ...interface{}) {
	args := make([]interface{}, 0, len(v)+1)
	args = append(append(args, "D! [sarama] "), v...)
	log.Print(args...)

}

func (*DebugLogger) Printf(format string, v ...interface{}) {
	log.Printf("D! [sarama] "+format, v...)
}

func (*DebugLogger) Println(v ...interface{}) {
	args := make([]interface{}, 0, len(v)+1)
	args = append(append(args, "D! [sarama] "), v...)
	log.Println(args...)
}

var sampleConfig = `
  ## URLs of kafka brokers
  brokers = ["localhost:9092"]
  ## Kafka topic for producer messages
  topic = "telegraf"

  ## The value of this tag will be used as the topic.  If not set the 'topic'
  ## option is used.
  # topic_tag = ""

  ## If true, the 'topic_tag' will be removed from to the metric.
  # exclude_topic_tag = false

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

  ## The routing tag specifies a tagkey on the metric whose value is used as
  ## the message key.  The message key is used to determine which partition to
  ## send the message to.  This tag is prefered over the routing_key option.
  routing_tag = "host"

  ## The routing key is set as the message key and used to determine which
  ## partition to send the message to.  This value is only used when no
  ## routing_tag is set or as a fallback when the tag specified in routing tag
  ## is not found.
  ##
  ## If set to "random", a random value will be generated for each message.
  ##
  ## When unset, no message key is added and each message is routed to a random
  ## partition.
  ##
  ##   ex: routing_key = "random"
  ##       routing_key = "telegraf"
  # routing_key = ""

  ## Compression codec represents the various compression codecs recognized by
  ## Kafka in messages.
  ##  0 : None
  ##  1 : Gzip
  ##  2 : Snappy
  ##  3 : LZ4
  ##  4 : ZSTD
  # compression_codec = 0

  ## Idempotent Writes
  ## If enabled, exactly one copy of each message is written.
  # idempotent_writes = false

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

  ## Optional SASL:
  ## one of: OAUTHBEARER, PLAIN, SCRAM-SHA-256, SCRAM-SHA-512, GSSAPI
  ## (defaults to PLAIN)
  # sasl_mechanism = ""

  ## used if sasl_mechanism is GSSAPI (experimental)
  # sasl_gssapi_service_name = ""
  # ## One of: KRB5_USER_AUTH and KRB5_KEYTAB_AUTH
  # sasl_gssapi_auth_type = "KRB5_USER_AUTH"
  # sasl_gssapi_kerberos_config_path = "/"
  # sasl_gssapi_realm = "realm"
  # sasl_gssapi_key_tab_path = ""
  # sasl_gssapi_disable_pafxfast = false

  ## used if sasl_mechanism is OAUTHBEARER (experimental)
  # sasl_access_token = ""

  ## SASL protocol version.  When connecting to Azure EventHub set to 0.
  # sasl_version = 1

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

func (k *Kafka) GetTopicName(metric telegraf.Metric) (telegraf.Metric, string) {
	topic := k.Topic
	if k.TopicTag != "" {
		if t, ok := metric.GetTag(k.TopicTag); ok {
			topic = t

			// If excluding the topic tag, a copy is required to avoid modifying
			// the metric buffer.
			if k.ExcludeTopicTag {
				metric = metric.Copy()
				metric.Accept()
				metric.RemoveTag(k.TopicTag)
			}
		}
	}

	var topicName string
	switch k.TopicSuffix.Method {
	case "measurement":
		topicName = topic + k.TopicSuffix.Separator + metric.Name()
	case "tags":
		var topicNameComponents []string
		topicNameComponents = append(topicNameComponents, topic)
		for _, tag := range k.TopicSuffix.Keys {
			tagValue := metric.Tags()[tag]
			if tagValue != "" {
				topicNameComponents = append(topicNameComponents, tagValue)
			}
		}
		topicName = strings.Join(topicNameComponents, k.TopicSuffix.Separator)
	default:
		topicName = topic
	}
	return metric, topicName
}

func (k *Kafka) SetSerializer(serializer serializers.Serializer) {
	k.serializer = serializer
}

func (k *Kafka) Init() error {
	err := ValidateTopicSuffixMethod(k.TopicSuffix.Method)
	if err != nil {
		return err
	}
	config := sarama.NewConfig()

	if err := k.SetConfig(config); err != nil {
		return err
	}

	// Legacy support ssl config
	if k.Certificate != "" {
		k.TLSCert = k.Certificate
		k.TLSCA = k.CA
		k.TLSKey = k.Key
	}

	producer, err := k.producerFunc(k.Brokers, config)
	if err != nil {
		return err
	}
	k.producer = producer
	return nil
}

func (k *Kafka) Connect() error {
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

func (k *Kafka) routingKey(metric telegraf.Metric) (string, error) {
	if k.RoutingTag != "" {
		key, ok := metric.GetTag(k.RoutingTag)
		if ok {
			return key, nil
		}
	}

	if k.RoutingKey == "random" {
		u, err := uuid.NewV4()
		if err != nil {
			return "", err
		}
		return u.String(), nil
	}

	return k.RoutingKey, nil
}

func (k *Kafka) Write(metrics []telegraf.Metric) error {
	msgs := make([]*sarama.ProducerMessage, 0, len(metrics))
	for _, metric := range metrics {
		metric, topic := k.GetTopicName(metric)

		buf, err := k.serializer.Serialize(metric)
		if err != nil {
			k.Log.Debugf("Could not serialize metric: %v", err)
			continue
		}

		m := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.ByteEncoder(buf),
		}

		// Negative timestamps are not allowed by the Kafka protocol.
		if !metric.Time().Before(zeroTime) {
			m.Timestamp = metric.Time()
		}

		key, err := k.routingKey(metric)
		if err != nil {
			return fmt.Errorf("could not generate routing key: %v", err)
		}

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
					k.Log.Error("Message too large, consider increasing `max_message_bytes`; dropping batch")
					return nil
				}
				if prodErr.Err == sarama.ErrInvalidTimestamp {
					k.Log.Error("The timestamp of the message is out of acceptable range, consider increasing broker `message.timestamp.difference.max.ms`; dropping batch")
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
	sarama.Logger = &DebugLogger{}
	outputs.Add("kafka", func() telegraf.Output {
		return &Kafka{
			WriteConfig: kafka.WriteConfig{
				MaxRetry:     3,
				RequiredAcks: -1,
			},
			producerFunc: sarama.NewSyncProducer,
		}
	})
}
