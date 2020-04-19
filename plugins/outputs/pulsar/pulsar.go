package pulsar

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	"strings"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

var ValidTopicSuffixMethods = []string{
	"",
	"measurement",
	"tags",
}

var zeroTime = time.Unix(0, 0)

type (
	Pulsar struct {
		URL             string      `toml:"url"`
		Topic           string      `toml:"topic"`
		TopicTag        string      `toml:"topic_tag"`
		ExcludeTopicTag bool        `toml:"exclude_topic_tag"`
		TopicSuffix     TopicSuffix `toml:"topic_suffix"`
		RoutingTag      string      `toml:"routing_tag"`
		RoutingKey      string      `toml:"routing_key"`

		//Client Options
		AuthProvider               string `toml:"auth_provider"`
		TLSAllowInsecureConnection bool   `toml:"tls_allow_insecure_connection"`
		TLSTrustCertsFilePath      string `toml:"tls_trust_certs_file_path"`
		TLSValidateHostname        bool   `toml:"tls_validate_host_name"`
		TLSCertificatePath         string `toml:"tls_certificate_path"`
		TLSPrivateKeyPath          string `toml:"tls_private_key_path"`
		AuthToken                  string `toml:"auth_token"`
		ConnectionTimeout          string `toml:"connection_timeout"`
		OperationTimeout           string `toml:"operation_timeout"`

		//Producer Options
		CompressionType         int    `toml:"compression_type"`
		MaxPendingMessages      int    `toml:"max_pending_messages"`
		HashingScheme           string `toml:"hashing_scheme"`
		DisableBatching         bool   `toml:"disable_batching"`
		BatchingMaxPublishDelay string `toml:"batching_max_publish_delay"`
		BatchingMaxMessages     uint   `toml:"batching_max_messages"`

		tenantNameSpace string
		producerCache   map[string]pulsar.Producer

		Log telegraf.Logger `toml:"-"`

		client pulsar.Client

		serializer serializers.Serializer
	}
	TopicSuffix struct {
		Method    string   `toml:"method"`
		Keys      []string `toml:"keys"`
		Separator string   `toml:"separator"`
	}
)

var sampleConfig = `
  [[outputs.pulsar]]
 ## URLs of pulsar url
  url = "pulsar://localhost:6650"
 ## Pulsar topic for producer messages
  topic = "persistent://public/default/telegraf"

 ## The value of this tag will be used as the topic.  If not set the 'topic'
 ## option is used.
 # topic_tag = "foo"

 ## If true, the 'topic_tag' will be removed from to the metric.
 # exclude_topic_tag = false

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

 ## Optional Authentication Provider Config Defaults to empty "" NoAuthentication
 ## if set to "token" provide the JWT token
 ## if set to "tls" the please mention tls_certificate_path and tls_private_key_path
 # auth_provider = ""

 # For token auth provider
 # auth_token = ""

 # Set the following values for tls auth provider
 # tls_allow_insecure_connection = false
 # tls_trust_certs_file_path = ""
 # tls_validate_host_name = true
 # tls_certificate_path = ""
 # tls_private_key_path = ""

 ## Optional timeout Config
 # connection_timeout = "30s"
 # operation_timeout = "30s"

 ## Optional Producer Config

 ## CompressionType represents the various compression codecs recognized by
 ## Pulsar in messages.
 ##  0 : No compression
 ##  1 : LZ4 compression
 ##  2 : ZLib compression
 ##  3 : ZSTD compression
 # compression_type = 0

 ## MaxPendingMessages set the max size of the queue holding the messages pending to receive an
 ## acknowledgment from the broker.
 # max_pending_messages = 1000

 ## BatchingMaxPublishDelay set the time period within which the messages sent will be batched (default: 10ms)
 ## if batch messages are enabled. If set to a non zero value, messages will be queued until this time
 ## interval or until
 # batching_max_publish_delay = "10ms"

 ## BatchingMaxMessages set the maximum number of messages permitted in a batch. (default: 1000)
 ## If set to a value greater than 1, messages will be queued until this threshold is reached or
 ## batch interval has elapsed.
 # batching_max_messages = 1000

 ## HashingScheme change the "HashingScheme" used to chose the partition on where to publish a particular message.
 ##  JavaStringHash : JavaStringHash Hshing
 ##  Murmur3_32Hash : Murmur3_32Hash Hashing
 # hashing_scheme = "JavaStringHash"

 ## Disable batching will reduce the throughput
 # disable_batching = false


 ## Data format to output.
 ## Each data format has its own unique set of configuration options, read
 ## more about them here:
 ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
 # data_format = "influx"

 ## Optional topic suffix configuration.
 ## If the section is omitted, no suffix is used.
 ## Following topic suffix methods are supported:
 ##   measurement - suffix equals to separator + measurement's name
 ##   tags        - suffix equals to separator + specified tags' values
 ##                 interleaved with separator
 ## The routing tag specifies a tagkey on the metric whose value is used as
 ## the message key.  The message key is used to determine which partition to
 ## send the message to.  This tag is prefered over the routing_key option.

 ## Suffix equals to "_" + measurement name
 #  [outputs.pulsar.topic_suffix]
 #    method = "measurement"
 #    separator = "-"

 ## Suffix equals to "__" + measurement's "foo" tag value.
 ##   If there's no such a tag, suffix equals to an empty string
 #  [outputs.pulsar.topic_suffix]
 #    method = "tags"
 #    keys = ["foo"]
 #    separator = "-"

 ## Suffix equals to "_" + measurement's "foo" and "bar"
 ##   tag values, separated by "_". If there is no such tags,
 ##   their values treated as empty strings.
 #  [outputs.pulsar.topic_suffix]
 #    method = "tags"
 #    keys = ["foo","bar"]
 #    separator = "-"

`

