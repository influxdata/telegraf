//go:generate ../../../tools/readme_config_includer/generator
package event_hubs

import (
	"context"
	_ "embed"
	"time"

	eventhub "github.com/Azure/azure-event-hubs-go/v3"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	eh_commons "github.com/influxdata/telegraf/plugins/common/eh"

	"github.com/influxdata/telegraf/plugins/outputs"
)

/* End wrapper interface */

type EventHubs struct {
	eh_commons.EventHubs
	serializer   telegraf.Serializer
	batchOptions []eventhub.BatchOption
}

const (
	defaultRequestTimeout = time.Second * 30
)

//go:embed sample.conf
var sampleConfig string

func (e *EventHubs) SampleConfig() string {
	return sampleConfig
}

func (e *EventHubs) Init() error {

	if e.MaxMessageSize > 0 {
		e.batchOptions = append(e.batchOptions, eventhub.BatchWithMaxSizeInBytes(e.MaxMessageSize))
	}

	return e.EventHubs.Init()
}

func (e *EventHubs) Connect() error {
	return nil
}

func (e *EventHubs) Close() error {
	return e.EventHubs.Close()
}

func (e *EventHubs) SetSerializer(serializer telegraf.Serializer) {
	e.serializer = serializer
}

func (e *EventHubs) Write(metrics []telegraf.Metric) error {
	events := make([]*eventhub.Event, 0, len(metrics))
	for _, metric := range metrics {
		payload, err := e.serializer.Serialize(metric)

		if err != nil {
			e.Log.Debugf("Could not serialize metric: %v", err)
			continue
		}

		event := eventhub.NewEvent(payload)
		if e.PartitionKey != "" {
			if key, ok := metric.GetTag(e.PartitionKey); ok {
				event.PartitionKey = &key
			} else if key, ok := metric.GetField(e.PartitionKey); ok {
				if strKey, ok := key.(string); ok {
					event.PartitionKey = &strKey
				}
			}
		}

		events = append(events, event)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.Timeout))
	defer cancel()

	err := e.Hub.SendBatch(ctx, eventhub.NewEventBatchIterator(events...), e.batchOptions...)

	if err != nil {
		return err
	}

	return nil
}

func init() {
	outputs.Add("event_hubs", func() telegraf.Output {
		return &EventHubs{
			EventHubs: eh_commons.EventHubs{
				Hub:     &eh_commons.EventHub{},
				Timeout: config.Duration(defaultRequestTimeout),
			},
		}
	})
}
