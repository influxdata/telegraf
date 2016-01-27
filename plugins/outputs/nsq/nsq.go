package nsq

import (
	"fmt"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/nsqio/go-nsq"
)

type NSQ struct {
	Server   string
	Topic    string
	producer *nsq.Producer
}

var sampleConfig = `
  # Location of nsqd instance listening on TCP
  server = "localhost:4150"
  # NSQ topic for producer messages
  topic = "telegraf"
`

func (n *NSQ) Connect() error {
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer(n.Server, config)

	if err != nil {
		return err
	}

	n.producer = producer
	return nil
}

func (n *NSQ) Close() error {
	n.producer.Stop()
	return nil
}

func (n *NSQ) SampleConfig() string {
	return sampleConfig
}

func (n *NSQ) Description() string {
	return "Send telegraf measurements to NSQD"
}

func (n *NSQ) Write(points []*client.Point) error {
	if len(points) == 0 {
		return nil
	}

	for _, p := range points {
		// Combine tags from Point and BatchPoints and grab the resulting
		// line-protocol output string to write to NSQ
		value := p.String()

		err := n.producer.Publish(n.Topic, []byte(value))

		if err != nil {
			return fmt.Errorf("FAILED to send NSQD message: %s", err)
		}
	}
	return nil
}

func init() {
	outputs.Add("nsq", func() outputs.Output {
		return &NSQ{}
	})
}
