package kafka_consumer

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

const (
	defaultMaxUndeliveredMessages = 1000
)

type empty struct{}
type semaphore chan empty

type Consumer interface {
	Errors() <-chan error
	Messages() <-chan *sarama.ConsumerMessage
	MarkOffset(msg *sarama.ConsumerMessage, metadata string)
	Close() error
}

type Kafka struct {
	ConsumerGroup          string   `toml:"consumer_group"`
	ClientID               string   `toml:"client_id"`
	Topics                 []string `toml:"topics"`
	Brokers                []string `toml:"brokers"`
	MaxMessageLen          int      `toml:"max_message_len"`
	Version                string   `toml:"version"`
	MaxUndeliveredMessages int      `toml:"max_undelivered_messages"`
	Offset                 string   `toml:"offset"`
	SASLUsername           string   `toml:"sasl_username"`
	SASLPassword           string   `toml:"sasl_password"`
	TopicTag               string   `toml:"topic_tag"`

	tls.ClientConfig

	cluster Consumer
	parser  parsers.Parser
	wg      *sync.WaitGroup
	cancel  context.CancelFunc

	// Unconfirmed messages
	messages map[telegraf.TrackingID]*sarama.ConsumerMessage

	// doNotCommitMsgs tells the parser not to call CommitUpTo on the consumer
	// this is mostly for test purposes, but there may be a use-case for it later.
	doNotCommitMsgs bool
}

var sampleConfig = `
  ## kafka servers
  brokers = ["localhost:9092"]
  ## topic(s) to consume
  topics = ["telegraf"]
  ## Add topic as tag if topic_tag is not empty
  # topic_tag = ""

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
  ## Maximum length of a message to consume, in bytes (default 0/unlimited);
  ## larger messages are dropped
  max_message_len = 1000000

  ## Maximum messages to read from the broker that have not been written by an
  ## output.  For best throughput set based on the number of metrics within
  ## each message and the size of the output's metric_batch_size.
  ##
  ## For example, if each message from the queue contains 10 metrics and the
  ## output metric_batch_size is 1000, setting this to 100 will ensure that a
  ## full batch is collected and the write is triggered immediately without
  ## waiting until the next flush_interval.
  # max_undelivered_messages = 1000

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
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
	var clusterErr error

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
		log.Printf("I! WARNING: Kafka consumer invalid offset '%s', using 'oldest'",
			k.Offset)
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	if k.cluster == nil {
		k.cluster, clusterErr = cluster.NewConsumer(
			k.Brokers,
			k.ConsumerGroup,
			k.Topics,
			config,
		)

		if clusterErr != nil {
			log.Printf("E! Error when creating Kafka Consumer, brokers: %v, topics: %v",
				k.Brokers, k.Topics)
			return clusterErr
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	k.cancel = cancel

	// Start consumer goroutine
	k.wg = &sync.WaitGroup{}
	k.wg.Add(1)
	go func() {
		defer k.wg.Done()
		k.receiver(ctx, acc)
	}()

	log.Printf("I! Started the kafka consumer service, brokers: %v, topics: %v",
		k.Brokers, k.Topics)
	return nil
}

// receiver() reads all incoming messages from the consumer, and parses them into
// influxdb metric points.
func (k *Kafka) receiver(ctx context.Context, ac telegraf.Accumulator) {
	k.messages = make(map[telegraf.TrackingID]*sarama.ConsumerMessage)

	acc := ac.WithTracking(k.MaxUndeliveredMessages)
	sem := make(semaphore, k.MaxUndeliveredMessages)

	for {
		select {
		case <-ctx.Done():
			return
		case track := <-acc.Delivered():
			<-sem
			k.onDelivery(track)
		case err := <-k.cluster.Errors():
			acc.AddError(err)
		case sem <- empty{}:
			select {
			case <-ctx.Done():
				return
			case track := <-acc.Delivered():
				// Once for the delivered message, once to leave the case
				<-sem
				<-sem
				k.onDelivery(track)
			case err := <-k.cluster.Errors():
				<-sem
				acc.AddError(err)
			case msg := <-k.cluster.Messages():
				err := k.onMessage(acc, msg)
				if err != nil {
					acc.AddError(err)
					<-sem
				}
			}
		}
	}
}

func (k *Kafka) markOffset(msg *sarama.ConsumerMessage) {
	if !k.doNotCommitMsgs {
		k.cluster.MarkOffset(msg, "")
	}
}

func (k *Kafka) onMessage(acc telegraf.TrackingAccumulator, msg *sarama.ConsumerMessage) error {
	if k.MaxMessageLen != 0 && len(msg.Value) > k.MaxMessageLen {
		k.markOffset(msg)
		return fmt.Errorf("Message longer than max_message_len (%d > %d)",
			len(msg.Value), k.MaxMessageLen)
	}

	metrics, err := k.parser.Parse(msg.Value)
	if err != nil {
		return err
	}
	if len(k.TopicTag) > 0 {
		for _, metric := range metrics {
			metric.AddTag(k.TopicTag, msg.Topic)
		}
	}
	id := acc.AddTrackingMetricGroup(metrics)
	k.messages[id] = msg

	return nil
}

func (k *Kafka) onDelivery(track telegraf.DeliveryInfo) {
	msg, ok := k.messages[track.ID()]
	if !ok {
		log.Printf("E! [inputs.kafka_consumer] Could not mark message delivered: %d", track.ID())
		return
	}

	if track.Delivered() {
		k.markOffset(msg)
	}
	delete(k.messages, track.ID())
}

func (k *Kafka) Stop() {
	k.cancel()
	k.wg.Wait()

	if err := k.cluster.Close(); err != nil {
		log.Printf("E! [inputs.kafka_consumer] Error closing consumer: %v", err)
	}
}

func (k *Kafka) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("kafka_consumer", func() telegraf.Input {
		return &Kafka{
			MaxUndeliveredMessages: defaultMaxUndeliveredMessages,
		}
	})
}
