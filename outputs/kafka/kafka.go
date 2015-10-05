package kafka

import (
	"errors"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/influxdb/influxdb/client"
	"github.com/koksan83/telegraf/outputs"
)

type Kafka struct {
	// Kafka brokers to send metrics to
	Brokers []string
	// Kafka topic
	Topic string

	producer sarama.SyncProducer
}

var sampleConfig = `
    # URLs of kafka brokers
    brokers = ["localhost:9092"]
    # Kafka topic for producer messages
    topic = "telegraf"
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
			value = p.MarshalString()
		}

		m := &sarama.ProducerMessage{
			Topic: k.Topic,
			Value: sarama.StringEncoder(value),
		}
		if h, ok := p.Tags["host"]; ok {
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
