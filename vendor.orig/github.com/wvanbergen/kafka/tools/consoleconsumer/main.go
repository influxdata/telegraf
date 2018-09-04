package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/Shopify/sarama"
)

var (
	brokerList = flag.String("brokers", "localhost:9092", "The comma separated list of brokers in the Kafka cluster")
	topic      = flag.String("topic", "", "The topic to consume")
	partition  = flag.Int("partition", 0, "The partition to consume")
	offset     = flag.String("offset", "oldest", "The offset to start with. Can be `oldest`, `newest`, or an actual offset")
	verbose    = flag.Bool("verbose", false, "Whether to turn on sarama logging")

	logger = log.New(os.Stderr, "", log.LstdFlags)
)

func main() {
	flag.Parse()

	if *verbose {
		sarama.Logger = logger
	}

	var (
		initialOffset int64
		offsetError   error
	)
	switch *offset {
	case "oldest":
		initialOffset = sarama.OffsetOldest
	case "newest":
		initialOffset = sarama.OffsetNewest
	default:
		initialOffset, offsetError = strconv.ParseInt(*offset, 10, 64)
	}

	if offsetError != nil {
		logger.Fatalln("Invalid initial offset:", *offset)
	}

	c, err := sarama.NewConsumer(strings.Split(*brokerList, ","), nil)
	if err != nil {
		logger.Fatalln(err)
	}

	pc, err := c.ConsumePartition(*topic, int32(*partition), initialOffset)
	if err != nil {
		logger.Fatalln(err)
	}

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Kill, os.Interrupt)
		<-signals
		pc.AsyncClose()
	}()

	for msg := range pc.Messages() {
		fmt.Printf("Offset: %d\n", msg.Offset)
		fmt.Printf("Key:    %s\n", string(msg.Key))
		fmt.Printf("Value:  %s\n", string(msg.Value))
		fmt.Println()
	}

	if err := c.Close(); err != nil {
		fmt.Println("Failed to close consumer: ", err)
	}
}
