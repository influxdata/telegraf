//go:generate ../../../tools/readme_config_includer/generator
package microsoft_fabric

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/serializers/json"
)

type eventstream struct {
	partitionKey   string
	maxMessageSize config.Size
	timeout        config.Duration

	connectionString string
	log              telegraf.Logger
	client           *azeventhubs.ProducerClient
	options          azeventhubs.EventDataBatchOptions
	serializer       telegraf.Serializer
}

var confKeys = []string{"PartitionKey", "MaxMessageSize"}

func (e *eventstream) Init() error {
	serializer := &json.Serializer{
		TimestampUnits:  config.Duration(time.Nanosecond),
		TimestampFormat: time.RFC3339Nano,
	}
	if err := serializer.Init(); err != nil {
		return err
	}
	e.serializer = serializer
	if e.maxMessageSize > 0 {
		e.options.MaxBytes = uint64(e.maxMessageSize)
	}
	return nil
}

func (e *eventstream) Connect() error {
	cfg := &azeventhubs.ProducerClientOptions{
		ApplicationID: internal.FormatFullVersion(),
		RetryOptions:  azeventhubs.RetryOptions{MaxRetries: -1},
	}

	client, err := azeventhubs.NewProducerClientFromConnectionString(e.connectionString, "", cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	e.client = client

	return nil
}

func (e *eventstream) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.timeout))
	defer cancel()

	return e.client.Close(ctx)
}

func (e *eventstream) SetSerializer(serializer telegraf.Serializer) {
	e.serializer = serializer
}

func (e *eventstream) Write(metrics []telegraf.Metric) error {
	// This context is only used for creating the batches which should not timeout as this is
	// not an I/O operation. Therefore avoid setting a timeout here.
	ctx := context.Background()

	batchOptions := e.options
	batches := make(map[string]*azeventhubs.EventDataBatch)
	// Cant use `for _, m := range metrics` as we need to move back when a new batch needs to be created
	for i := 0; i < len(metrics); i++ {
		m := metrics[i]

		// Prepare the payload
		payload, err := e.serializer.Serialize(m)
		if err != nil {
			e.log.Errorf("Could not serialize metric: %v", err)
			e.log.Tracef("metric: %+v", m)
			continue
		}

		// Get the batcher for the chosen partition
		partition := "<default>"
		if e.partitionKey != "" {
			if key, ok := m.GetTag(e.partitionKey); ok {
				partition = key
			} else if key, ok := m.GetField(e.partitionKey); ok {
				if k, ok := key.(string); ok {
					partition = k
				}
			}
		}
		batchOptions.PartitionKey = &partition
		if _, found := batches[partition]; !found {
			batches[partition], err = e.client.NewEventDataBatch(ctx, &batchOptions)
			if err != nil {
				return fmt.Errorf("creating batch for partition %q failed: %w", partition, err)
			}
		}

		// Add the event to the partition and send it if the batch is full
		err = batches[partition].AddEventData(&azeventhubs.EventData{Body: payload}, nil)
		if err == nil {
			continue
		}

		// If the event doesn't fit into the batch anymore, send the batch
		if !errors.Is(err, azeventhubs.ErrEventDataTooLarge) {
			return fmt.Errorf("adding metric to batch for partition %q failed: %w", partition, err)
		}

		// The event is larger than the maximum allowed size so there
		// is nothing we can do here but have to drop the metric.
		if batches[partition].NumEvents() == 0 {
			e.log.Errorf("Metric with %d bytes exceeds the maximum allowed size and must be dropped!", len(payload))
			e.log.Tracef("metric: %+v", m)
			continue
		}
		if err := e.send(ctx, batches[partition]); err != nil {
			return fmt.Errorf("sending batch for partition %q failed: %w", partition, err)
		}

		// Create a new metric and reiterate over the current metric to be
		// added in the next iteration of the for loop.
		batches[partition], err = e.client.NewEventDataBatch(ctx, &e.options)
		if err != nil {
			return fmt.Errorf("creating batch for partition %q failed: %w", partition, err)
		}
		i--
	}

	// Send the remaining batches that never exceeded the batch size
	for partition, batch := range batches {
		if batch.NumBytes() == 0 {
			continue
		}
		if err := e.send(ctx, batch); err != nil {
			return fmt.Errorf("sending batch for partition %q failed: %w", partition, err)
		}
	}
	return nil
}

func (e *eventstream) parseconnectionString(cs string) error {
	// Parse the connection string
	if cs == "" {
		return errors.New("connection string must not be empty")
	}
	// Split the connection string into key-value pairs
	pairs := strings.Split(cs, ";")
	for _, pair := range pairs {
		// Split each pair into key and value
		k, v, found := strings.Cut(pair, "=")
		if !found {
			return fmt.Errorf("invalid connection string format: %s", pair)
		}
		k = strings.ToLower(strings.TrimSpace(k))
		v = strings.TrimSpace(v)
		if slices.Contains(confKeys, k) {
			switch k {
			case "partitionkey", "partition key":
				e.partitionKey = v
			case "maxmessagesize", "max message size":
				if sz, err := strconv.ParseInt(v, 10, 64); err != nil {
					e.maxMessageSize = config.Size(sz)
				}
			}
		}
	}
	return nil
}

func (e *eventstream) send(ctx context.Context, batch *azeventhubs.EventDataBatch) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(e.timeout))
	defer cancel()

	return e.client.SendEventDataBatch(ctx, batch, nil)
}
