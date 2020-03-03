package eventhub

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"

	eventhub "github.com/Azure/azure-event-hubs-go/v3"
	"github.com/Azure/azure-event-hubs-go/v3/persist"
)

const (
	defaultMaxUndeliveredMessages = 1000
	defaultSystemPropertiesPrefix = "SystemProperties."
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
	SystemPropertiesPrefix string   `toml:"system_properties_prefix"`
	EnqTimeTs              bool     `toml:"enq_time_ts"`
	IotHubEnqTimeTs        bool     `toml:"iot_hub_enq_time_ts"`

	// Azure
	hub    *eventhub.Hub
	cancel context.CancelFunc

	// Influx
	parser parsers.Parser

	// Metrics tracking
	acc     telegraf.TrackingAccumulator
	tracker MessageTracker
	wg      sync.WaitGroup
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
  ## This can either be the associated environment variable or hard coded directly.
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

  ## Prefix to use for the system properties of Event Hub and IoT Hub messages
  # system_properties_prefix = "SystemProperties."

  ## Set either option below to true to use a system property as timestamp.
  ## You have the choice between EnqueuedTime and IoTHubEnqueuedTime.
  ## It is recommended to use this setting when the data itself has no timestamp.
  # enq_time_ts = true
  # iot_hub_enq_time_ts = true

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

// Init the EventHub ServiceInput
func (e *EventHub) Init() (err error) {
	if e.MaxUndeliveredMessages == 0 {
		e.MaxUndeliveredMessages = defaultMaxUndeliveredMessages
	}

	if e.SystemPropertiesPrefix == "" {
		e.SystemPropertiesPrefix = defaultSystemPropertiesPrefix
	}

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
		hubOpts = append(hubOpts, eventhub.HubWithUserAgent(internal.ProductToken()))
	}

	// Create event hub connection
	if e.ConnectionString != "" {
		e.hub, err = eventhub.NewHubFromConnectionString(e.ConnectionString, hubOpts...)
	} else {
		e.hub, err = eventhub.NewHubFromEnvironment(hubOpts...)
	}

	return err
}

// Start the EventHub ServiceInput
func (e *EventHub) Start(acc telegraf.Accumulator) error {

	// Init metric tracking
	e.acc = acc.WithTracking(e.MaxUndeliveredMessages)
	e.tracker = MessageTracker{messages: make(map[telegraf.TrackingID][]telegraf.Metric, e.MaxUndeliveredMessages)}

	var ctx context.Context
	ctx, e.cancel = context.WithCancel(context.Background())

	// Start tracking
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		e.startTracking(ctx)
	}()

	// Configure receiver options
	receiveOpts, err := e.configureReceiver()
	if err != nil {
		return err
	}

	if len(e.PartitionIDs) == 0 { // Default behavior: receive from all partitions

		// Get runtime information
		runtimeinfo, err := e.hub.GetRuntimeInformation(ctx)
		if err != nil {
			return err
		}

		for _, partitionID := range runtimeinfo.PartitionIDs {

			_, err = e.hub.Receive(ctx, partitionID, e.onMessage, receiveOpts...)
			if err != nil {
				log.Printf("E! [inputs.eventhub] error creating receiver for partition \"%s\"", partitionID)
				return err
			}
		}
	} else { // Custom behavior: receive from a subset of partitions

		for _, partitionID := range e.PartitionIDs {

			_, err = e.hub.Receive(ctx, partitionID, e.onMessage, receiveOpts...)
			if err != nil {
				log.Printf("E! [inputs.eventhub] error creating receiver for partition \"%s\"", partitionID)
				return err
			}
		}
	}

	return nil
}

