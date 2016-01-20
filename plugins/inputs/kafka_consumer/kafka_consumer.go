package kafka_consumer

import (
	"log"
	"strings"
	"sync"

	"github.com/influxdata/influxdb/models"
	"github.com/influxdata/telegraf/plugins/inputs"

	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kafka/consumergroup"
)

type Kafka struct {
	ConsumerGroup  string
	Topics         []string
	ZookeeperPeers []string
	Consumer       *consumergroup.ConsumerGroup
	PointBuffer    int
	Offset         string

	sync.Mutex

	// channel for all incoming kafka messages
	in <-chan *sarama.ConsumerMessage
	// channel for all kafka consumer errors
	errs <-chan *sarama.ConsumerError
	// channel for all incoming parsed kafka points
	pointChan chan models.Point
	done      chan struct{}

	// doNotCommitMsgs tells the parser not to call CommitUpTo on the consumer
	// this is mostly for test purposes, but there may be a use-case for it later.
	doNotCommitMsgs bool
}

var sampleConfig = `
  # topic(s) to consume
  topics = ["telegraf"]
  # an array of Zookeeper connection strings
  zookeeper_peers = ["localhost:2181"]
  # the name of the consumer group
  consumer_group = "telegraf_metrics_consumers"
  # Maximum number of points to buffer between collection intervals
  point_buffer = 100000
  # Offset (must be either "oldest" or "newest")
  offset = "oldest"
`

func (k *Kafka) SampleConfig() string {
	return sampleConfig
}

func (k *Kafka) Description() string {
	return "Read line-protocol metrics from Kafka topic(s)"
}

func (k *Kafka) Start() error {
	k.Lock()
	defer k.Unlock()
	var consumerErr error

	config := consumergroup.NewConfig()
	switch strings.ToLower(k.Offset) {
	case "oldest", "":
		config.Offsets.Initial = sarama.OffsetOldest
	case "newest":
		config.Offsets.Initial = sarama.OffsetNewest
	default:
		log.Printf("WARNING: Kafka consumer invalid offset '%s', using 'oldest'\n",
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
	if k.PointBuffer == 0 {
		k.PointBuffer = 100000
	}
	k.pointChan = make(chan models.Point, k.PointBuffer)

	// Start the kafka message reader
	go k.parser()
	log.Printf("Started the kafka consumer service, peers: %v, topics: %v\n",
		k.ZookeeperPeers, k.Topics)
	return nil
}

// parser() reads all incoming messages from the consumer, and parses them into
// influxdb metric points.
func (k *Kafka) parser() {
	for {
		select {
		case <-k.done:
			return
		case err := <-k.errs:
			log.Printf("Kafka Consumer Error: %s\n", err.Error())
		case msg := <-k.in:
			points, err := models.ParsePoints(msg.Value)
			if err != nil {
				log.Printf("Could not parse kafka message: %s, error: %s",
					string(msg.Value), err.Error())
			}

			for _, point := range points {
				select {
				case k.pointChan <- point:
					continue
				default:
					log.Printf("Kafka Consumer buffer is full, dropping a point." +
						" You may want to increase the point_buffer setting")
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
		log.Printf("Error closing kafka consumer: %s\n", err.Error())
	}
}

func (k *Kafka) Gather(acc inputs.Accumulator) error {
	k.Lock()
	defer k.Unlock()
	npoints := len(k.pointChan)
	for i := 0; i < npoints; i++ {
		point := <-k.pointChan
		acc.AddFields(point.Name(), point.Fields(), point.Tags(), point.Time())
	}
	return nil
}

func init() {
	inputs.Add("kafka_consumer", func() inputs.Input {
		return &Kafka{}
	})
}