func ValidateTopicSuffixMethod(method string) error {
	for _, validMethod := range ValidTopicSuffixMethods {
		if method == validMethod {
			return nil
		}
	}
	return fmt.Errorf("Unknown topic suffix method provided: %s", method)
}

func (p *Pulsar) GetTopicName(metric telegraf.Metric) (telegraf.Metric, string) {
	topic := p.Topic
	if p.TopicTag != "" {
		if t, ok := metric.GetTag(p.TopicTag); ok {
			topic = p.tenantNameSpace + "/" + t
			// If excluding the topic tag, a copy is required to avoid modifying
			// the metric buffer.
			if p.ExcludeTopicTag {
				metric = metric.Copy()
				metric.Accept()
				metric.RemoveTag(p.TopicTag)
			}
		}
	}

	var topicName string
	switch p.TopicSuffix.Method {
	case "measurement":
		topicName = topic + p.TopicSuffix.Separator + metric.Name()
	case "tags":
		var topicNameComponents []string
		topicNameComponents = append(topicNameComponents, topic)
		for _, tag := range p.TopicSuffix.Keys {
			tagValue := metric.Tags()[tag]
			if tagValue != "" {
				topicNameComponents = append(topicNameComponents, tagValue)
			}
		}
		topicName = strings.Join(topicNameComponents, p.TopicSuffix.Separator)
	default:
		topicName = topic
	}
	return metric, topicName
}

func (p *Pulsar) GetProducer(topic string) (pulsar.Producer, error) {
	if producer, ok := p.producerCache[topic]; !ok {
		producerOptions := pulsar.ProducerOptions{
			Topic: topic,
		}
		switch p.CompressionType {
		case 0:
			producerOptions.CompressionType = pulsar.NoCompression
		case 1:
			producerOptions.CompressionType = pulsar.LZ4
		case 2:
			producerOptions.CompressionType = pulsar.ZLib
		case 4:
			producerOptions.CompressionType = pulsar.ZSTD
		default:
			producerOptions.CompressionType = pulsar.NoCompression
		}

		switch p.HashingScheme {
		case "Murmur3_32Hash":
			producerOptions.HashingScheme = pulsar.Murmur3_32Hash
		case "JavaStringHash":
			producerOptions.HashingScheme = pulsar.JavaStringHash
		default:
			producerOptions.HashingScheme = pulsar.JavaStringHash
		}
		if p.MaxPendingMessages > 0 {
			producerOptions.MaxPendingMessages = p.MaxPendingMessages
		}
		if producerOptions.DisableBatching == true {
			producerOptions.DisableBatching = p.DisableBatching
		}
		if p.BatchingMaxMessages > 0 {
			producerOptions.BatchingMaxMessages = p.BatchingMaxMessages
		}

		duration, err := time.ParseDuration(p.BatchingMaxPublishDelay)
		if err != nil {
			producerOptions.BatchingMaxPublishDelay = duration
		}
		producer, err := p.client.CreateProducer(producerOptions)
		if err != nil {
			return nil, err
		} else {
			p.Log.Infof("Created producer for topic %s", topic)
			p.producerCache[topic] = producer
			return producer, nil
		}

	} else {
		p.Log.Debugf("Returning cached producer for topic %s", topic)
		return producer, nil
	}

}
func (p *Pulsar) SetSerializer(serializer serializers.Serializer) {
	p.serializer = serializer
}

