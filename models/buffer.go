package models

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
)

var (
	AgentMetricsWritten = selfstat.Register("agent", "metrics_written", map[string]string{})
	AgentMetricsDropped = selfstat.Register("agent", "metrics_dropped", map[string]string{})
)

type Buffer interface {

	// Len returns the number of metrics currently in the buffer.
	Len() int

	// Add adds metrics to the buffer and returns number of dropped metrics.
	Add(metrics ...telegraf.Metric) int

	// Batch returns a slice containing up to batchSize of the oldest metrics not
	// yet dropped.  Metrics are ordered from oldest to newest in the batch.  The
	// batch must not be modified by the client.
	Batch(batchSize int) []telegraf.Metric

	// Accept marks the batch, acquired from Batch(), as successfully written.
	Accept(metrics []telegraf.Metric)

	// Reject returns the batch, acquired from Batch(), to the buffer and marks it
	// as unsent.
	Reject([]telegraf.Metric)

	Stats() BufferStats
}

// BufferStats holds common metrics used for buffer implementations.
// Implementations of Buffer should embed this struct in them.
type BufferStats struct {
	MetricsAdded   selfstat.Stat
	MetricsWritten selfstat.Stat
	MetricsDropped selfstat.Stat
	BufferSize     selfstat.Stat
	BufferLimit    selfstat.Stat
}

// NewBuffer returns a new empty Buffer with the given capacity.
func NewBuffer(name string, alias string, capacity int, strategy string, path string) (Buffer, error) {
	bm := NewBufferMetrics(name, alias, capacity)

	switch strategy {
	case "", "memory":
		return NewMemoryBuffer(capacity, bm)
	case "disk":
		return NewDiskBuffer(name, path, bm)
	}

	return nil, fmt.Errorf("invalid buffer strategy %q", strategy)
}

func NewBufferMetrics(name string, alias string, capacity int) BufferStats {
	tags := map[string]string{"output": name}
	if alias != "" {
		tags["alias"] = alias
	}

	bm := BufferStats{
		MetricsAdded: selfstat.Register(
			"write",
			"metrics_added",
			tags,
		),
		MetricsWritten: selfstat.Register(
			"write",
			"metrics_written",
			tags,
		),
		MetricsDropped: selfstat.Register(
			"write",
			"metrics_dropped",
			tags,
		),
		BufferSize: selfstat.Register(
			"write",
			"buffer_size",
			tags,
		),
		BufferLimit: selfstat.Register(
			"write",
			"buffer_limit",
			tags,
		),
	}
	bm.BufferSize.Set(int64(0))
	bm.BufferLimit.Set(int64(capacity))
	return bm
}

func (b *BufferStats) metricAdded() {
	b.MetricsAdded.Incr(1)
}

func (b *BufferStats) metricWritten(metric telegraf.Metric) {
	AgentMetricsWritten.Incr(1)
	b.MetricsWritten.Incr(1)
	metric.Accept()
}

func (b *BufferStats) metricDropped(metric telegraf.Metric) {
	AgentMetricsDropped.Incr(1)
	b.MetricsDropped.Incr(1)
	metric.Reject()
}
