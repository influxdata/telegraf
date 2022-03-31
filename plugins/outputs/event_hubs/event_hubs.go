package event_hubs

import (
	"context"
	"time"

	eventhub "github.com/Azure/azure-event-hubs-go/v3"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

/*
** Wrapper interface for eventhub.Hub
 */

type EventHubInterface interface {
	GetHub(s string) error
	Close(ctx context.Context) error
	SendBatch(ctx context.Context, iterator eventhub.BatchIterator, opts ...eventhub.BatchOption) error
}

type eventHub struct {
	hub *eventhub.Hub
}

func (eh *eventHub) GetHub(s string) error {
	hub, err := eventhub.NewHubFromConnectionString(s)

	if err != nil {
		return err
	}

	eh.hub = hub

	return nil
}

func (eh *eventHub) Close(ctx context.Context) error {
	return eh.hub.Close(ctx)
}

func (eh *eventHub) SendBatch(ctx context.Context, iterator eventhub.BatchIterator, opts ...eventhub.BatchOption) error {
	return eh.hub.SendBatch(ctx, iterator, opts...)
}

/* End wrapper interface */

type EventHubs struct {
	Log              telegraf.Logger `toml:"-"`
	ConnectionString string          `toml:"connection_string"`
	Timeout          config.Duration

	Hub        EventHubInterface
	serializer serializers.Serializer
}

const (
	defaultRequestTimeout = time.Second * 30
)

func (e *EventHubs) Init() error {
	err := e.Hub.GetHub(e.ConnectionString)

	if err != nil {
		return err
	}

	return nil
}

func (e *EventHubs) Connect() error {
	return nil
}

func (e *EventHubs) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.Timeout))
	defer cancel()

	err := e.Hub.Close(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (e *EventHubs) SetSerializer(serializer serializers.Serializer) {
	e.serializer = serializer
}

func (e *EventHubs) Write(metrics []telegraf.Metric) error {
	var events []*eventhub.Event

	for _, metric := range metrics {
		payload, err := e.serializer.Serialize(metric)

		if err != nil {
			e.Log.Debugf("Could not serialize metric: %v", err)
			continue
		}

		events = append(events, eventhub.NewEvent(payload))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.Timeout))
	defer cancel()

	err := e.Hub.SendBatch(ctx, eventhub.NewEventBatchIterator(events...))

	if err != nil {
		return err
	}

	return nil
}

func init() {
	outputs.Add("event_hubs", func() telegraf.Output {
		return &EventHubs{
			Hub:     &eventHub{},
			Timeout: config.Duration(defaultRequestTimeout),
		}
	})
}
