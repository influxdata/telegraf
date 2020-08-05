package kafka_avro

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/avro"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"

	"github.com/Shopify/sarama"
	"github.com/linkedin/goavro"
)

// ValidTopicSuffixMethods - Valid topic suffice methods
var ValidTopicSuffixMethods = []string{
	"",
	"measurement",
	"tags",
}

type (
	// Kafka - the kafka struct
	Kafka struct {
		// Kafka brokers to send metrics to
		Brokers []string
		// Kafka topic
		Topic string
		// Kafka client id
		ClientID string `toml:"client_id"`
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

		// Avro Schema
		avroSchema   string
		avroSchemaID int

		// Avro schema registry
		SchemaRegistry string `toml:"schema_registry"`

		// Additional avro fields
		Loglevel      string `toml:"loglevel"`
		Logtype       string `toml:"type"`
		Servicelevel  string `toml:"servicelevel"`
		PayloadFormat string `toml:"payload_format"`
		PayloadType   string `toml:"payload_type"`
		Source        string `toml:"source"`
		DataCenter    string `toml:"datacenter"`

		// avro serialize config
		IgnoreAvroSerializeErr bool `toml:"ignore_avro_serialize_err"`
		AvroMagicByteRequired  bool `toml:"avro_magic_byte_required"`
	}

	//TopicSuffix - Topic suffix
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
  ## Kafka features and APIs.  Of particular interested, lz4 compression
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

  ## Suffix equals to "__" + measuremeGnt's "foo" tag value.
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

  ## CompressionCodec represents the various compression codecs recognized by
  ## Kafka in messages.
  ##  0 : No compression
  ##  1 : Gzip compression
  ##  2 : Snappy compression
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

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Optional SASL Config
  # sasl_username = "kafka"
  # sasl_password = "secret"

  ## Avro SSL client auth
  # schema_registry_certificate = "/tmp/client.crt"
  # schema_registry_key = "/tmp/client.key"
  # schema_registry_ca = "/tmp/ca.pem"
  # schema_registry = "https://my-kafka-schema-registry.example.com/subjects/LmaEventSchema/versions/1"
  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  # data_format = "influx"
  ## additonal avro fields
  data_format = "influx"
  datacenter = "local"
  payload_format = "influx"
  source = "telegraf"
  ## ignore avro serialize err
  ignore_avro_serialize_err = true
  ## add avro magic byte to msg
  avro_magic_byte_required = true
`

//  ValidateTopicSuffixMethod - validate topic suffix
func ValidateTopicSuffixMethod(method string) error {
	for _, validMethod := range ValidTopicSuffixMethods {
		if method == validMethod {
			return nil
		}
	}
	return fmt.Errorf("Unknown topic suffix method provided: %s", method)
}

//  GetTopicName - Get topic name
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

// SetSerializer - Set serializer
func (k *Kafka) SetSerializer(serializer serializers.Serializer) {
	k.serializer = serializer
}

// Connect - Connect
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

	// set avroSchema
	k.avroSchema, k.avroSchemaID = avro.GetSchema(k.SchemaRegistry)

	return nil
}

//Close - Close connection
func (k *Kafka) Close() error {
	return k.producer.Close()
}

//SampleConfig - Sample config
func (k *Kafka) SampleConfig() string {
	return sampleConfig
}

//Description - Configuration for the Kafka server to send metrics to
func (k *Kafka) Description() string {
	return "Configuration for the Kafka server to send metrics to"
}

func (k *Kafka) avroSerialize(metric telegraf.Metric, payload string) ([]byte, error) {
	schema := k.avroSchema

	output := make(map[string]interface{})
	codec, err := goavro.NewCodec(schema)
	if err != nil {
		fmt.Println(err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	avroSource := make(map[string]interface{})
	avroType := make(map[string]interface{})
	avroLevel := make(map[string]interface{})
	avroPayloadFormat := make(map[string]interface{})
	avroPayload := make(map[string]interface{})
	avroHostname := make(map[string]interface{})
	avroDataCenter := make(map[string]interface{})

	avroSource["string"] = k.Source
	avroType["string"] = k.Logtype
	avroLevel["string"] = k.Loglevel
	avroPayloadFormat["string"] = k.PayloadFormat
	avroPayload["string"] = payload
	avroHostname["string"] = hostname
	avroDataCenter["string"] = k.DataCenter

	output["source"] = avroSource
	output["type"] = avroType
	output["loglevel"] = avroLevel
	output["payload_format"] = avroPayloadFormat
	output["payload"] = avroPayload
	output["hostname"] = avroHostname
	output["datacenter"] = avroDataCenter

	serialized, err := codec.BinaryFromNative(nil, output)

	if err != nil {
		return []byte{}, err
	}

	if k.AvroMagicByteRequired {
		bs := make([]byte, 4)
		schemaID := uint32(k.avroSchemaID)
		binary.BigEndian.PutUint32(bs, schemaID)
		hdr := append([]byte{0}, bs...)
		return append(hdr, serialized...), nil
	}
	return serialized, nil
}

func (k *Kafka) Write(metrics []telegraf.Metric) error {
	msgs := make([]*sarama.ProducerMessage, 0, len(metrics))
	for _, metric := range metrics {
		buf, err := k.serializer.Serialize(metric)
		buf, err = k.avroSerialize(metric, string(buf))

		if err != nil {
			if k.IgnoreAvroSerializeErr {
				log.Println("E! Error format message to avro format (" + err.Error() + ")")
				continue
			} else {
				return err
			}
		}

		m := &sarama.ProducerMessage{
			Topic: k.GetTopicName(metric),
			Value: sarama.ByteEncoder(buf),
		}
		if h, ok := metric.GetTag(k.RoutingTag); ok {
			m.Key = sarama.StringEncoder(h)
		}
		msgs = append(msgs, m)
	}
	if len(msgs) > 0 {
		err := k.producer.SendMessages(msgs)
		if err != nil {
			// We could have many errors, return only the first encountered.
			if errs, ok := err.(sarama.ProducerErrors); ok {
				for _, prodErr := range errs {
					return prodErr
				}
			}
			return err
		}
	}

	return nil
}

func init() {
	outputs.Add("kafka_avro", func() telegraf.Output {
		return &Kafka{
			MaxRetry:     3,
			RequiredAcks: -1,
		}
	})
}
