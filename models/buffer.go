package models

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
)

var (
	AgentMetricsWritten  = selfstat.Register("agent", "metrics_written", make(map[string]string))
	AgentMetricsRejected = selfstat.Register("agent", "metrics_rejected", make(map[string]string))
	AgentMetricsDropped  = selfstat.Register("agent", "metrics_dropped", make(map[string]string))
)

type Transaction struct {
	// Batch of metrics to write
	Batch []telegraf.Metric

	// Accept denotes the indices of metrics that were successfully written
	Accept []int
	// Reject denotes the indices of metrics that were not written but should
	// not be requeued
	Reject []int

	// Marks this transaction as valid
	valid bool

	// Internal state that can be used by the buffer implementation
	state any
}

func (tx *Transaction) AcceptAll() {
	tx.Accept = make([]int, len(tx.Batch))
	for i := range tx.Batch {
		tx.Accept[i] = i
	}
}

func (*Transaction) KeepAll() {}

func (tx *Transaction) InferKeep() []int {
	used := make([]bool, len(tx.Batch))
	for _, idx := range tx.Accept {
		used[idx] = true
	}
	for _, idx := range tx.Reject {
		used[idx] = true
	}

	keep := make([]int, 0, len(tx.Batch))
	for i := range tx.Batch {
		if !used[i] {
			keep = append(keep, i)
		}
	}
	return keep
}

type Buffer interface {
	// Len returns the number of metrics currently in the buffer.
	Len() int

	// Add adds metrics to the buffer and returns number of dropped metrics.
	Add(metrics ...telegraf.Metric) int

	// Batch starts a transaction by returning a slice of metrics up to the
	// given batch-size starting from the oldest metric in the buffer. Metrics
	// are ordered from oldest to newest and must not be modified by the plugin.
	BeginTransaction(batchSize int) *Transaction

	// Flush ends a metric and persists the buffer state
	EndTransaction(*Transaction)

	// Stats returns the buffer statistics such as rejected, dropped and accepted metrics
	Stats() BufferStats

	// Close finalizes the buffer and closes all open resources
	Close() error
}

// BufferStats holds common metrics used for buffer implementations.
// Implementations of Buffer should embed this struct in them.
type BufferStats struct {
	MetricsAdded    selfstat.Stat
	MetricsWritten  selfstat.Stat
	MetricsRejected selfstat.Stat
	MetricsDropped  selfstat.Stat
	BufferSize      selfstat.Stat
	BufferLimit     selfstat.Stat
}

const (
	// DropOldest makes a full buffer discard its oldest metrics to make room
	// for new ones. This is the default.
	DropOldest = "drop_oldest"
	// DropNewest makes a full buffer discard new metrics and keep the ones
	// that are already buffered.
	DropNewest = "drop_newest"
)

type bufferOptions struct {
	dropStrategy string
}

// BufferOption customizes the buffer created by NewBuffer.
type BufferOption func(*bufferOptions)

// WithDropStrategy sets which metrics a full buffer drops, either DropOldest
// (default) or DropNewest. It only affects buffers with a maximum capacity,
// i.e. the memory buffer.
func WithDropStrategy(strategy string) BufferOption {
	return func(o *bufferOptions) {
		o.dropStrategy = strategy
	}
}

// CheckDropStrategy verifies that the given drop strategy is valid.
func CheckDropStrategy(strategy string) error {
	switch strategy {
	case "", DropOldest, DropNewest:
		return nil
	}
	return fmt.Errorf("invalid buffer drop strategy %q", strategy)
}

// NewBuffer returns a new empty Buffer with the given capacity.
//
//nolint:revive //will move to structs later
func NewBuffer(name, id, alias string, capacity int, strategy, path string, diskSync bool, opts ...BufferOption) (Buffer, error) {
	var options bufferOptions
	for _, opt := range opts {
		opt(&options)
	}
	if err := CheckDropStrategy(options.dropStrategy); err != nil {
		return nil, err
	}

	tags := map[string]string{
		"_id":    id,
		"output": name,
	}
	if alias != "" {
		tags["alias"] = alias
	}
	bs := NewBufferStats(tags, capacity)

	switch strategy {
	case "", "memory":
		buf, err := NewMemoryBuffer(capacity, bs)
		if err != nil {
			return nil, err
		}
		buf.dropNewest = options.dropStrategy == DropNewest
		return buf, nil
	case "disk_write_through":
		return NewDiskBuffer(id, path, bs, diskSync)
	case "discard":
		return newDiscardBuffer(bs), nil
	}
	return nil, fmt.Errorf("invalid buffer strategy %q", strategy)
}

// CheckBufferSettings verifies that the buffer settings are valid without
// opening or allocating the buffer.
func CheckBufferSettings(strategy string) error {
	switch strategy {
	case "", "memory", "disk_write_through":
		return nil
	}
	return fmt.Errorf("invalid buffer strategy %q", strategy)
}

func NewBufferStats(tags map[string]string, capacity int) BufferStats {
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
		MetricsRejected: selfstat.Register(
			"write",
			"metrics_rejected",
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

func (b *BufferStats) metricAdded(count int64) {
	b.MetricsAdded.Incr(count)
}

func (b *BufferStats) metricWritten(m telegraf.Metric) {
	AgentMetricsWritten.Incr(1)
	b.MetricsWritten.Incr(1)
	m.Accept()
}

func (b *BufferStats) metricRejected(m telegraf.Metric) {
	AgentMetricsRejected.Incr(1)
	b.MetricsRejected.Incr(1)
	m.Reject()
}

func (b *BufferStats) metricDropped(m telegraf.Metric) {
	AgentMetricsDropped.Incr(1)
	b.MetricsDropped.Incr(1)
	m.Reject()
}
