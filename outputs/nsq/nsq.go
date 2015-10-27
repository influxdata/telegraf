package nsq

import (
	"fmt"
	"github.com/influxdb/influxdb/client"
	"github.com/influxdb/telegraf/outputs"
	"github.com/nsqio/go-nsq"
	"time"
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

func (n *NSQ) Write(bp client.BatchPoints) error {
	if len(bp.Points) == 0 {
		return nil
	}

	var zeroTime time.Time
	for _, p := range bp.Points {
		// Combine tags from Point and BatchPoints and grab the resulting
		// line-protocol output string to write to NSQ
		var value string
		if p.Raw != "" {
			value = p.Raw
		} else {
			for k, v := range bp.Tags {
				if p.Tags == nil {
					p.Tags = make(map[string]string, len(bp.Tags))
				}
				p.Tags[k] = v
			}
			if p.Time == zeroTime {
				if bp.Time == zeroTime {
					p.Time = time.Now()
				} else {
					p.Time = bp.Time
				}
			}
			value = p.MarshalString()
		}

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
