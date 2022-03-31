package kafka_consumer_legacy

import (
	"fmt"
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

	Log telegraf.Logger

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
		k.Log.Infof("WARNING: Kafka consumer invalid offset '%s', using 'oldest'\n",
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
	k.Log.Infof("Started the kafka consumer service, peers: %v, topics: %v\n",
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
				k.acc.AddError(fmt.Errorf("consumer Error: %s", err))
			}
		case msg := <-k.in:
			if k.MaxMessageLen != 0 && len(msg.Value) > k.MaxMessageLen {
				k.acc.AddError(fmt.Errorf("message longer than max_message_len (%d > %d)",
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
				err := k.Consumer.CommitUpto(msg)
				k.Unlock()
				if err != nil {
					k.acc.AddError(fmt.Errorf("committing to consumer failed: %v", err))
				}
			}
		}
	}
}

func (k *Kafka) Stop() {
	k.Lock()
	defer k.Unlock()
	close(k.done)
	if err := k.Consumer.Close(); err != nil {
		k.acc.AddError(fmt.Errorf("error closing consumer: %s", err.Error()))
	}
}

func (k *Kafka) Gather(_ telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("kafka_consumer_legacy", func() telegraf.Input {
		return &Kafka{}
	})
}
