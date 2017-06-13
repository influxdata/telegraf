package nsq_consumer

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/nsqio/go-nsq"
)

//NSQConsumer represents the configuration of the plugin
type NSQConsumer struct {
	Server      string
	Topic       string
	Channel     string
	MaxInFlight int
	parser      parsers.Parser
	consumer    *nsq.Consumer
	acc         telegraf.Accumulator
}

var sampleConfig = `
  ## An string representing the NSQD TCP Endpoint
  server = "localhost:4150"
  topic = "telegraf"
  channel = "consumer"
  max_in_flight = 100

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

func init() {
	inputs.Add("nsq_consumer", func() telegraf.Input {
		return &NSQConsumer{}
	})
}

// SetParser takes the data_format from the config and finds the right parser for that format
func (n *NSQConsumer) SetParser(parser parsers.Parser) {
	n.parser = parser
}

// SampleConfig returns config values for generating a sample configuration file
func (n *NSQConsumer) SampleConfig() string {
	return sampleConfig
}

// Description prints description string
func (n *NSQConsumer) Description() string {
	return "Read NSQ topic for metrics."
}

// Start pulls data from nsq
func (n *NSQConsumer) Start(acc telegraf.Accumulator) error {
	n.acc = acc
	n.connect()
	n.consumer.AddConcurrentHandlers(nsq.HandlerFunc(func(message *nsq.Message) error {
		metrics, err := n.parser.Parse(message.Body)
		if err != nil {
			acc.AddError(fmt.Errorf("E! NSQConsumer Parse Error\nmessage:%s\nerror:%s", string(message.Body), err.Error()))
			return nil
		}
		for _, metric := range metrics {
			n.acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
		}
		message.Finish()
		return nil
	}), n.MaxInFlight)
	n.consumer.ConnectToNSQD(n.Server)
	return nil
}

// Stop processing messages
func (n *NSQConsumer) Stop() {
	n.consumer.Stop()
}

// Gather is a noop
func (n *NSQConsumer) Gather(acc telegraf.Accumulator) error {
	return nil
}

func (n *NSQConsumer) connect() error {
	if n.consumer == nil {
		config := nsq.NewConfig()
		config.MaxInFlight = n.MaxInFlight
		consumer, err := nsq.NewConsumer(n.Topic, n.Channel, config)
		if err != nil {
			return err
		}
		n.consumer = consumer
	}
	return nil
}
