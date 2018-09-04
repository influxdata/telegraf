package main

import (
	"flag"
	"log"
	"os"

	"github.com/Shopify/sarama"
	"github.com/wvanbergen/kazoo-go"
)

var (
	zookeeper = flag.String("zookeeper", os.Getenv("ZOOKEEPER_PEERS"), "The zookeeper connection string")
	groupName = flag.String("group", "", "The consumer group to transfer offsets for")
)

func main() {
	flag.Parse()

	if *zookeeper == "" {
		log.Fatal("The -zookeeper command line argument is required")
	}

	if *groupName == "" {
		log.Fatal("The -group command line argument is required")
	}

	var (
		config         = kazoo.NewConfig()
		zookeeperNodes []string
	)

	zookeeperNodes, config.Chroot = kazoo.ParseConnectionString(*zookeeper)
	kz, err := kazoo.NewKazoo(zookeeperNodes, config)
	if err != nil {
		log.Fatal("[ERROR] Failed to connect to the zookeeper cluster:", err)
	}
	defer kz.Close()

	brokerList, err := kz.BrokerList()
	if err != nil {
		log.Fatal("[ERROR] Failed to retrieve Kafka broker list from zookeeper:", err)
	}

	group := kz.Consumergroup(*groupName)
	if exists, err := group.Exists(); err != nil {
		log.Fatal(err)
	} else if !exists {
		log.Fatalf("[ERROR] The consumergroup %s is not registered in Zookeeper", *groupName)
	}

	if instances, err := group.Instances(); err != nil {
		log.Fatal("[ERROR] Failed to get running instances from Zookeeper:", err)
	} else if len(instances) > 0 {
		log.Printf("[WARNING] This consumergroup has %d running instances. You should probably stop them before transferring offsets.", len(instances))
	}

	offsets, err := group.FetchAllOffsets()
	if err != nil {
		log.Fatal("[ERROR] Failed to retrieve offsets from zookeeper:", err)
	}

	client, err := sarama.NewClient(brokerList, nil)
	if err != nil {
		log.Fatal("[ERROR] Failed to connect to Kafka cluster:", err)
	}
	defer client.Close()

	coordinator, err := client.Coordinator(group.Name)
	if err != nil {
		log.Fatal("[ERROR] Failed to obtain coordinator for consumer group:", err)
	}

	request := &sarama.OffsetCommitRequest{
		Version:       1,
		ConsumerGroup: group.Name,
	}

	for topic, partitionOffsets := range offsets {
		for partition, nextOffset := range partitionOffsets {
			// In Zookeeper, we store the next offset to process.
			// In Kafka, we store the last offset that was processed.
			// So we have to fix an off by one error.
			lastOffset := nextOffset - 1
			request.AddBlock(topic, partition, lastOffset, 0, "")
		}
	}

	response, err := coordinator.CommitOffset(request)
	if err != nil {
		log.Fatal("[ERROR] Failed to commit offsets to Kafka:", err)
	}

	var errorsFound bool
	for topic, partitionOffsets := range offsets {
		for partition, nextOffset := range partitionOffsets {
			if _, ok := response.Errors[topic]; !ok {
				log.Printf("[WARNING] %s/%d: topic was not present in response and may not be committed!", topic, partition)
				continue
			}
			if _, ok := response.Errors[topic][partition]; !ok {
				log.Printf("[WARNING] %s/%d: partition was not present in response and may not be committed!", topic, partition)
				continue
			}

			if err := response.Errors[topic][partition]; err == sarama.ErrNoError {
				log.Printf("%s/%d: %d committed as last processed offset", topic, partition, nextOffset-1)
			} else {
				errorsFound = true
				log.Printf("[WARNING] %s/%d: offset %d was not committed: %s", topic, partition, nextOffset-1, err)
			}
		}
	}

	if !errorsFound {
		log.Print("[SUCCESS] Offsets successfully committed to Kafka!")
	}
}
