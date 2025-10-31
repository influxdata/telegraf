//go:generate ../../../tools/readme_config_includer/generator
package microsoft_fabric

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/serializers/json"
)

type eventstream struct {
	connectionString string
	timeout          config.Duration
	log              telegraf.Logger

	partitionKey   string
	maxMessageSize config.Size

	client     *azeventhubs.ProducerClient
	options    azeventhubs.EventDataBatchOptions
	serializer telegraf.Serializer
}

func (e *eventstream) init() error {
	// Parse the connection string by splitting it into key-value pairs
	// and extract the extra keys used for plugin configuration
	pairs := strings.Split(e.connectionString, ";")
	for _, pair := range pairs {
		// Skip empty pairs
		if strings.TrimSpace(pair) == "" {
			continue
		}
		// Split each pair into key and value
		k, v, found := strings.Cut(pair, "=")
		if !found {
			return fmt.Errorf("invalid connection string format: %q", pair)
		}

		// Only lowercase the keys as the values might be case sensitive
		k = strings.ToLower(strings.TrimSpace(k))
		v = strings.TrimSpace(v)

		key := strings.ReplaceAll(k, " ", "")
		switch key {
		case "partitionkey":
			e.partitionKey = v
		case "maxmessagesize":
			msgsize, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid max message size: %w", err)
			}
			if msgsize > 0 {
				e.options.MaxBytes = msgsize
			}
		}
	}

	// Setup the JSON serializer
	serializer := &json.Serializer{
		TimestampUnits:  config.Duration(time.Nanosecond),
		TimestampFormat: time.RFC3339Nano,
	}
	if err := serializer.Init(); err != nil {
		return fmt.Errorf("setting up JSON serializer failed: %w", err)
	}
	e.serializer = serializer

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
	if e.client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.timeout))
	defer cancel()

	return e.client.Close(ctx)
}

func (e *eventstream) Write(metrics []telegraf.Metric) error {
	// This context is only used for creating the batches which should not timeout as this is
	// not an I/O operation. Therefore avoid setting a timeout here.
	ctx := context.Background()

	// Iterate over the metrics and group them to batches
	batchOptions := e.options
	batches := make(map[string]*azeventhubs.EventDataBatch)
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

func (e *eventstream) send(ctx context.Context, batch *azeventhubs.EventDataBatch) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(e.timeout))
	defer cancel()

	return e.client.SendEventDataBatch(ctx, batch, nil)
}

func isEventstreamEndpoint(endpoint string) bool {
	return strings.HasPrefix(strings.ToLower(endpoint), "endpoint=sb")
}
