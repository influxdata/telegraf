package kafka

import (
	"errors"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/influxdb/influxdb/client/v2"
	"github.com/influxdb/telegraf/plugins/outputs"
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

func (k *Kafka) Write(points []*client.Point) error {
	if len(points) == 0 {
		return nil
	}

	for _, p := range points {
		// Combine tags from Point and BatchPoints and grab the resulting
		// line-protocol output string to write to Kafka
		value := p.String()

		m := &sarama.ProducerMessage{
			Topic: k.Topic,
			Value: sarama.StringEncoder(value),
		}
		if h, ok := p.Tags()[k.RoutingTag]; ok {
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
