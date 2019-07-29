package kafka_consumer_legacy

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"

	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kafka/consumergroup"
)

type Kafka struct {
	ConsumerGroup   string
	Topics          []string
	MaxMessageLen   int
	ZookeeperPeers  []string
	ZookeeperChroot string
	Consumer        *consumergroup.ConsumerGroup

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
  ## topic(s) to consume
  topics = ["telegraf"]
  ## an array of Zookeeper connection strings
  zookeeper_peers = ["localhost:2181"]
  ## Zookeeper Chroot
  zookeeper_chroot = ""
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
  max_message_len = 65536
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
	var consumerErr error

	k.acc = acc

	config := consumergroup.NewConfig()
	config.Zookeeper.Chroot = k.ZookeeperChroot
	switch strings.ToLower(k.Offset) {
	case "oldest", "":
		config.Offsets.Initial = sarama.OffsetOldest
	case "newest":
		config.Offsets.Initial = sarama.OffsetNewest
	default:
		log.Printf("I! WARNING: Kafka consumer invalid offset '%s', using 'oldest'\n",
			k.Offset)
		config.Offsets.Initial = sarama.OffsetOldest
	}

	if k.Consumer == nil || k.Consumer.Closed() {
		k.Consumer, consumerErr = consumergroup.JoinConsumerGroup(
			k.ConsumerGroup,
			k.Topics,
			k.ZookeeperPeers,
			config,
		)
		if consumerErr != nil {
			return consumerErr
		}

		// Setup message and error channels
		k.in = k.Consumer.Messages()
		k.errs = k.Consumer.Errors()
	}

	k.done = make(chan struct{})

	// Start the kafka message reader
	go k.receiver()
	log.Printf("I! Started the kafka consumer service, peers: %v, topics: %v\n",
		k.ZookeeperPeers, k.Topics)
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
				k.Consumer.CommitUpto(msg)
				k.Unlock()
			}
		}
	}
}

func (k *Kafka) Stop() {
	k.Lock()
	defer k.Unlock()
	close(k.done)
	if err := k.Consumer.Close(); err != nil {
		k.acc.AddError(fmt.Errorf("Error closing consumer: %s\n", err.Error()))
	}
}

func (k *Kafka) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("kafka_consumer_legacy", func() telegraf.Input {
		return &Kafka{}
	})
}
