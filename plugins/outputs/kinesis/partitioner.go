package kinesis

import (
	"github.com/influxdata/telegraf"
	uuid "github.com/satori/go.uuid"
)

const (
	// Maximum metrics to place in a single record.
	//
	// A single record can hold up to 1 MB of data.  It is expected that 1000
	// metrics will be approximately 500 KB without compression.
	maxMetricsRecord = 1000
)

type Partition struct {
	Key     string
	Metrics []telegraf.Metric
}

// SingleRecordPartitioner handles partitioning for all cases when not using
// batch format.
type SingleRecordPartitioner struct {
	config *PartitionConfig
}

func (p *SingleRecordPartitioner) Partition(metrics []telegraf.Metric) []Partition {
	var partitions []Partition
	for _, metric := range metrics {
		partitions = append(partitions, Partition{
			Key:     partitionKey(p.config, metric),
			Metrics: []telegraf.Metric{metric},
		})
	}
	return partitions
}

// FixedBatchPartitioner handles partitioning when *not* using a random
// partition keys and batch format.
type FixedBatchPartitioner struct {
	config *PartitionConfig
}

func (p *FixedBatchPartitioner) Partition(metrics []telegraf.Metric) []Partition {
	// Partition metrics based on their partition key
	parts := make(map[string][]telegraf.Metric)
	for _, metric := range metrics {
		key := partitionKey(p.config, metric)
		if _, ok := parts[key]; !ok {
			parts[key] = make([]telegraf.Metric, 1)
		}
		parts[key] = append(parts[key], metric)
	}

	// Further restrict partitions to a fixed metric length to avoid exceeding
	// AWS limits.
	var partitions []Partition
	for key, metrics := range parts {
		var batches [][]telegraf.Metric
		for maxMetricsRecord < len(metrics) {
			metrics = metrics[maxMetricsRecord:]
			batches = append(batches, metrics[0:maxMetricsRecord:maxMetricsRecord])
		}
		batches = append(batches, metrics)

		for _, batch := range batches {
			partitions = append(partitions, Partition{
				Key:     key,
				Metrics: batch,
			})
		}
	}
	return partitions
}

// RandomBatchPartitioner handles partitioning with random partition keys and
// batch format.
type RandomBatchPartitioner struct {
	config *PartitionConfig
}

func (p *RandomBatchPartitioner) Partition(metrics []telegraf.Metric) []Partition {
	var partitions []Partition
	var batches [][]telegraf.Metric
	for maxMetricsRecord < len(metrics) {
		metrics = metrics[maxMetricsRecord:]
		batches = append(batches, metrics[0:maxMetricsRecord:maxMetricsRecord])
	}
	batches = append(batches, metrics)

	for _, batch := range batches {
		partitions = append(partitions, Partition{
			Key:     uuid.NewV4().String(),
			Metrics: batch,
		})
	}
	return partitions
}

func partitionKey(config *PartitionConfig, metric telegraf.Metric) string {
	switch config.Method {
	case "static":
		return config.Key
	case "random":
		return uuid.NewV4().String()
	case "measurement":
		return metric.Name()
	case "tag":
		if t, ok := metric.GetTag(config.Key); ok {
			return t
		}
		return "telegraf"
	default:
		panic("unsupported partition method")
	}
}
