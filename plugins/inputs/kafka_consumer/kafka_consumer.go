package kafka_consumer

import (
	"crypto/tls"
	"log"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/wvanbergen/kafka/consumergroup"
)

type Kafka struct {
	// new kafka consumer
	NewConsumer bool
	// common for both versions
	ConsumerGroup string
	Topics        []string
	Offset        string

	// for 0.8
	ZookeeperPeers  []string
	ZookeeperChroot string
	Consumer        *consumergroup.ConsumerGroup

	// for 0.9+
	BrokerList []string
	Consumer9  *cluster.Consumer
	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`

	// Skip SSL verification
	InsecureSkipVerify bool

	tlsConfig tls.Config

	// Legacy metric buffer support
	MetricBuffer int
	// TODO remove PointBuffer, legacy support
	PointBuffer int

	parser parsers.Parser

	sync.Mutex

	// channel for all incoming kafka messages
	in <-chan *sarama.ConsumerMessage

	// channel for all kafka consumer errors
	errs  <-chan *sarama.ConsumerError
	errs9 <-chan error

	done chan struct{}

	// keep the accumulator internally:
	acc telegraf.Accumulator

	// doNotCommitMsgs tells the parser not to call CommitUpTo on the consumer
	// this is mostly for test purposes, but there may be a use-case for it later.
	doNotCommitMsgs bool
}

var sampleConfig = `
  ## is new consumer?
  new_consumer = true
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
  ## Each data format has it's own unique set of configuration options, read
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
	k.Lock()
	defer k.Unlock()
	var consumerErr error

	k.acc = acc

	log.Println(k.NewConsumer)

	if !k.NewConsumer {
		config := consumergroup.NewConfig()

		config.Zookeeper.Chroot = k.ZookeeperChroot
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
		// Start the kafka message reader
		go k.receiver()
		log.Printf("Started the kafka consumer service, peers: %v, topics: %v\n",
			k.ZookeeperPeers, k.Topics)
	} else {
		config := cluster.NewConfig()

		tlsConfig, err := internal.GetTLSConfig(k.SSLCert, k.SSLKey, k.SSLCA, k.InsecureSkipVerify)
		if err != nil {
			return err
		}

		if tlsConfig != nil {
			config.Net.TLS.Config = tlsConfig
			config.Net.TLS.Enable = true
		}

		switch strings.ToLower(k.Offset) {
		case "oldest", "":
			config.Consumer.Offsets.Initial = sarama.OffsetOldest
		case "newest":
			config.Consumer.Offsets.Initial = sarama.OffsetNewest
		default:
			log.Printf("WARNING: Kafka consumer invalid offset '%s', using 'oldest'\n",
				k.Offset)
			config.Consumer.Offsets.Initial = sarama.OffsetOldest
		}

		// TODO: make this configurable
		config.Consumer.Return.Errors = true

		if err := config.Validate(); err != nil {
			return err
		}

		k.Consumer9, err = cluster.NewConsumer(k.BrokerList, k.ConsumerGroup, k.Topics, config)
		if err != nil {
			return err
		}
		// Setup message and error channels
		k.in = k.Consumer9.Messages()
		k.errs9 = k.Consumer9.Errors()
		k.done = make(chan struct{})
		// Start the kafka message reader for 0.9
		go k.collector()
		log.Printf("Started the kafka consumer service with new consumer, brokers: %v, topics: %v\n",
			k.BrokerList, k.Topics)
	}

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
			log.Printf("Kafka Consumer Error: %s\n", err.Error())
		case msg := <-k.in:
			metrics, err := k.parser.Parse(msg.Value)
			if err != nil {
				log.Printf("KAFKA PARSE ERROR\nmessage: %s\nerror: %s",
					string(msg.Value), err.Error())
			}

			for _, metric := range metrics {
				k.acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
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

// this is for kafka new consumer
func (k *Kafka) collector() {
	for {
		select {
		case <-k.done:
			return
		case err := <-k.errs9:
			log.Printf("Kafka Consumer Error: %s\n", err.Error())
		case msg := <-k.in:
			metrics, err := k.parser.Parse(msg.Value)

			if err != nil {
				log.Printf("KAFKA PARSE ERROR\nmessage: %s\nerror: %s",
					string(msg.Value), err.Error())
			}

			for _, metric := range metrics {
				k.acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
			}

			if !k.doNotCommitMsgs {
				k.Consumer9.MarkOffset(msg, "")
			}
		}
	}
}

func (k *Kafka) Stop() {
	k.Lock()
	defer k.Unlock()
	close(k.done)
	if !k.NewConsumer {
		if err := k.Consumer.Close(); err != nil {
			log.Printf("Error closing kafka consumer: %s\n", err.Error())
		}
	} else {
		if err := k.Consumer9.Close(); err != nil {
			log.Printf("Error closing kafka consumer: %s\n", err.Error())
		}
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