func (p *Pulsar) Connect() error {
	p.producerCache = make(map[string]pulsar.Producer)
	splits := strings.Split(p.Topic, "/")
	p.tenantNameSpace = strings.Join(splits[0:4], "/")
	err := ValidateTopicSuffixMethod(p.TopicSuffix.Method)
	if err != nil {
		return err
	}
	clientOptions := pulsar.ClientOptions{
		URL:                        p.URL,
		TLSAllowInsecureConnection: p.TLSAllowInsecureConnection,
		TLSTrustCertsFilePath:      p.TLSTrustCertsFilePath,
		TLSValidateHostname:        p.TLSValidateHostname,
	}
	if p.ConnectionTimeout != "" {
		connectionTimeout, err := time.ParseDuration(p.ConnectionTimeout)
		if err != nil {
			p.Log.Error("Invalid ConnectionTimeout. Setting it to 30s ", err.Error())
			clientOptions.ConnectionTimeout = 30 * time.Second
		} else {
			clientOptions.ConnectionTimeout = connectionTimeout
		}
	}
	if p.OperationTimeout != "" {
		operationTimeout, err := time.ParseDuration(p.OperationTimeout)
		if err != nil {
			p.Log.Error("Invalid OperationTimeout. Setting it to 30s", err.Error())
			clientOptions.OperationTimeout = 30 * time.Second
		} else {
			clientOptions.OperationTimeout = operationTimeout
		}

	}

	if p.AuthProvider == "tls" {
		if p.TLSCertificatePath != "" && p.TLSPrivateKeyPath != "" {
			clientOptions.Authentication = pulsar.NewAuthenticationTLS(p.TLSCertificatePath, p.TLSPrivateKeyPath)
		} else {
			return fmt.Errorf("tls auth specified but tls_certificate_path or tls_private_key_path is miising")
		}

	} else if p.AuthProvider == "token" {
		if p.AuthToken != "" {
			clientOptions.Authentication = pulsar.NewAuthenticationToken(p.AuthToken)
		} else {
			return fmt.Errorf("auth provider is token but auth_token is empty")
		}

	}

	client, err := pulsar.NewClient(clientOptions)
	if err != nil {
		return err
	}
	p.client = client
	return nil
}

func (p *Pulsar) Close() error {
	if p.client != nil {
		p.client.Close()
	}
	return nil
}

func (p *Pulsar) SampleConfig() string {
	return sampleConfig
}

func (p *Pulsar) Description() string {
	return "Configuration to send metrics to pulsar broker"
}

func (p *Pulsar) routingKey(metric telegraf.Metric) string {
	if p.RoutingTag != "" {
		key, ok := metric.GetTag(p.RoutingTag)
		if ok {
			return key
		}
	}
	if p.RoutingKey == "random" {
		u, err := uuid.NewV4()
		if err != nil {
			p.Log.Errorf("Unable to generate uudid %s", err.Error())
			return ""
		}
		return u.String()
	}
	return p.RoutingKey
}

func (p *Pulsar) Write(metrics []telegraf.Metric) error {

	for _, metric := range metrics {
		metric, topic := p.GetTopicName(metric)
		producer, err := p.GetProducer(topic)
		if err == nil {
			buf, err := p.serializer.Serialize(metric)
			if err != nil {
				p.Log.Debugf("Could not serialize metric: %v", err)
				continue
			}
			m := pulsar.ProducerMessage{
				Payload:   buf,
				EventTime: metric.Time(),
				Key:       p.routingKey(metric),
			}

			// Negative timestamps are not allowed by the Pulsar protocol.
			if !metric.Time().Before(zeroTime) {
				m.EventTime = metric.Time()
			}
			_, err = producer.Send(context.Background(), &m)
			if err != nil {
				p.Log.Errorf("Unable to publish the message %s: dropping metrics", err.Error())
				continue

			}

		} else {
			p.Log.Errorf("Unable to create producer for topic %s : %s", topic, err.Error())
		}

	}

	return nil
}

func init() {
	outputs.Add("pulsar", func() telegraf.Output {
		return &Pulsar{}
	})
}
