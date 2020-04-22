package pulsar

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	"log"
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
		URL             string `toml:"url"`
		Topic           string `toml:"topic"`
		TopicTag        string `toml:"topic_tag"`
		ExcludeTopicTag bool   `toml:"exclude_topic_tag"`
		RoutingTag      string `toml:"routing_tag"`
		RoutingKey      string `toml:"routing_key"`

		//Client Options
		AuthProvider               string `toml:"auth_provider"`
		TLSAllowInsecureConnection bool   `toml:"insecure_skip_verify"`
		TLSTrustCertsFilePath      string `toml:"tls_ca"`
		TLSValidateHostname        bool   `toml:"tls_validate_host_name"`
		TLSCertificatePath         string `toml:"tls_cert"`
		TLSPrivateKeyPath          string `toml:"tls_key"`
		AuthToken                  string `toml:"auth_token"`
		ConnectionTimeout          string `toml:"connection_timeout"`
		OperationTimeout           string `toml:"operation_timeout"`

		//Producer Options
		CompressionType         int    `toml:"compression_type"`
		MaxPendingMessages      int    `toml:"max_pending_messages"`
		HashingScheme           string `toml:"hashing_scheme"`
		BatchingMaxPublishDelay string `toml:"batching_max_publish_delay"`
		BatchingMaxMessages     uint   `toml:"batching_max_messages"`

		tenantNameSpace string
		producerCache   map[string]pulsar.Producer

		Log telegraf.Logger `toml:"-"`

		client pulsar.Client

		serializer serializers.Serializer
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
 ## if set to "tls" the please mention tls_cert and tls_key
 # auth_provider = ""

 # For token auth provider
 # auth_token = ""

 # Set the following values for tls auth provider
 # insecure_skip_verify = false
 # tls_ca = ""
 # tls_validate_host_name = true
 # tls_cert = ""
 # tls_key = ""

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

 ## BatchingMaxMessages set the maximum number of messages permitted in a batch. (default: metric_batch_size)
 ## If set to a value greater than 1, messages will be queued until this threshold is reached or
 ## batch interval has elapsed. By Default it is set to "metric_batch_size"
 # batching_max_messages = 1000

 ## HashingScheme change the "HashingScheme" used to chose the partition on where to publish a particular message.
 ##  JavaStringHash : JavaStringHash Hshing
 ##  Murmur3_32Hash : Murmur3_32Hash Hashing
 # hashing_scheme = "JavaStringHash"


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
	return metric, topic
}

func (p *Pulsar) GetProducer(topic string, metricBatchSize uint) (pulsar.Producer, error) {
	if producer, ok := p.producerCache[topic]; !ok {
		producerOptions := pulsar.ProducerOptions{
			Topic:               topic,
			BatchingMaxMessages: metricBatchSize,
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
func sendAsyncCallback(id pulsar.MessageID, message *pulsar.ProducerMessage, err error) {
	if err != nil {
		log.Printf("Failed to publish metrics : %s", err.Error())
	}

}

func (p *Pulsar) Write(metrics []telegraf.Metric) error {
	var batchSize uint
	// Setting the batch size to metric_batch_size is unset
	if p.BatchingMaxMessages == 0 {
		batchSize = uint(len(metrics))
	} else {
		batchSize = p.BatchingMaxMessages
	}
	for _, metric := range metrics {
		metric, topic := p.GetTopicName(metric)
		producer, err := p.GetProducer(topic, batchSize)
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
			producer.SendAsync(context.Background(), &m, sendAsyncCallback)
		} else {
			p.Log.Errorf("Unable to create producer for topic %s : %s", topic, err.Error())
			return err
		}

	}
	// We flush and make sure all the metrics are published and acknowledged by the broker
	// so that in case of failure the metrics are retried
	for topic, producer := range p.producerCache {
		err := producer.Flush()
		if err != nil {
			p.Log.Errorf("Failed to publish the metrics in topic %s : %s ", topic, err.Error())
			//return error on first occurrence of  flush error
			return err
		}
	}

	return nil
}

func init() {
	outputs.Add("pulsar", func() telegraf.Output {
		return &Pulsar{
			BatchingMaxPublishDelay: "1000ms",
		}
	})
}
