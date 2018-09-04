package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/bsm/sarama-cluster"
)

var (
	groupID    = flag.String("group", "", "REQUIRED: The shared consumer group name")
	brokerList = flag.String("brokers", os.Getenv("KAFKA_PEERS"), "The comma separated list of brokers in the Kafka cluster")
	topicList  = flag.String("topics", "", "REQUIRED: The comma separated list of topics to consume")
	offset     = flag.String("offset", "newest", "The offset to start with. Can be `oldest`, `newest`")
	verbose    = flag.Bool("verbose", false, "Whether to turn on sarama logging")

	logger = log.New(os.Stderr, "", log.LstdFlags)
)

func main() {
	flag.Parse()

	if *groupID == "" {
		printUsageErrorAndExit("You have to provide a -group name.")
	} else if *brokerList == "" {
		printUsageErrorAndExit("You have to provide -brokers as a comma-separated list, or set the KAFKA_PEERS environment variable.")
	} else if *topicList == "" {
		printUsageErrorAndExit("You have to provide -topics as a comma-separated list.")
	}

	// Init config
	config := cluster.NewConfig()
	if *verbose {
		sarama.Logger = logger
	} else {
		config.Consumer.Return.Errors = true
		config.Group.Return.Notifications = true
	}

	switch *offset {
	case "oldest":
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	case "newest":
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	default:
		printUsageErrorAndExit("-offset should be `oldest` or `newest`")
	}

	// Init consumer, consume errors & messages
	consumer, err := cluster.NewConsumer(strings.Split(*brokerList, ","), *groupID, strings.Split(*topicList, ","), config)
	if err != nil {
		printErrorAndExit(69, "Failed to start consumer: %s", err)
	}
	defer consumer.Close()

	// Create signal channel
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	// Consume all channels, wait for signal to exit
	for {
		select {
		case msg, more := <-consumer.Messages():
			if more {
				fmt.Fprintf(os.Stdout, "%s/%d/%d\t%s\n", msg.Topic, msg.Partition, msg.Offset, msg.Value)
				consumer.MarkOffset(msg, "")
			}
		case ntf, more := <-consumer.Notifications():
			if more {
				logger.Printf("Rebalanced: %+v\n", ntf)
			}
		case err, more := <-consumer.Errors():
			if more {
				logger.Printf("Error: %s\n", err.Error())
			}
		case <-sigchan:
			return
		}
	}
}

func printErrorAndExit(code int, format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n\n", values...)
	os.Exit(code)
}

func printUsageErrorAndExit(format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n\n", values...)
	flag.Usage()
	os.Exit(64)
}
