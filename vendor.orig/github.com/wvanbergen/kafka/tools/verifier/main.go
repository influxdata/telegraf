package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
)

const (
	ExpectationBufferSize = 1000
)

var (
	brokerList = flag.String("brokers", "localhost:9092", "The comma separated list of brokers in the Kafka cluster")
	topic      = flag.String("topic", "", "The topic to consume")
	sleep      = flag.Int("sleep", 1000, "The number of nanoseconds to sleep between producing messages")
	batchSize  = flag.Int("batch-size", 1000000, "The number of messages to produce")

	verbose = flag.Bool("verbose", false, "Whether to turn on sarama logging")

	logger   = log.New(os.Stderr, "", log.LstdFlags)
	stats    = &Stats{}
	shutdown = make(chan os.Signal, 1)
)

func monitor() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for _ = range ticker.C {
		stats.Print()
	}
}

func main() {
	flag.Parse()

	if *verbose {
		sarama.Logger = logger
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Consumer.Return.Errors = true

	client, err := sarama.NewClient(strings.Split(*brokerList, ","), config)
	if err != nil {
		logger.Fatalln("Failed to start Kafka client:", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			logger.Println("Failed to close client:", err)
		}
	}()

	producer, err := sarama.NewAsyncProducerFromClient(client)
	if err != nil {
		logger.Fatalln("Failed to start Kafka producer:", err)
	}

	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		logger.Fatalln("Failed to start Kafka consumer:", err)
	}

	signal.Notify(shutdown, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGTERM)
	expectations := make(chan *sarama.ProducerMessage, ExpectationBufferSize)

	started := time.Now()

	var verifierWg sync.WaitGroup
	verifierWg.Add(2)
	go expectationProducer(producer, expectations, &verifierWg)
	go expectationConsumer(consumer, expectations, &verifierWg)
	verifierWg.Wait()

	stats.Print()

	logger.Println()
	logger.Printf("Done after %0.2fs.\n", float64(time.Since(started))/float64(time.Second))
}

func expectationProducer(p sarama.AsyncProducer, expectations chan<- *sarama.ProducerMessage, wg *sync.WaitGroup) {
	defer wg.Done()

	var producerWg sync.WaitGroup

	producerWg.Add(1)
	go func() {
		defer producerWg.Done()
		for msg := range p.Successes() {
			stats.LogProduced(msg)
			expectations <- msg
		}
	}()

	producerWg.Add(1)
	go func() {
		defer producerWg.Done()
		for err := range p.Errors() {
			logger.Println("Failed to produce message:", err)
		}
	}()

	go monitor()
	logger.Printf("Producing %d messages...\n", *batchSize)

ProducerLoop:
	for i := 0; i < *batchSize; i++ {
		msg := &sarama.ProducerMessage{
			Topic:    *topic,
			Key:      sarama.StringEncoder(fmt.Sprintf("%d", i)),
			Value:    nil,
			Metadata: &MessageMetadata{Enqueued: time.Now()},
		}

		select {
		case <-shutdown:
			logger.Println("Early shutdown initiated...")
			break ProducerLoop
		case p.Input() <- msg:
			stats.LogEnqueued(msg)
		}

		if *sleep > 0 {
			time.Sleep(time.Duration(*sleep))
		}
	}

	p.AsyncClose()
	producerWg.Wait()
	close(expectations)
}

type partitionVerifier struct {
	pc           sarama.PartitionConsumer
	expectations chan *sarama.ProducerMessage
}

func expectationConsumer(c sarama.Consumer, expectations <-chan *sarama.ProducerMessage, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if err := c.Close(); err != nil {
			logger.Println("Failed to close consumer:", err)
		}
	}()

	var (
		partitionVerifiers = make(map[int32]*partitionVerifier)
		consumerWg         sync.WaitGroup
	)

	for expectation := range expectations {
		partition := expectation.Partition

		if partitionVerifiers[partition] == nil {
			logger.Printf("Starting message verifier for partition %d...\n", partition)
			pc, err := c.ConsumePartition(*topic, partition, expectation.Offset)
			if err != nil {
				logger.Fatalf("Failed to open partition consumer for %s/%d: %s", *topic, expectation.Partition, err)
			}

			partitionExpectations := make(chan *sarama.ProducerMessage)
			partitionVerifiers[partition] = &partitionVerifier{pc: pc, expectations: partitionExpectations}

			consumerWg.Add(1)
			go partitionExpectationConsumer(pc, partitionExpectations, &consumerWg)
		}

		partitionVerifiers[partition].expectations <- expectation
	}

	for _, pv := range partitionVerifiers {
		close(pv.expectations)
	}

	consumerWg.Wait()
}

func partitionExpectationConsumer(pc sarama.PartitionConsumer, expectations <-chan *sarama.ProducerMessage, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if err := pc.Close(); err != nil {
			logger.Println("Failed to close partitionconsumer:", err)
		}
	}()

	for expectation := range expectations {
		msg := <-pc.Messages()

		if msg.Offset != expectation.Offset {
			fmt.Printf("Unexpected offset %d!\n", msg.Offset)
		}

		key, _ := expectation.Key.Encode()
		if string(key) != string(msg.Key) {
			fmt.Printf("Unexpected key: %v!\n", msg.Key)
		}
		if msg.Value != nil {
			fmt.Printf("Unexpected value: %v!\n", msg.Value)
		}

		stats.LogConsumed(expectation)
	}
}
