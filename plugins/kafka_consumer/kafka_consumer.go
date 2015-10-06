package kafka_consumer

import (
	"os"
	"os/signal"
	"time"

	"github.com/Shopify/sarama"
	"github.com/influxdb/influxdb/models"
	"github.com/influxdb/telegraf/plugins"
	"github.com/wvanbergen/kafka/consumergroup"
)

type Kafka struct {
	ConsumerGroupName string
	Topic             string
	ZookeeperPeers    []string
	Consumer          *consumergroup.ConsumerGroup
	BatchSize         int
}

var sampleConfig = `
	# topic to consume
	topic = "topic_with_metrics"

	# the name of the consumer group
	consumerGroupName = "telegraf_metrics_consumers"

	# an array of Zookeeper connection strings
	zookeeperPeers = ["localhost:2181"]

	# Batch size of points sent to InfluxDB
	batchSize = 1000
`

func (k *Kafka) SampleConfig() string {
	return sampleConfig
}

func (k *Kafka) Description() string {
	return "read metrics from a Kafka topic"
}

type Metric struct {
	Measurement string                 `json:"measurement"`
	Values      map[string]interface{} `json:"values"`
	Tags        map[string]string      `json:"tags"`
	Time        time.Time              `json:"time"`
}

func (k *Kafka) Gather(acc plugins.Accumulator) error {
	var consumerErr error
	metricQueue := make(chan []byte, 200)

	if k.Consumer == nil {
		k.Consumer, consumerErr = consumergroup.JoinConsumerGroup(
			k.ConsumerGroupName,
			[]string{k.Topic},
			k.ZookeeperPeers,
			nil,
		)

		if consumerErr != nil {
			return consumerErr
		}

		c := make(chan os.Signal, 1)
		halt := make(chan bool, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			<-c
			halt <- true
			emitMetrics(k, acc, metricQueue)
			k.Consumer.Close()
		}()

		go readFromKafka(k.Consumer.Messages(), metricQueue, k.BatchSize, k.Consumer.CommitUpto, halt)
	}

	return emitMetrics(k, acc, metricQueue)
}

func emitMetrics(k *Kafka, acc plugins.Accumulator, metricConsumer <-chan []byte) error {
	timeout := time.After(1 * time.Second)

	for {
		select {
		case batch := <-metricConsumer:
			var points []models.Point
			var err error
			if points, err = models.ParsePoints(batch); err != nil {
				return err
			}

			for _, point := range points {
				acc.AddFieldsWithTime(point.Name(), point.Fields(), point.Tags(), point.Time())
			}
		case <-timeout:
			return nil
		}
	}
}

const millisecond = 1000000 * time.Nanosecond

type ack func(*sarama.ConsumerMessage) error

func readFromKafka(kafkaMsgs <-chan *sarama.ConsumerMessage, metricProducer chan<- []byte, maxBatchSize int, ackMsg ack, halt <-chan bool) {
	batch := make([]byte, 0)
	currentBatchSize := 0
	timeout := time.After(500 * millisecond)
	var msg *sarama.ConsumerMessage

	for {
		select {
		case msg = <-kafkaMsgs:
			if currentBatchSize != 0 {
				batch = append(batch, '\n')
			}

			batch = append(batch, msg.Value...)
			currentBatchSize++

			if currentBatchSize == maxBatchSize {
				metricProducer <- batch
				currentBatchSize = 0
				batch = make([]byte, 0)
				ackMsg(msg)
			}
		case <-timeout:
			if currentBatchSize != 0 {
				metricProducer <- batch
				currentBatchSize = 0
				batch = make([]byte, 0)
				ackMsg(msg)
			}

			timeout = time.After(500 * millisecond)
		case <-halt:
			if currentBatchSize != 0 {
				metricProducer <- batch
				ackMsg(msg)
			}

			return
		}
	}
}

func init() {
	plugins.Add("kafka", func() plugins.Plugin {
		return &Kafka{}
	})
}
