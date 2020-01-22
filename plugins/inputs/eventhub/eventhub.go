package eventhub

// TODO: investigate why waitgroup inhibits exiting telegraf, is it even needed?
// TODO: (optional) Test authentication with AAD TokenProvider environment variables?
// TODO: (optional) Event Processor Host, only applicable for multiple Telegraf instances?

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"

	eventhub "github.com/Azure/azure-event-hubs-go/v3"
	"github.com/Azure/azure-event-hubs-go/v3/persist"
)

// EventHub is the top level struct for this plugin
type EventHub struct {
	// Configuration
	ConnectionString       string   `toml:"connection_string"`
	PersistenceDir         string   `toml:"persistence_dir"`
	ConsumerGroup          string   `toml:"consumer_group"`
	FromTimestamp          string   `toml:"from_timestamp"`
	StartingOffset         string   `toml:"starting_offset"`
	Latest                 bool     `toml:"latest"`
	PrefetchCount          uint32   `toml:"prefetch_count"`
	Epoch                  int64    `toml:"epoch"`
	UserAgent              string   `toml:"user_agent"`
	PartitionIDs           []string `toml:"partition_ids"`
	MaxUndeliveredMessages int      `toml:"max_undelivered_messages"`

	// Azure
	hub    *eventhub.Hub
	cancel context.CancelFunc

	// Influx
	parser parsers.Parser

	// Metrics tracking
	acc     telegraf.TrackingAccumulator
	tracker MessageTracker
	// wg      sync.WaitGroup
}

// MessageTracker is a struct with a lock and list of tracked messages
type MessageTracker struct {
	messages map[telegraf.TrackingID][]telegraf.Metric
	mux      sync.Mutex
}

// SampleConfig is provided here
func (*EventHub) SampleConfig() string {
	return `
  ## The default behavior is to create a new Event Hub client from environment variables.
  ## This requires one of the following sets of environment variables to be set:
  ##
  ## 1) Expected Environment Variables:
  ##    - "EVENTHUB_NAMESPACE"
  ##    - "EVENTHUB_NAME"
  ##    - "EVENTHUB_CONNECTION_STRING"
  ##
  ## 2) Expected Environment Variables:
  ##    - "EVENTHUB_NAMESPACE"
  ##    - "EVENTHUB_NAME"
  ##    - "EVENTHUB_KEY_NAME"
  ##    - "EVENTHUB_KEY_VALUE"

  ## Uncommenting the option below will create an Event Hub client based solely on the connection string.
  ## This can either be the associated environment variable or hardcoded directly.
  # connection_string = "$EVENTHUB_CONNECTION_STRING"

  ## Set persistence directory to a valid folder to use a file persister instead of an in-memory persister
  # persistence_dir = ""

  ## Change the default consumer group
  # consumer_group = ""

  ## By default the event hub receives all messages present on the broker.
  ## Alternative modes can be set below. The timestamp should be in RFC3339 format.
  ## The 3 options below only apply if no valid persister is read from memory or file (e.g. first run).
  # from_timestamp = ""
  # starting_offset = ""
  # latest = true

  ## Set a custom prefetch count for the receiver(s)
  # prefetch_count = 1000

  ## Add an epoch to the receiver(s)
  # epoch = 0

  ## Change to set a custom user agent, "telegraf" is used by default
  # user_agent = "telegraf"
  
  ## To consume from a specific partition, set the partition_ids option. 
  ## An empty array will result in receiving from all partitions.
  # partition_ids = ["0","1"]

  ## Max undelivered messages
  # max_undelivered_messages = 1000

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
  `
}

// Description of the plugin
func (*EventHub) Description() string {
	return "Azure Event Hubs service input plugin"
}

// SetParser sets the parser
func (e *EventHub) SetParser(parser parsers.Parser) {
	e.parser = parser
}

// Gather function is unused
func (*EventHub) Gather(telegraf.Accumulator) error {
	return nil
}