func (e *EventHub) configureReceiver() (receiveOpts []eventhub.ReceiveOption, err error) {

	if e.ConsumerGroup != "" {
		receiveOpts = append(receiveOpts, eventhub.ReceiveWithConsumerGroup(e.ConsumerGroup))
	}

	if e.FromTimestamp != "" {
		ts, err := time.Parse(time.RFC3339, e.FromTimestamp)
		if err != nil {
			log.Printf("E! [inputs.eventhub] error in parsing timestamp: %s", err)
			return receiveOpts, err
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

	return receiveOpts, err
}

func (e *EventHub) onMessage(ctx context.Context, event *eventhub.Event) (err error) {

	metrics, err := e.parser.Parse(event.Data)
	if err != nil {
		log.Printf("E! [inputs.eventhub] error %s", err)
		return err
	}

	log.Printf("D! [inputs.eventhub] %d metrics found after parsing", len(metrics))

	for i := range metrics {
		metrics[i].AddField(e.SystemPropertiesPrefix+"SequenceNumber", *event.SystemProperties.SequenceNumber)

		if e.EnqTimeTs {
			metrics[i].SetTime(*event.SystemProperties.EnqueuedTime)
		} else {
			metrics[i].AddField(e.SystemPropertiesPrefix+"EnqueuedTime", (*event.SystemProperties.EnqueuedTime).UnixNano()/int64(time.Millisecond))
		}

		metrics[i].AddField(e.SystemPropertiesPrefix+"Offset", *event.SystemProperties.Offset)

		if event.SystemProperties.PartitionID != nil {
			metrics[i].AddTag(e.SystemPropertiesPrefix+"PartitionID", string(*event.SystemProperties.PartitionID))
		}
		if event.SystemProperties.PartitionKey != nil {
			metrics[i].AddTag(e.SystemPropertiesPrefix+"PartitionKey", *event.SystemProperties.PartitionKey)
		}
		if event.SystemProperties.IoTHubDeviceConnectionID != nil {
			metrics[i].AddTag(e.SystemPropertiesPrefix+"IoTHubDeviceConnectionID", *event.SystemProperties.IoTHubDeviceConnectionID)
		}
		if event.SystemProperties.IoTHubAuthGenerationID != nil {
			metrics[i].AddTag(e.SystemPropertiesPrefix+"IoTHubAuthGenerationID", *event.SystemProperties.IoTHubAuthGenerationID)
		}
		if event.SystemProperties.IoTHubConnectionAuthMethod != nil {
			metrics[i].AddTag(e.SystemPropertiesPrefix+"IoTHubConnectionAuthMethod", *event.SystemProperties.IoTHubConnectionAuthMethod)
		}
		if event.SystemProperties.IoTHubConnectionModuleID != nil {
			metrics[i].AddTag(e.SystemPropertiesPrefix+"IoTHubConnectionModuleID", *event.SystemProperties.IoTHubConnectionModuleID)
		}
		if event.SystemProperties.IoTHubEnqueuedTime != nil {
			if e.IotHubEnqTimeTs {
				metrics[i].SetTime(*event.SystemProperties.IoTHubEnqueuedTime)
			} else {
				metrics[i].AddField(e.SystemPropertiesPrefix+"IoTHubEnqueuedTime", (*event.SystemProperties.IoTHubEnqueuedTime).UnixNano()/int64(time.Millisecond))
			}
		}
	}

	id := e.acc.AddTrackingMetricGroup(metrics)

	e.tracker.mux.Lock()
	e.tracker.messages[id] = metrics
	e.tracker.mux.Unlock()

	return nil
}

func (e *EventHub) startTracking(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			if len(e.tracker.messages) == 0 { // everything has been delivered
				return
			}
		case DeliveryInfo := <-e.acc.Delivered():
			log.Printf("D! [inputs.eventhub] tracking:: ID %d delivered: %t", DeliveryInfo.ID(), DeliveryInfo.Delivered())

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

			log.Printf("D! [inputs.eventhub] tracking:: message queue length: %d", len(e.tracker.messages))
		}
	}
}

// Stop the EventHub ServiceInput
func (e *EventHub) Stop() {
	e.cancel()
	err := e.hub.Close(context.Background())

	e.wg.Wait()
	if err != nil {
		log.Printf("E! [inputs.eventhub] error in closing event hub connection: %s", err)
	}

	log.Printf("D! [inputs.eventhub] event hub connection closed")
}

func init() {
	inputs.Add("eventhub", func() telegraf.Input {
		return &EventHub{}
	})
}
