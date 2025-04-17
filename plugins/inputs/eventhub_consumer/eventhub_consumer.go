//go:generate ../../../tools/readme_config_includer/generator
package eventhub_consumer

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"sync"
	"time"

	eventhub "github.com/Azure/azure-event-hubs-go/v3"
	"github.com/Azure/azure-event-hubs-go/v3/persist"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

var once sync.Once

const (
	defaultMaxUndeliveredMessages = 1000
)

type EventHub struct {
	// Configuration
	ConnectionString       string    `toml:"connection_string"`
	PersistenceDir         string    `toml:"persistence_dir"`
	ConsumerGroup          string    `toml:"consumer_group"`
	FromTimestamp          time.Time `toml:"from_timestamp"`
	Latest                 bool      `toml:"latest"`
	PrefetchCount          uint32    `toml:"prefetch_count"`
	Epoch                  int64     `toml:"epoch"`
	UserAgent              string    `toml:"user_agent"`
	PartitionIDs           []string  `toml:"partition_ids"`
	MaxUndeliveredMessages int       `toml:"max_undelivered_messages"`
	EnqueuedTimeAsTS       bool      `toml:"enqueued_time_as_ts"`
	IotHubEnqueuedTimeAsTS bool      `toml:"iot_hub_enqueued_time_as_ts"`

	// Metadata
	ApplicationPropertyFields     []string `toml:"application_property_fields"`
	ApplicationPropertyTags       []string `toml:"application_property_tags"`
	SequenceNumberField           string   `toml:"sequence_number_field"`
	EnqueuedTimeField             string   `toml:"enqueued_time_field"`
	OffsetField                   string   `toml:"offset_field"`
	PartitionIDTag                string   `toml:"partition_id_tag"`
	PartitionKeyTag               string   `toml:"partition_key_tag"`
	IoTHubDeviceConnectionIDTag   string   `toml:"iot_hub_device_connection_id_tag"`
	IoTHubAuthGenerationIDTag     string   `toml:"iot_hub_auth_generation_id_tag"`
	IoTHubConnectionAuthMethodTag string   `toml:"iot_hub_connection_auth_method_tag"`
	IoTHubConnectionModuleIDTag   string   `toml:"iot_hub_connection_module_id_tag"`
	IoTHubEnqueuedTimeField       string   `toml:"iot_hub_enqueued_time_field"`

	Log telegraf.Logger `toml:"-"`

	// Azure
	hub    *eventhub.Hub
	cancel context.CancelFunc
	wg     sync.WaitGroup

	parser telegraf.Parser
	in     chan []telegraf.Metric
}

type (
	empty     struct{}
	semaphore chan empty
)

func (*EventHub) SampleConfig() string {
	return sampleConfig
}

func (e *EventHub) Init() (err error) {
	if e.MaxUndeliveredMessages == 0 {
		e.MaxUndeliveredMessages = defaultMaxUndeliveredMessages
	}

	// Set hub options
	hubOpts := make([]eventhub.HubOption, 0, 2)

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

func (e *EventHub) SetParser(parser telegraf.Parser) {
	e.parser = parser
}

func (e *EventHub) Start(acc telegraf.Accumulator) error {
	e.in = make(chan []telegraf.Metric)

	var ctx context.Context
	ctx, e.cancel = context.WithCancel(context.Background())

	// Start tracking
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		e.startTracking(ctx, acc)
	}()

	// Configure receiver options
	receiveOpts := e.configureReceiver()
	partitions := e.PartitionIDs

	if len(e.PartitionIDs) == 0 {
		runtimeinfo, err := e.hub.GetRuntimeInformation(ctx)
		if err != nil {
			return err
		}

		partitions = runtimeinfo.PartitionIDs
	}

	for _, partitionID := range partitions {
		_, err := e.hub.Receive(ctx, partitionID, e.onMessage, receiveOpts...)
		if err != nil {
			return fmt.Errorf("creating receiver for partition %q: %w", partitionID, err)
		}
	}

	return nil
}

func (*EventHub) Gather(telegraf.Accumulator) error {
	return nil
}

func (e *EventHub) Stop() {
	err := e.hub.Close(context.Background())
	if err != nil {
		e.Log.Errorf("Error closing Event Hub connection: %v", err)
	}
	e.cancel()
	e.wg.Wait()
}

func (e *EventHub) configureReceiver() []eventhub.ReceiveOption {
	receiveOpts := make([]eventhub.ReceiveOption, 0, 4)

	if e.ConsumerGroup != "" {
		receiveOpts = append(receiveOpts, eventhub.ReceiveWithConsumerGroup(e.ConsumerGroup))
	}

	if !e.FromTimestamp.IsZero() {
		receiveOpts = append(receiveOpts, eventhub.ReceiveFromTimestamp(e.FromTimestamp))
	} else if e.Latest {
		receiveOpts = append(receiveOpts, eventhub.ReceiveWithLatestOffset())
	}

	if e.PrefetchCount != 0 {
		receiveOpts = append(receiveOpts, eventhub.ReceiveWithPrefetchCount(e.PrefetchCount))
	}

	if e.Epoch != 0 {
		receiveOpts = append(receiveOpts, eventhub.ReceiveWithEpoch(e.Epoch))
	}

	return receiveOpts
}

// OnMessage handles an Event.  When this function returns without error the
// Event is immediately accepted and the offset is updated.  If an error is
// returned the Event is marked for redelivery.
func (e *EventHub) onMessage(ctx context.Context, event *eventhub.Event) error {
	metrics, err := e.createMetrics(event)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case e.in <- metrics:
		return nil
	}
}

