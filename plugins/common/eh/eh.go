package eh

import (
	"context"
	"time"

	eventhub "github.com/Azure/azure-event-hubs-go/v3"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
)

var sampleConfig string

/*
** Wrapper interface for eventhub.Hub
 */

type EventHubInterface interface {
	GetHub(s string) error
	Close(ctx context.Context) error
	SendBatch(ctx context.Context, iterator eventhub.BatchIterator, opts ...eventhub.BatchOption) error
}

type EventHub struct {
	hub *eventhub.Hub
}

func (eh *EventHub) GetHub(s string) error {
	hub, err := eventhub.NewHubFromConnectionString(s)

	if err != nil {
		return err
	}

	eh.hub = hub

	return nil
}

func (eh *EventHub) Close(ctx context.Context) error {
	return eh.hub.Close(ctx)
}

func (eh *EventHub) SendBatch(ctx context.Context, iterator eventhub.BatchIterator, opts ...eventhub.BatchOption) error {
	return eh.hub.SendBatch(ctx, iterator, opts...)
}

/* End wrapper interface */

type EventHubs struct {
	Log              telegraf.Logger `toml:"-"`
	ConnectionString string          `toml:"connection_string"`
	Timeout          config.Duration `toml:"timeout"`
	PartitionKey     string          `toml:"partition_key"`
	MaxMessageSize   int             `toml:"max_message_size"`

	Hub EventHubInterface
}

func (*EventHubs) SampleConfig() string {
	return sampleConfig
}

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
