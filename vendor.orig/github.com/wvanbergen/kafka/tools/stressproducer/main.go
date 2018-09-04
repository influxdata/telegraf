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

var (
	brokerList      = flag.String("brokers", "localhost:9092", "The comma separated list of brokers in the Kafka cluster")
	topic           = flag.String("topic", "test.throughput", "The topic to produce messages to")
	messageBodySize = flag.Int("message-body-size", 100, "The size of the message payload")
	waitForAll      = flag.Bool("wait-for-all", false, "Whether to wait for all ISR to Ack the message")
	sleep           = flag.Int("sleep", 1000, "The number of nanoseconds to sleep between messages")
	verbose         = flag.Bool("verbose", false, "Whether to enable Sarama logging")
	statFrequency   = flag.Int("statFrequency", 1000, "How frequently (in messages) to print throughput and latency")
)

type MessageMetadata struct {
	EnqueuedAt time.Time
}

func (mm *MessageMetadata) Latency() time.Duration {
	return time.Since(mm.EnqueuedAt)
}

func producerConfiguration() *sarama.Config {
	config := sarama.NewConfig()
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true

	if *waitForAll {
		config.Producer.RequiredAcks = sarama.WaitForAll
	} else {
		config.Producer.RequiredAcks = sarama.WaitForLocal
	}

	return config
}

func main() {
	flag.Parse()

	var (
		wg                            sync.WaitGroup
		enqueued, successes, failures int
		totalLatency                  time.Duration
	)

	if *verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	producer, err := sarama.NewAsyncProducer(strings.Split(*brokerList, ","), producerConfiguration())
	if err != nil {
		log.Fatalln("Failed to start producer:", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		var (
			latency, batchDuration time.Duration
			batchStartedAt         time.Time
			rate                   float64
		)

		batchStartedAt = time.Now()
		for message := range producer.Successes() {
			totalLatency += message.Metadata.(*MessageMetadata).Latency()
			successes++

			if successes%*statFrequency == 0 {

				batchDuration = time.Since(batchStartedAt)
				rate = float64(*statFrequency) / (float64(batchDuration) / float64(time.Second))
				latency = totalLatency / time.Duration(*statFrequency)

				log.Printf("Rate: %0.2f/s; latency: %0.2fms\n", rate, float64(latency)/float64(time.Millisecond))

				totalLatency = 0
				batchStartedAt = time.Now()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range producer.Errors() {
			log.Println("FAILURE:", err)
			failures++
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGTERM)

	messageBody := sarama.ByteEncoder(make([]byte, *messageBodySize))
ProducerLoop:
	for {
		message := &sarama.ProducerMessage{
			Topic:    *topic,
			Key:      sarama.StringEncoder(fmt.Sprintf("%d", enqueued)),
			Value:    messageBody,
			Metadata: &MessageMetadata{EnqueuedAt: time.Now()},
		}

		select {
		case <-signals:
			producer.AsyncClose()
			break ProducerLoop
		case producer.Input() <- message:
			enqueued++
		}

		if *sleep > 0 {
			time.Sleep(time.Duration(*sleep))
		}
	}

	fmt.Println("Waiting for in flight messages to be processed...")
	wg.Wait()

	log.Println()
	log.Printf("Enqueued: %d; Produced: %d; Failed: %d.\n", enqueued, successes, failures)
}
