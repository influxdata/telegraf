package kafka_consumer

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

type Kafka struct {
	ConsumerGroup string
	ClientID      string `toml:"client_id"`
	Topics        []string
	Brokers       []string
	MaxMessageLen int
	Version       string `toml:"version"`

	Cluster *cluster.Consumer

	tls.ClientConfig

	// SASL Username
	SASLUsername string `toml:"sasl_username"`
	// SASL Password
	SASLPassword string `toml:"sasl_password"`

	// Legacy metric buffer support
	MetricBuffer int
	// TODO remove PointBuffer, legacy support
	PointBuffer int

	Offset string
	parser parsers.Parser

	sync.Mutex

	// channel for all incoming kafka messages
	in <-chan *sarama.ConsumerMessage
	// channel for all kafka consumer errors
	errs <-chan error
	done chan struct{}

	// keep the accumulator internally:
	acc telegraf.Accumulator

	// doNotCommitMsgs tells the parser not to call CommitUpTo on the consumer
	// this is mostly for test purposes, but there may be a use-case for it later.
	doNotCommitMsgs bool
}

var sampleConfig = `
  ## kafka servers
  brokers = ["localhost:9092"]
  ## topic(s) to consume
  topics = ["telegraf"]

  ## Optional Client id
  # client_id = "Telegraf"

  ## Set the minimal supported Kafka version.  Setting this enables the use of new
  ## Kafka features and APIs.  Of particular interest, lz4 compression
  ## requires at least version 0.10.0.0.
  ##   ex: version = "1.1.0"
  # version = ""

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Optional SASL Config
  # sasl_username = "kafka"
  # sasl_password = "secret"

  ## the name of the consumer group
  consumer_group = "telegraf_metrics_consumers"
  ## Offset (must be either "oldest" or "newest")
  offset = "oldest"

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"

  ## Maximum length of a message to consume, in bytes (default 0/unlimited);
  ## larger messages are dropped
  max_message_len = 1000000
`

func (k *Kafka) SampleConfig() string {
	return sampleConfig
}

func (k *Kafka) Description() string {
	return "Read metrics from Kafka topic(s)"
}

func (k *Kafka) SetParser(parser parsers.Parser) {
	k.parser = parser
}

func (k *Kafka) Start(acc telegraf.Accumulator) error {
	k.Lock()
	defer k.Unlock()
	var clusterErr error

	k.acc = acc

	config := cluster.NewConfig()

	if k.Version != "" {
		version, err := sarama.ParseKafkaVersion(k.Version)
		if err != nil {
			return err
		}
		config.Version = version
	}

	config.Consumer.Return.Errors = true

	tlsConfig, err := k.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	if k.ClientID != "" {
		config.ClientID = k.ClientID
	} else {
		config.ClientID = "Telegraf"
	}

	if tlsConfig != nil {
		log.Printf("D! TLS Enabled")
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Enable = true
	}
	if k.SASLUsername != "" && k.SASLPassword != "" {
		log.Printf("D! Using SASL auth with username '%s',",
			k.SASLUsername)
		config.Net.SASL.User = k.SASLUsername
		config.Net.SASL.Password = k.SASLPassword
		config.Net.SASL.Enable = true
	}

	switch strings.ToLower(k.Offset) {
	case "oldest", "":
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	case "newest":
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	default:
		log.Printf("I! WARNING: Kafka consumer invalid offset '%s', using 'oldest'\n",
			k.Offset)
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	if k.Cluster == nil {
		k.Cluster, clusterErr = cluster.NewConsumer(
			k.Brokers,
			k.ConsumerGroup,
			k.Topics,
			config,
		)

		if clusterErr != nil {
			log.Printf("E! Error when creating Kafka Consumer, brokers: %v, topics: %v\n",
				k.Brokers, k.Topics)
			return clusterErr
		}

		// Setup message and error channels
		k.in = k.Cluster.Messages()
		k.errs = k.Cluster.Errors()
	}

	k.done = make(chan struct{})
	// Start the kafka message reader
	go k.receiver()
	log.Printf("I! Started the kafka consumer service, brokers: %v, topics: %v\n",
		k.Brokers, k.Topics)
	return nil
}

// receiver() reads all incoming messages from the consumer, and parses them into
// influxdb metric points.
func (k *Kafka) receiver() {
	for {
		select {
		case <-k.done:
			return
		case err := <-k.errs:
			if err != nil {
				k.acc.AddError(fmt.Errorf("Consumer Error: %s\n", err))
			}
		case msg := <-k.in:
			if k.MaxMessageLen != 0 && len(msg.Value) > k.MaxMessageLen {
				k.acc.AddError(fmt.Errorf("Message longer than max_message_len (%d > %d)",
					len(msg.Value), k.MaxMessageLen))
			} else {
				metrics, err := k.parser.Parse(msg.Value)
				if err != nil {
					k.acc.AddError(fmt.Errorf("Message Parse Error\nmessage: %s\nerror: %s",
						string(msg.Value), err.Error()))
				}
				for _, metric := range metrics {
					k.acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
				}
			}

			if !k.doNotCommitMsgs {
				// TODO(cam) this locking can be removed if this PR gets merged:
				// https://github.com/wvanbergen/kafka/pull/84
				k.Lock()
				k.Cluster.MarkOffset(msg, "")
				k.Unlock()
			}
		}
	}
}

func (k *Kafka) Stop() {
	k.Lock()
	defer k.Unlock()
	close(k.done)
	if err := k.Cluster.Close(); err != nil {
		k.acc.AddError(fmt.Errorf("Error closing consumer: %s\n", err.Error()))
	}
}

func (k *Kafka) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("kafka_consumer", func() telegraf.Input {
		return &Kafka{}
	})
}
