package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/wvanbergen/kazoo-go"
)

var (
	zookeeper        = flag.String("zookeeper", os.Getenv("ZOOKEEPER_PEERS"), "Zookeeper connection string. It can include a chroot.")
	zookeeperTimeout = flag.Int("zookeeper-timeout", 1000, "Zookeeper timeout in milliseconds.")
)

func main() {
	flag.Parse()

	if *zookeeper == "" {
		printUsageErrorAndExit("You have to provide a zookeeper connection string using -zookeeper, or the ZOOKEEPER_PEERS environment variable")
	}

	conf := kazoo.NewConfig()
	conf.Timeout = time.Duration(*zookeeperTimeout) * time.Millisecond

	kz, err := kazoo.NewKazooFromConnectionString(*zookeeper, conf)
	if err != nil {
		printErrorAndExit(69, "Failed to connect to Zookeeper: %v", err)
	}
	defer func() { _ = kz.Close() }()

	topics, err := kz.Topics()
	if err != nil {
		printErrorAndExit(69, "Failed to get Kafka topics from Zookeeper: %v", err)
	}
	sort.Sort(topics)

	var (
		wg     sync.WaitGroup
		l      sync.Mutex
		stdout = make([]string, len(topics))
	)

	for i, topic := range topics {
		wg.Add(1)
		go func(i int, topic *kazoo.Topic) {
			defer wg.Done()

			buffer := bytes.NewBuffer(make([]byte, 0))

			partitions, err := topic.Partitions()
			if err != nil {
				printErrorAndExit(69, "Failed to get Kafka topic partitions from Zookeeper: %v", err)
			}

			fmt.Fprintf(buffer, "Topic: %s\tPartitions: %d\n", topic.Name, len(partitions))

			for _, partition := range partitions {
				leader, _ := partition.Leader()
				isr, _ := partition.ISR()

				fmt.Fprintf(buffer, "\tPartition: %d\tReplicas: %v\tLeader: %d\tISR: %v\n", partition.ID, partition.Replicas, leader, isr)
			}

			l.Lock()
			stdout[i] = buffer.String()
			l.Unlock()
		}(i, topic)
	}

	wg.Wait()
	for _, msg := range stdout {
		fmt.Print(msg)
	}
}

func printUsageErrorAndExit(format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", fmt.Sprintf(format, values...))
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Available command line options:")
	flag.PrintDefaults()
	os.Exit(64)
}

func printErrorAndExit(code int, format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", fmt.Sprintf(format, values...))
	fmt.Fprintln(os.Stderr)
	os.Exit(code)
}
