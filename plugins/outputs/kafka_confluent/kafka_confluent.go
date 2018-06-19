package kafka_confluent

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type (
	KafkaConfluent struct {
		// Kafka brokers to send metrics to
		Brokers string
		// Kafka topic
		Topic string

		producer   *kafka.Producer
		serializer serializers.Serializer
	}
)

var sampleConfig = `
# Send metrics to PostgreSQL using COPY
[[outputs.kafka_confluent]]
  ## URLs of kafka brokers
  brokers = "localhost:9092"
  ## Kafka topic for producer messages
  topic = "telegraf"
`

func (k *KafkaConfluent) SampleConfig() string {
	return sampleConfig
}

func (k *KafkaConfluent) Description() string {
	return "Configuration for the Kafka server to send metrics to"
}

func (k *KafkaConfluent) SetSerializer(serializer serializers.Serializer) {
	k.serializer = serializer
}

func (k *KafkaConfluent) Connect() error {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": k.Brokers})
	if err != nil {
		return err
	}

	k.producer = producer

	go func() {
		for e := range k.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition)
				}
			}
		}
	}()

	return nil
}

func (k *KafkaConfluent) Close() error {
	k.producer.Close()
	return nil
}

func (k *KafkaConfluent) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		buf, err := k.serializer.Serialize(metric)
		if err != nil {
			return err
		}
		err = k.producer.Produce(
			&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &k.Topic, Partition: kafka.PartitionAny},
				Value:          []byte(buf),
			}, nil)

		if err != nil {
			fmt.Errorf("FAILED to send kafka message: %s\n", err)
		}
	}
	k.producer.Flush(15 * 1000)

	return nil
}

func init() {
	outputs.Add("kafka_confluent", func() telegraf.Output {
		return &KafkaConfluent{}
	})
}
