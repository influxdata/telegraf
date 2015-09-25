package kafka

import (
	"errors"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/influxdb/influxdb/client"
	"github.com/influxdb/telegraf/outputs"
)

type Kafka struct {
	// Kafka brokers to send metrics to
	Brokers []string
	// Kafka topic
	Topic string
	// Routing Key Tag
	RoutingTag string `toml:"routing_tag"`

	producer sarama.SyncProducer
}

var sampleConfig = `
	# URLs of kafka brokers
	brokers = ["localhost:9092"]
	# Kafka topic for producer messages
	topic = "telegraf"
	# Telegraf tag to use as a routing key
	#  ie, if this tag exists, it's value will be used as the routing key
	routing_tag = "host"
`

func (k *Kafka) Connect() error {
	producer, err := sarama.NewSyncProducer(k.Brokers, nil)
	if err != nil {
		return err
	}
	k.producer = producer
	return nil
}

func (k *Kafka) Close() error {
	return k.producer.Close()
}

func (k *Kafka) SampleConfig() string {
	return sampleConfig
}

func (k *Kafka) Description() string {
	return "Configuration for the Kafka server to send metrics to"
}

func (k *Kafka) Write(bp client.BatchPoints) error {
	if len(bp.Points) == 0 {
		return nil
	}

	var zero_time time.Time
	for _, p := range bp.Points {
		// Combine tags from Point and BatchPoints and grab the resulting
		// line-protocol output string to write to Kafka
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
			if p.Time == zero_time {
				if bp.Time == zero_time {
					p.Time = time.Now()
				} else {
					p.Time = bp.Time
				}
			}
			value = p.MarshalString()
		}

		m := &sarama.ProducerMessage{
			Topic: k.Topic,
			Value: sarama.StringEncoder(value),
		}
		if h, ok := p.Tags[k.RoutingTag]; ok {
			m.Key = sarama.StringEncoder(h)
		}

		_, _, err := k.producer.SendMessage(m)
		if err != nil {
			return errors.New(fmt.Sprintf("FAILED to send kafka message: %s\n",
				err))
		}
	}
	return nil
}

func init() {
	outputs.Add("kafka", func() outputs.Output {
		return &Kafka{}
	})
}
