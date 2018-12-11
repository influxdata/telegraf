package models

import (
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
)

var (
	AgentMetricsWritten = selfstat.Register("agent", "metrics_written", map[string]string{})
	AgentMetricsDropped = selfstat.Register("agent", "metrics_dropped", map[string]string{})
)

// Buffer stores metrics in a circular buffer.
type Buffer struct {
	sync.Mutex
	buf   []telegraf.Metric
	first int // index of the first/oldest metric
	last  int // one after the index of the last/newest metric
	size  int // number of metrics currently in the buffer
	cap   int // the capacity of the buffer

	batchFirst int // index of the first metric in the batch
	batchLast  int // one after the index of the last metric in the batch
	batchSize  int // number of metrics current in the batch

	MetricsAdded   selfstat.Stat
	MetricsWritten selfstat.Stat
	MetricsDropped selfstat.Stat
}

// NewBuffer returns a new empty Buffer with the given capacity.
func NewBuffer(name string, capacity int) *Buffer {
	b := &Buffer{
		buf:   make([]telegraf.Metric, capacity),
		first: 0,
		last:  0,
		size:  0,
		cap:   capacity,

		MetricsAdded: selfstat.Register(
			"write",
			"metrics_added",
			map[string]string{"output": name},
		),
		MetricsWritten: selfstat.Register(
			"write",
			"metrics_written",
			map[string]string{"output": name},
		),
		MetricsDropped: selfstat.Register(
			"write",
			"metrics_dropped",
			map[string]string{"output": name},
		),
	}
	return b
}

// Len returns the number of metrics currently in the buffer.
func (b *Buffer) Len() int {
	b.Lock()
	defer b.Unlock()

	return b.size
}

func (b *Buffer) metricAdded() {
	b.MetricsAdded.Incr(1)
}

func (b *Buffer) metricWritten(metric telegraf.Metric) {
	AgentMetricsWritten.Incr(1)
	b.MetricsWritten.Incr(1)
	metric.Accept()
}

func (b *Buffer) metricDropped(metric telegraf.Metric) {
	AgentMetricsDropped.Incr(1)
	b.MetricsDropped.Incr(1)
	metric.Reject()
}

func (b *Buffer) inBatch() bool {
	if b.batchSize == 0 {
		return false
	}

	if b.batchFirst < b.batchLast {
		return b.last >= b.batchFirst && b.last < b.batchLast
	} else {
		return b.last >= b.batchFirst || b.last < b.batchLast
	}
}

func (b *Buffer) add(m telegraf.Metric) {
	// Check if Buffer is full
	if b.size == b.cap {
		if b.batchSize == 0 {
			// No batch taken by the output, we can drop the metric now.
			b.metricDropped(b.buf[b.last])
		} else if b.inBatch() {
			// There is an outstanding batch and this will overwrite a metric
			// in it, delay the dropping only in case the batch gets rejected.
			b.batchSize--
			b.batchFirst++
			b.batchFirst %= b.cap
		} else {
			// There is an outstanding batch, but this overwrites a metric
			// outside of it.
			b.metricDropped(b.buf[b.last])
		}
	}

	b.metricAdded()

	b.buf[b.last] = m
	b.last++
	b.last %= b.cap

	if b.size == b.cap {
		b.first++
		b.first %= b.cap
	}

	b.size = min(b.size+1, b.cap)
}

// Add adds metrics to the buffer
func (b *Buffer) Add(metrics ...telegraf.Metric) {
	b.Lock()
	defer b.Unlock()

	for i := range metrics {
		b.add(metrics[i])
	}
}

// Batch returns a slice containing up to batchSize of the most recently added
// metrics.
//
// The metrics contained in the batch are not removed from the buffer, instead
// the last batch is recorded and removed only if Accept is called.
func (b *Buffer) Batch(batchSize int) []telegraf.Metric {
	b.Lock()
	defer b.Unlock()

	outLen := min(b.size, batchSize)
	out := make([]telegraf.Metric, outLen)
	if outLen == 0 {
		return out
	}

	b.batchFirst = b.first
	b.batchLast = b.first + outLen
	b.batchLast %= b.cap
	b.batchSize = outLen

	until := min(b.cap, b.first+outLen)

	n := copy(out, b.buf[b.first:until])
	if n < outLen {
		copy(out[n:], b.buf[:outLen-n])
	}
	return out
}

// Accept removes the metrics contained in the last batch.
func (b *Buffer) Accept(batch []telegraf.Metric) {
	b.Lock()
	defer b.Unlock()

	for _, m := range batch {
		b.metricWritten(m)
	}

	b.size -= b.batchSize
	for i := 0; i < b.batchSize; i++ {
		b.buf[b.first] = nil
		b.first++
		b.first %= b.cap
	}

	b.resetBatch()
}

// Reject clears the current batch record so that calls to Accept will have no
// effect.
func (b *Buffer) Reject(batch []telegraf.Metric) {
	b.Lock()
	defer b.Unlock()

	if len(batch) > b.batchSize {
		// Part or all of the batch was dropped before reject was called.
		for _, m := range batch[b.batchSize:] {
			b.metricDropped(m)
		}
	}

	b.resetBatch()
}

func (b *Buffer) resetBatch() {
	b.batchFirst = 0
	b.batchLast = 0
	b.batchSize = 0
}

func min(a, b int) int {
	if b < a {
		return b
	}
	return a
}
