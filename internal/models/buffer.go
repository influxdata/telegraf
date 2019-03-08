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
	batchSize  int // number of metrics currently in the batch

	MetricsAdded   selfstat.Stat
	MetricsWritten selfstat.Stat
	MetricsDropped selfstat.Stat
	BufferSize     selfstat.Stat
	BufferLimit    selfstat.Stat
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
		BufferSize: selfstat.Register(
			"write",
			"buffer_size",
			map[string]string{"output": name},
		),
		BufferLimit: selfstat.Register(
			"write",
			"buffer_limit",
			map[string]string{"output": name},
		),
	}
	b.BufferSize.Set(int64(0))
	b.BufferLimit.Set(int64(capacity))
	return b
}

// Len returns the number of metrics currently in the buffer.
func (b *Buffer) Len() int {
	b.Lock()
	defer b.Unlock()

	return b.length()
}

func (b *Buffer) length() int {
	return min(b.size+b.batchSize, b.cap)
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

func (b *Buffer) add(m telegraf.Metric) {
	// Check if Buffer is full
	if b.size == b.cap {
		b.metricDropped(b.buf[b.last])

		if b.last == b.batchFirst && b.batchSize > 0 {
			b.batchSize--
			b.batchFirst = b.next(b.batchFirst)
		}
	}

	b.metricAdded()

	b.buf[b.last] = m
	b.last = b.next(b.last)

	if b.size == b.cap {
		b.first = b.next(b.first)
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

	b.BufferSize.Set(int64(b.length()))
}

// Batch returns a slice containing up to batchSize of the most recently added
// metrics.  Metrics are ordered from newest to oldest in the batch.  The
// batch must not be modified by the client.
func (b *Buffer) Batch(batchSize int) []telegraf.Metric {
	b.Lock()
	defer b.Unlock()

	outLen := min(b.size, batchSize)
	out := make([]telegraf.Metric, outLen)
	if outLen == 0 {
		return out
	}

	b.batchFirst = b.cap + b.last - outLen
	b.batchFirst %= b.cap
	b.batchSize = outLen

	batchIndex := b.batchFirst
	for i := range out {
		out[len(out)-1-i] = b.buf[batchIndex]
		b.buf[batchIndex] = nil
		batchIndex = b.next(batchIndex)
	}

	b.last = b.batchFirst
	b.size -= outLen
	return out
}

// Accept marks the batch, acquired from Batch(), as successfully written.
func (b *Buffer) Accept(batch []telegraf.Metric) {
	b.Lock()
	defer b.Unlock()

	for _, m := range batch {
		b.metricWritten(m)
	}

	b.resetBatch()
	b.BufferSize.Set(int64(b.length()))
}

// Reject returns the batch, acquired from Batch(), to the buffer and marks it
// as unsent.
func (b *Buffer) Reject(batch []telegraf.Metric) {
	b.Lock()
	defer b.Unlock()

	if len(batch) == 0 {
		return
	}

	older := b.dist(b.first, b.batchFirst)
	free := b.cap - b.size
	restore := min(len(batch), free+older)

	// Rotate newer metrics forward the number of metrics that we can restore.
	rb := b.batchFirst
	rp := b.last
	re := b.nextby(rp, restore)
	b.last = re

	for rb != rp && rp != re {
		rp = b.prev(rp)
		re = b.prev(re)

		if b.buf[re] != nil {
			b.metricDropped(b.buf[re])
			b.first = b.next(b.first)
		}

		b.buf[re] = b.buf[rp]
		b.buf[rp] = nil
	}

	// Copy metrics from the batch back into the buffer; recall that the
	// batch is in reverse order compared to b.buf
	for i := range batch {
		if i < restore {
			re = b.prev(re)
			b.buf[re] = batch[i]
			b.size = min(b.size+1, b.cap)
		} else {
			b.metricDropped(batch[i])
		}
	}

	b.resetBatch()
	b.BufferSize.Set(int64(b.length()))
}

// dist returns the distance between two indexes.  Because this data structure
// uses a half open range the arguments must both either left side or right
// side pairs.
func (b *Buffer) dist(begin, end int) int {
	if begin <= end {
		return end - begin
	} else {
		return b.cap - begin + end
	}
}

// next returns the next index with wrapping.
func (b *Buffer) next(index int) int {
	index++
	if index == b.cap {
		return 0
	}
	return index
}

// next returns the index that is count newer with wrapping.
func (b *Buffer) nextby(index, count int) int {
	index += count
	index %= b.cap
	return index
}

// next returns the prev index with wrapping.
func (b *Buffer) prev(index int) int {
	index--
	if index < 0 {
		return b.cap - 1
	}
	return index
}

func (b *Buffer) resetBatch() {
	b.batchFirst = 0
	b.batchSize = 0
}

func min(a, b int) int {
	if b < a {
		return b
	}
	return a
}
