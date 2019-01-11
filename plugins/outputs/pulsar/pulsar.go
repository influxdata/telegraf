package pulsar

import (
	"context"
	"fmt"
	"time"

	plsr "github.com/apache/incubator-pulsar/pulsar-client-go/pulsar"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

var sampleConfig = `
  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"

  ## URL to Pulsar cluster
  ## If you use SSL, then the protocol should be "pulsar+ssl"
  url = "pulsar://localhost:6650"

  ## Number of threads to be used for handling connections to brokers
  # iothreads = 1

  ## Producer-create, subscribe and unsubscribe operations will be retried until this interval, after which the
  ## operation will be marked as failed
  # operation_timeout = "30s"

  ## Set the number of threads to be used for message listeners
  # message_listener_threads = 1

  ## Number of concurrent lookup-requests allowed to send on each broker-connection to prevent overload on broker.
  ## It should be configured with higher value only in case of it requires to produce/subscribe
  ## on thousands of topic using created Pulsar Client
  # concurrent_lookup_requests = 5000

  ## Set the path to the trusted TLS certificate file
  # tls_trust_certs_path = ""

  ## Configure whether the Pulsar client accept untrusted TLS certificate from broker
  # tls_allow_insecure_connection = false

  ## Set the interval between each stat info. Stats will be activated with positive
  ## stats_interval_in_seconds It should be set to at least 1 second
  # stats_interval = 60

  ## Configure the authentication provider.
  # [auth]

  ## Create new Athenz Authentication provider with configuration in JSON form
  #   athenz = ""

  ## Create new Authentication provider with specified TLS certificate and private key
  #   cert_path = ""
  #   key_path = ""

  ## Configure the producer
  [producer]

  ## Set topic of the message, required
	topic = ""

  ## Specify a name for the producer
  ## If not assigned, the system will generate a globally unique name.
  ## When specifying a name, it is up to the user to ensure that, for a given topic, the producer name is unique
  ## across all Pulsar's clusters. Brokers will enforce that only a single producer a given name can be publishing on
  ## a topic.
  #    name = ""
  
  ## Attach a set of application defined properties to the producer
  ## This properties will be visible in the topic stats
  #    properties = { foo = "bar" }

  ## Set the send timeout.
  ## If a message is not acknowledged by the server before the send_timeout expires, an error will be reported.
  #    send_timeout = "30s"

  ## Set the max size of the queue holding the messages pending to receive an acknowledgment from the broker.
  ## When the queue is full, by default, all calls will fail unless block_if_queue_full is set to true.
  #    max_pending_messages = 64

  ## Set the number of max pending messages across all the partitions
  ## This setting will be used to lower the max pending messages for each partition, if the total exceeds the configured value.
  #    max_pending_messages_across_partitions = 512

  ## Set whether the send operations should block when the outgoing message queue is full. If set to false, send operations will immediately fail
  ## when there is no space left in pending queue.
  #    block_if_queue_full = false

  ## Set the message routing mode for the partitioned producer.
  ## 0 = Round robin
  ## 1 = Use single partition
  ## 2 = Custom partition
  #    message_routing_mode = 0

  ## Change the hashing scheme used to chose the partition on where to publish a particular message.
  ## 0 = Java String.hashCode() equivalent
  ## 1 = Use Murmur3 hashing function
  ## 2 = C++ based boost::hash
  #    hashing_scheme = 0

  ## Set the compression type for the producer.
  ## 0 = No compression
  ## 1 = LZ4
  ## 2 = ZLIB
  #    compression_type = 0

  ## Control whether automatic batching of messages is enabled for the producer.
  #    batching = false

  ## Set the time period within which the messages sent will be batched if batch messages are
  ## enabled. If set, messages will be queued until this time interval or until.
  #    batching_max_publish_delay = "10ms"

  ## Set the maximum number of messages permitted in a batch. If set,
  ## messages will be queued until this threshold is reached or batch interval has elapsed
  #    batching_max_messages = 1000
`

type Pulsar struct {
	serializer serializers.Serializer
	client     plsr.Client
	producer   plsr.Producer

	URL                        string `toml:"url"`
	IOThreads                  int    `toml:"iothreads,omitempty"`
	OperationTimeout           string `toml:"operation_timeout,omitempty"`
	MessageListenerThreads     int    `toml:"message_listener_threads,omitempty"`
	ConcurrentLookupRequests   int    `toml:"concurrent_lookup_requests,omitempty"`
	TLSTrustCertsPath          string `toml:"tls_trust_certs_path,omitempty"`
	TLSAllowInsecureConnection bool   `toml:"tls_allow_insecure_connection,omitempty"`
	StatsIntervalInSeconds     int    `toml:"stats_interval,omitempty"`

	Auth     *AuthOpts     `toml:"auth,omitempty"`
	Producer *ProducerOpts `toml:"producer"`
}

type AuthOpts struct {
	Athenz   string `toml:"athenz,omitempty"`
	CertPath string `toml:"cert_path,omitempty"`
	KeyPath  string `toml:"key_path,omitempty"`
	Token    string `toml:"token,omitempty"`
}

type ProducerOpts struct {
	Topic                              string            `toml:"topic"`
	Name                               string            `toml:"name,omitempty"`
	Properties                         map[string]string `toml:"properties,omitempty"`
	SendTimeout                        string            `toml:"send_timeout,omitempty"`
	sendTimeout                        time.Duration
	MaxPendingMessages                 int  `toml:"max_pending_messages,omitempty"`
	MaxPendingMessagesAcrossPartitions int  `toml:"max_pending_messages_across_partitions,omitempty"`
	BlockIfQueueFull                   bool `toml:"block_if_queue_full,omitempty"`
	MessageRoutingMode                 int  `toml:"message_routing_mode,omitempty"`
	messageRoutingMode                 plsr.MessageRoutingMode
	HashingScheme                      int `toml:"hashing_scheme,omitempty"`
	hashingScheme                      plsr.HashingScheme
	CompressionType                    int `toml:"compression_type,omitempty"`
	compressionType                    plsr.CompressionType
	Batching                           bool   `toml:"batching,omitempty"`
	BatchingMaxPublishDelay            string `toml:"batching_max_publish_delay,omitempty"`
	batchingMaxPublishDelay            time.Duration
	BatchingMaxMessages                uint `toml:"batching_max_messages,omitempty"`
}

func (p *Pulsar) SetSerializer(serializer serializers.Serializer) {
	p.serializer = serializer
}

func (p *Pulsar) Connect() error {
	var err error

	conf := plsr.ClientOptions{}
	conf.URL = p.URL

	// General
	if p.IOThreads > 0 {
		conf.IOThreads = p.IOThreads
	}
	if p.OperationTimeout != "" {
		conf.OperationTimeoutSeconds, err = time.ParseDuration(p.OperationTimeout)
		if err != nil {
			return err
		}
	}
	if p.MessageListenerThreads > 0 {
		conf.MessageListenerThreads = p.MessageListenerThreads
	}
	if p.ConcurrentLookupRequests > 0 {
		conf.ConcurrentLookupRequests = p.ConcurrentLookupRequests
	}
	if p.TLSTrustCertsPath != "" {
		conf.TLSTrustCertsFilePath = p.TLSTrustCertsPath
	}
	if p.TLSAllowInsecureConnection {
		conf.TLSAllowInsecureConnection = p.TLSAllowInsecureConnection
	}
	if p.StatsIntervalInSeconds > 0 {
		conf.StatsIntervalInSeconds = p.StatsIntervalInSeconds
	}

	// Auth
	if p.Auth != nil {
		if p.Auth.Athenz != "" {
			conf.Authentication = plsr.NewAuthenticationAthenz(p.Auth.Athenz)
		}
		if p.Auth.CertPath != "" && p.Auth.KeyPath != "" {
			conf.Authentication = plsr.NewAuthenticationTLS(p.Auth.CertPath, p.Auth.KeyPath)
		}
	}

	if p.Producer == nil {
		err = fmt.Errorf("producer options cannot be empty, at least topic is needed")
		return err
	}

	// Producer
	if p.Producer.SendTimeout != "" {
		p.Producer.sendTimeout, err = time.ParseDuration(p.Producer.SendTimeout)
	}
	if p.Producer.MessageRoutingMode > 0 {
		p.Producer.messageRoutingMode = plsr.MessageRoutingMode(p.Producer.MessageRoutingMode)
	}
	if p.Producer.HashingScheme > 0 {
		p.Producer.hashingScheme = plsr.HashingScheme(p.Producer.HashingScheme)
	}
	if p.Producer.CompressionType > 0 {
		p.Producer.compressionType = plsr.CompressionType(p.Producer.CompressionType)
	}
	if p.Producer.BatchingMaxPublishDelay != "" {
		p.Producer.batchingMaxPublishDelay, err = time.ParseDuration(p.Producer.BatchingMaxPublishDelay)
		if err != nil {
			return err
		}
	}

	p.client, err = plsr.NewClient(conf)
	if err != nil {
		return err
	}

	p.producer, err = p.client.CreateProducer(plsr.ProducerOptions{
		Topic:                              p.Producer.Topic,
		Name:                               p.Producer.Name,
		Properties:                         p.Producer.Properties,
		SendTimeout:                        p.Producer.sendTimeout,
		MaxPendingMessages:                 p.Producer.MaxPendingMessages,
		MaxPendingMessagesAcrossPartitions: p.Producer.MaxPendingMessagesAcrossPartitions,
		BlockIfQueueFull:                   p.Producer.BlockIfQueueFull,
		MessageRoutingMode:                 p.Producer.messageRoutingMode,
		HashingScheme:                      p.Producer.hashingScheme,
		CompressionType:                    p.Producer.compressionType,
		Batching:                           p.Producer.Batching,
		BatchingMaxPublishDelay:            p.Producer.batchingMaxPublishDelay,
		BatchingMaxMessages:                p.Producer.BatchingMaxMessages,
	})
	return err
}

func (p *Pulsar) Close() error {
	return p.client.Close()
}

func (p *Pulsar) SampleConfig() string {
	return sampleConfig
}

func (p *Pulsar) Description() string {
	return "Send telegraf measurements to Apache Pulsar"
}

func (p *Pulsar) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		buf, err := p.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		ctx := context.Background()
		err = p.producer.Send(ctx, plsr.ProducerMessage{
			Payload: buf,
			// Key: ,
			// Properties: ,
			// EventTime: ,
			// ReplicationClusters: ,
		})
		if err != nil {
			return fmt.Errorf("FAILED to send Pulsar message: %s", err)
		}
	}
	return nil
}

func init() {
	outputs.Add("pulsar", func() telegraf.Output {
		return &Pulsar{}
	})
}
