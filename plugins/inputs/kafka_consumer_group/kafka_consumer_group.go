package kafka_consumer

import (
	"context"
	"fmt"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  ## Kafka brokers.
  brokers = ["localhost:9092"]

  ## Topics to consume.
  consumer_groups = ["telegraf"]

`

type KafkaConsumerGroup struct {
	Brokers        []string `toml:"brokers"`
	ConsumerGroups []string `toml:"consumer_groups"`

	config       *sarama.Config
	client       sarama.Client
	clusterAdmin sarama.ClusterAdmin
	wg           sync.WaitGroup
	cancel       context.CancelFunc
	offsetLookup map[string]int64
}

func (k *KafkaConsumerGroup) SampleConfig() string {
	return sampleConfig
}

func (k *KafkaConsumerGroup) Description() string {
	return "Configuration for Kafka consumer group"
}

func (k *KafkaConsumerGroup) iteratePartitions() error {
	var err error
	err = k.getCli()
	if err != nil {
		k.cliClose()
		return err
	}

	for _, consumerGroup := range k.ConsumerGroups {
		groupOffsets, err := k.clusterAdmin.ListConsumerGroupOffsets(consumerGroup, nil)
		if err != nil {
			k.cliClose()
			return err
		}

		offsetLookup := map[string]int64{}
		for topic, partitions := range groupOffsets.Blocks {
			for partition, _ := range partitions {
				latestOffset, err := k.client.GetOffset(topic, partition, sarama.OffsetNewest)
				if err != nil {
					k.cliClose()
					return err
				}
				key := k.getTopicPartitionKey(topic, partition)
				offsetLookup[key] = latestOffset
			}
		}
		k.offsetLookup = offsetLookup

	}

	return nil
}
func (k *KafkaConsumerGroup) getTopicPartitionKey(topic string, partition int32) string {

	return fmt.Sprint(topic, "\\", partition)
}

func (k *KafkaConsumerGroup) Stop() {
	k.cancel()
	k.wg.Wait()

	if k.clusterAdmin != nil {
		k.clusterAdmin.Close()
	}
	if k.client != nil {
		k.client.Close()
	}
}

func (k *KafkaConsumerGroup) Init() error {
	if len(k.Brokers) == 0 {
		return fmt.Errorf("broker list must not be empty")
	}
	if len(k.ConsumerGroups) < 1 {
		return fmt.Errorf("consumer groups must not be empty")
	}

	cfg := sarama.NewConfig()
	cfg.Version = sarama.V0_10_2_0

	k.config = cfg

	return nil
}

func (k *KafkaConsumerGroup) Gather(acc telegraf.Accumulator) error {

	var err error
	err = k.iteratePartitions()
	if err != nil {
		k.cliClose()
		return err
	}

	err = k.getCli()
	if err != nil {
		k.cliClose()
		return err
	}

	for _, consumerGroup := range k.ConsumerGroups {
		groupOffsets, err := k.clusterAdmin.ListConsumerGroupOffsets(consumerGroup, nil)
		if err != nil {
			k.cliClose()
			return err
		}

		for topic, partitions := range groupOffsets.Blocks {
			for partition, block := range partitions {
				key := k.getTopicPartitionKey(topic, partition)

				if _, ok := k.offsetLookup[key]; !ok {
					continue
				}
				latestOffset := k.offsetLookup[key]

				offsetLag := latestOffset - block.Offset
				if offsetLag < 0 {
					offsetLag = 0
				}

				tags := map[string]string{
					"consumerGroup": consumerGroup,
					"topic":         topic,
					"partition":     fmt.Sprintf("%d", partition),
				}
				fields := map[string]interface{}{
					"groupOffset":  block.Offset,
					"latestOffset": latestOffset,
					"lag":          offsetLag,
				}
				acc.AddFields("kafka_consumer_group_offsets", fields, tags)
			}
		}
	}

	return nil
}

func init() {
	inputs.Add("kafka_consumer_group", func() telegraf.Input {
		return &KafkaConsumerGroup{offsetLookup: map[string]int64{}}
	})
}

func (k *KafkaConsumerGroup) getCli() error {

	if k.client == nil {

		var err error
		k.client, err = sarama.NewClient(k.Brokers, k.config)
		if err != nil {
			return err
		}
		if k.client == nil {
			return fmt.Errorf("Kafka client not initialized")
		}

		k.clusterAdmin, err = sarama.NewClusterAdminFromClient(k.client)
		if err != nil {
			return err
		}
	}

	return nil
}

func (k *KafkaConsumerGroup) cliClose() {
	k.clusterAdmin.Close()
	k.client = nil
}
