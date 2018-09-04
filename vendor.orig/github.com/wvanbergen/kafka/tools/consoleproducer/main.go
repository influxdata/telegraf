package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Shopify/sarama"
)

var (
	brokerList  = flag.String("brokers", "localhost:9092", "The comma separated list of brokers in the Kafka cluster")
	topic       = flag.String("topic", "", "The topic to produce to")
	key         = flag.String("key", "", "The key of the message to produce")
	value       = flag.String("value", "", "The value of the message to produce")
	partitioner = flag.String("partitioner", "hash", "The partitioning scheme to use. Can be `hash`, or `random`")
	verbose     = flag.Bool("verbose", false, "Whether to turn on sarama logging")

	logger = log.New(os.Stderr, "", log.LstdFlags)
)

func main() {
	flag.Parse()

	if *verbose {
		sarama.Logger = logger
	}

	var partitionerConstructor sarama.PartitionerConstructor
	switch *partitioner {
	case "hash":
		partitionerConstructor = sarama.NewHashPartitioner
	case "random":
		partitionerConstructor = sarama.NewRandomPartitioner
	default:
		log.Fatalf("Partitioner %s not supported.", *partitioner)
	}

	var keyEncoder, valueEncoder sarama.Encoder
	if *key != "" {
		keyEncoder = sarama.StringEncoder(*key)
	}
	if *value != "" {
		valueEncoder = sarama.StringEncoder(*value)
	}

	config := sarama.NewConfig()
	config.Producer.Partitioner = partitionerConstructor

	producer, err := sarama.NewSyncProducer(strings.Split(*brokerList, ","), config)
	if err != nil {
		logger.Fatalln("FAILED to open the producer:", err)
	}
	defer producer.Close()

	partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: *topic,
		Key:   keyEncoder,
		Value: valueEncoder,
	})

	if err != nil {
		logger.Println("FAILED to produce message:", err)
	} else {
		fmt.Printf("topic=%s\tpartition=%d\toffset=%d\n", *topic, partition, offset)
	}
}
