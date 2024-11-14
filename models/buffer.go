package models

import (
	"fmt"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/selfstat"
)

var (
	AgentMetricsWritten = selfstat.Register("agent", "metrics_written", make(map[string]string))
	AgentMetricsDropped = selfstat.Register("agent", "metrics_dropped", make(map[string]string))

	registerGob = sync.OnceFunc(func() { metric.Init() })
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

	// Stats returns the buffer statistics such as rejected, dropped and accepred metrics
	Stats() BufferStats

	// Close finalizes the buffer and closes all open resources
	Close() error
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
func NewBuffer(name, id, alias string, capacity int, strategy, path string) (Buffer, error) {
	registerGob()

	bs := NewBufferStats(name, alias, capacity)

	switch strategy {
	case "", "memory":
		return NewMemoryBuffer(capacity, bs)
	case "disk":
		return NewDiskBuffer(name, id, path, bs)
	}
	return nil, fmt.Errorf("invalid buffer strategy %q", strategy)
}

func NewBufferStats(name, alias string, capacity int) BufferStats {
	tags := map[string]string{"output": name}
	if alias != "" {
		tags["alias"] = alias
	}

	bs := BufferStats{
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
	bs.BufferSize.Set(int64(0))
	bs.BufferLimit.Set(int64(capacity))
	return bs
}

func (b *BufferStats) metricAdded() {
	b.MetricsAdded.Incr(1)
}

func (b *BufferStats) metricWritten(m telegraf.Metric) {
	AgentMetricsWritten.Incr(1)
	b.MetricsWritten.Incr(1)
	m.Accept()
}

func (b *BufferStats) metricDropped(m telegraf.Metric) {
	AgentMetricsDropped.Incr(1)
	b.MetricsDropped.Incr(1)
	m.Reject()
}