// Start the EventHub ServiceInput
func (e *EventHub) Start(acc telegraf.Accumulator) error {

	// Set hub options
	hubOpts := []eventhub.HubOption{}

	if e.PersistenceDir != "" {
		persister, err := persist.NewFilePersister(e.PersistenceDir)

		if err != nil {
			return err
		}

		hubOpts = append(hubOpts, eventhub.HubWithOffsetPersistence(persister))
	}

	if e.UserAgent != "" {
		hubOpts = append(hubOpts, eventhub.HubWithUserAgent(e.UserAgent))
	} else {
		hubOpts = append(hubOpts, eventhub.HubWithUserAgent("telegraf"))
	}

	// Create event hub connection
	var err error
	if e.ConnectionString != "" {
		e.hub, err = eventhub.NewHubFromConnectionString(e.ConnectionString, hubOpts...)
	} else {
		e.hub, err = eventhub.NewHubFromEnvironment(hubOpts...)
	}

	if err != nil {
		return err
	}

	// Init metric tracking
	e.acc = acc.WithTracking(e.MaxUndeliveredMessages)
	e.tracker = MessageTracker{messages: make(map[telegraf.TrackingID][]telegraf.Metric)}

	// Start tracking
	// e.wg.Add(1)
	go e.startTracking()

	var ctx context.Context
	ctx, e.cancel = context.WithCancel(context.Background())

	// Get runtime information
	runtimeinfo, err := e.hub.GetRuntimeInformation(ctx)

	if err != nil {
		return err
	}

	// Handler function to handle event hub events
	handler := func(c context.Context, event *eventhub.Event) error {

		metrics, err := e.parser.Parse(event.Data)

		if err != nil {
			log.Printf("E! [inputs.eventhub] %s", err)
			return err
		}

		log.Printf("D! [inputs.eventhub] %d metrics found after parsing", len(metrics))

		id := e.acc.AddTrackingMetricGroup(metrics)

		e.tracker.mux.Lock()
		e.tracker.messages[id] = metrics
		e.tracker.mux.Unlock()

		return nil
	}

	// Set receiver options
	receiveOpts := []eventhub.ReceiveOption{}

	if e.ConsumerGroup != "" {
		receiveOpts = append(receiveOpts, eventhub.ReceiveWithConsumerGroup(e.ConsumerGroup))
	}

	if e.FromTimestamp != "" {
		ts, err := time.Parse(time.RFC3339, e.FromTimestamp)

		if err != nil {
			return err
		}

		receiveOpts = append(receiveOpts, eventhub.ReceiveFromTimestamp(ts))

	} else if e.StartingOffset != "" {
		receiveOpts = append(receiveOpts, eventhub.ReceiveWithStartingOffset(e.StartingOffset))
	} else if e.Latest {
		receiveOpts = append(receiveOpts, eventhub.ReceiveWithLatestOffset())
	}

	if e.PrefetchCount != 0 {
		receiveOpts = append(receiveOpts, eventhub.ReceiveWithPrefetchCount(e.PrefetchCount))
	}

	if e.Epoch != 0 {
		receiveOpts = append(receiveOpts, eventhub.ReceiveWithEpoch(e.Epoch))
	}

	if len(e.PartitionIDs) == 0 {
		// Default behavior: receive from all partitions

		for _, partitionID := range runtimeinfo.PartitionIDs {

			_, err = e.hub.Receive(ctx, partitionID, handler, receiveOpts...)

			if err != nil {
				log.Printf("E! [inputs.eventhub] error creating receiver for partition %v", partitionID)
				return err
			}
		}
	} else {
		// Custom behavior: receive from a subset of partitions
		// Explicit check for valid partition selection, built in error handling is unreliable

		// Create map of valid partitions
		idlist := make(map[string]bool)

		for _, partitionID := range runtimeinfo.PartitionIDs {
			idlist[partitionID] = false
		}

		// Loop over selected partitions
		for _, partitionID := range e.PartitionIDs {

			// Check if partition exists on event hub
			if _, ok := idlist[partitionID]; ok {
				_, err = e.hub.Receive(ctx, partitionID, handler, receiveOpts...)

				if err != nil {
					log.Printf("E! [inputs.eventhub] error creating receiver for partition %v", partitionID)
					return err
				}
			} else {
				log.Printf("E! [inputs.eventhub] selected partition with ID \"%s\" not found on event hub", partitionID)
			}
		}
	}

	return nil
}

// startTracking monitors the message tracker and delivery info
func (e *EventHub) startTracking() {
	// defer e.wg.Done()

	for DeliveryInfo := range e.acc.Delivered() {

		log.Printf("D! [inputs.eventhub] tracking:: ID: %v - delivered: %v", DeliveryInfo.ID(), DeliveryInfo.Delivered())
		log.Printf("D! [inputs.eventhub] tracking:: message queue length: %d", len(e.tracker.messages))

		if DeliveryInfo.Delivered() {
			e.tracker.mux.Lock()
			delete(e.tracker.messages, DeliveryInfo.ID())
			e.tracker.mux.Unlock()

			log.Printf("D! [inputs.eventhub] tracking:: deleted ID %d from tracked message queue", DeliveryInfo.ID())
		} else {
			log.Printf("E! [inputs.eventhub] tracking:: undelivered message ID %d, retrying", DeliveryInfo.ID())

			e.tracker.mux.Lock()
			id := e.acc.AddTrackingMetricGroup(e.tracker.messages[DeliveryInfo.ID()])
			e.tracker.messages[id] = e.tracker.messages[DeliveryInfo.ID()]
			delete(e.tracker.messages, DeliveryInfo.ID())
			e.tracker.mux.Unlock()
		}
	}
}

// Stop the EventHub ServiceInput
func (e *EventHub) Stop() {
	e.cancel()
	// e.wg.Wait()
	err := e.hub.Close(context.Background())

	if err != nil {
		log.Printf("E! [inputs.eventhub] error closing Azure EventHub connection: %s", err)
	}

	log.Printf("D! [inputs.eventhub] event hub connection closed")
}

func init() {
	inputs.Add("eventhub", func() telegraf.Input {
		return &EventHub{
			MaxUndeliveredMessages: 1000, // default value
		}
	})
}
