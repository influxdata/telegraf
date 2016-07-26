package kafka_consumer_tls

import (
	"crypto/tls"
	"log"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type Kafka struct {
	Consumer *cluster.Consumer

	// if read message from beginning
	FromBeginning bool

	BrokerList []string

	ConsumerGroup string

	Topics []string

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`

	// Skip SSL verification
	InsecureSkipVerify bool

	tlsConfig tls.Config

	parser parsers.Parser

	sync.Mutex

	// channel for all incoming kafka messages
	in <-chan *sarama.ConsumerMessage
	// channel for all kafka consumer errors
	errs <-chan error
	// dummy channel to tell the end
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
  ## an array of kafka 0.9+ brokers
  broker_list = ["localhost:2181"]
  ## the name of the consumer group
  consumer_group = "telegraf_metrics_consumers"
  ## from beginning
  from_beginning = true

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

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

	k.acc = acc

	config := cluster.NewConfig()

	tlsConfig, err := internal.GetTLSConfig(k.SSLCert, k.SSLKey, k.SSLCA, k.InsecureSkipVerify)
	if err != nil {
		return err
	}

	if tlsConfig != nil {
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Enable = true
	}

	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	if !k.FromBeginning {
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	}

	// TODO: make this configurable
	config.Consumer.Return.Errors = true

	if err := config.Validate(); err != nil {
		return err
	}

	k.Consumer, err = cluster.NewConsumer(k.BrokerList, k.ConsumerGroup, k.Topics, config)
	if err != nil {
		return err
	}
	// Setup message and error channels
	k.in = k.Consumer.Messages()
	k.errs = k.Consumer.Errors()

	k.done = make(chan struct{})

	go k.collector()

	return nil
}

func (k *Kafka) collector() {
	for {
		select {
		case <-k.done:
			return
		case err := <-k.errs:
			log.Printf("Kafka Consumer Error: %s\n", err.Error())
		case msg := <-k.in:
			metrics, err := k.parser.Parse(msg.Value)
			if err != nil {
				log.Printf("KAFKA PARSE ERROR\nmessage: %s\nerror: %s", string(msg.Value), err.Error())
			}

			for _, metric := range metrics {
				k.acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
			}

			if !k.doNotCommitMsgs {
				k.Consumer.MarkOffset(msg, "")
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

func (k *Kafka) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("kafka_consumer_tls", func() telegraf.Input {
		return &Kafka{}
	})
}