// OnDelivery returns true if a new slot has opened up in the TrackingAccumulator.
func (e *EventHub) onDelivery(
	acc telegraf.TrackingAccumulator,
	groups map[telegraf.TrackingID][]telegraf.Metric,
	track telegraf.DeliveryInfo,
) bool {
	if track.Delivered() {
		delete(groups, track.ID())
		return true
	}

	// The metric was already accepted when onMessage completed, so we can't
	// fallback on redelivery from Event Hub.  Add a new copy of the metric for
	// reprocessing.
	metrics, ok := groups[track.ID()]
	delete(groups, track.ID())
	if !ok {
		// The metrics should always be found, this message indicates a programming error.
		e.Log.Errorf("Could not find delivery: %d", track.ID())
		return true
	}

	backup := deepCopyMetrics(metrics)
	id := acc.AddTrackingMetricGroup(metrics)
	groups[id] = backup
	return false
}

func (e *EventHub) startTracking(ctx context.Context, ac telegraf.Accumulator) {
	acc := ac.WithTracking(e.MaxUndeliveredMessages)
	sem := make(semaphore, e.MaxUndeliveredMessages)
	groups := make(map[telegraf.TrackingID][]telegraf.Metric, e.MaxUndeliveredMessages)

	for {
		select {
		case <-ctx.Done():
			return
		case track := <-acc.Delivered():
			if e.onDelivery(acc, groups, track) {
				<-sem
			}
		case sem <- empty{}:
			select {
			case <-ctx.Done():
				return
			case track := <-acc.Delivered():
				if e.onDelivery(acc, groups, track) {
					<-sem
					<-sem
				}
			case metrics := <-e.in:
				backup := deepCopyMetrics(metrics)
				id := acc.AddTrackingMetricGroup(metrics)
				groups[id] = backup
			}
		}
	}
}

func deepCopyMetrics(in []telegraf.Metric) []telegraf.Metric {
	metrics := make([]telegraf.Metric, 0, len(in))
	for _, m := range in {
		metrics = append(metrics, m.Copy())
	}
	return metrics
}

// CreateMetrics returns the Metrics from the Event.
func (e *EventHub) createMetrics(event *eventhub.Event) ([]telegraf.Metric, error) {
	metrics, err := e.parser.Parse(event.Data)
	if err != nil {
		return nil, err
	}

	if len(metrics) == 0 {
		once.Do(func() {
			e.Log.Debug(internal.NoMetricsCreatedMsg)
		})
	}

	for i := range metrics {
		for _, field := range e.ApplicationPropertyFields {
			if val, ok := event.Get(field); ok {
				metrics[i].AddField(field, val)
			}
		}

		for _, tag := range e.ApplicationPropertyTags {
			if val, ok := event.Get(tag); ok {
				metrics[i].AddTag(tag, fmt.Sprintf("%v", val))
			}
		}

		if e.SequenceNumberField != "" {
			metrics[i].AddField(e.SequenceNumberField, *event.SystemProperties.SequenceNumber)
		}

		if e.EnqueuedTimeAsTS {
			metrics[i].SetTime(*event.SystemProperties.EnqueuedTime)
		} else if e.EnqueuedTimeField != "" {
			metrics[i].AddField(e.EnqueuedTimeField, (*event.SystemProperties.EnqueuedTime).UnixNano()/int64(time.Millisecond))
		}

		if e.OffsetField != "" {
			metrics[i].AddField(e.OffsetField, *event.SystemProperties.Offset)
		}

		if event.SystemProperties.PartitionID != nil && e.PartitionIDTag != "" {
			metrics[i].AddTag(e.PartitionIDTag, strconv.Itoa(int(*event.SystemProperties.PartitionID)))
		}
		if event.SystemProperties.PartitionKey != nil && e.PartitionKeyTag != "" {
			metrics[i].AddTag(e.PartitionKeyTag, *event.SystemProperties.PartitionKey)
		}
		if event.SystemProperties.IoTHubDeviceConnectionID != nil && e.IoTHubDeviceConnectionIDTag != "" {
			metrics[i].AddTag(e.IoTHubDeviceConnectionIDTag, *event.SystemProperties.IoTHubDeviceConnectionID)
		}
		if event.SystemProperties.IoTHubAuthGenerationID != nil && e.IoTHubAuthGenerationIDTag != "" {
			metrics[i].AddTag(e.IoTHubAuthGenerationIDTag, *event.SystemProperties.IoTHubAuthGenerationID)
		}
		if event.SystemProperties.IoTHubConnectionAuthMethod != nil && e.IoTHubConnectionAuthMethodTag != "" {
			metrics[i].AddTag(e.IoTHubConnectionAuthMethodTag, *event.SystemProperties.IoTHubConnectionAuthMethod)
		}
		if event.SystemProperties.IoTHubConnectionModuleID != nil && e.IoTHubConnectionModuleIDTag != "" {
			metrics[i].AddTag(e.IoTHubConnectionModuleIDTag, *event.SystemProperties.IoTHubConnectionModuleID)
		}
		if event.SystemProperties.IoTHubEnqueuedTime != nil {
			if e.IotHubEnqueuedTimeAsTS {
				metrics[i].SetTime(*event.SystemProperties.IoTHubEnqueuedTime)
			} else if e.IoTHubEnqueuedTimeField != "" {
				metrics[i].AddField(e.IoTHubEnqueuedTimeField, (*event.SystemProperties.IoTHubEnqueuedTime).UnixNano()/int64(time.Millisecond))
			}
		}
	}

	return metrics, nil
}

func init() {
	inputs.Add("eventhub_consumer", func() telegraf.Input {
		return &EventHub{}
	})
}
