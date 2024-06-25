package models

import (
	"sync"

	"github.com/influxdata/telegraf"
)

// MemoryBuffer stores metrics in a circular buffer.
type MemoryBuffer struct {
	sync.Mutex
	BufferStats

	buf   []telegraf.Metric
	first int // index of the first/oldest metric
	last  int // one after the index of the last/newest metric
	size  int // number of metrics currently in the buffer
	cap   int // the capacity of the buffer

	batchFirst int // index of the first metric in the batch
	batchSize  int // number of metrics currently in the batch
}

func NewMemoryBuffer(capacity int, stats BufferStats) (*MemoryBuffer, error) {
	return &MemoryBuffer{
		BufferStats: stats,
		buf:         make([]telegraf.Metric, capacity),
		cap:         capacity,
	}, nil
}

func (b *MemoryBuffer) Len() int {
	b.Lock()
	defer b.Unlock()

	return b.length()
}

func (b *MemoryBuffer) length() int {
	return min(b.size+b.batchSize, b.cap)
}

func (b *MemoryBuffer) addMetric(m telegraf.Metric) int {
	dropped := 0
	// Check if Buffer is full
	if b.size == b.cap {
		b.metricDropped(b.buf[b.last])
		dropped++

		if b.batchSize > 0 {
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
	return dropped
}

func (b *MemoryBuffer) Add(metrics ...telegraf.Metric) int {
	b.Lock()
	defer b.Unlock()

	dropped := 0
	for i := range metrics {
		if n := b.addMetric(metrics[i]); n != 0 {
			dropped += n
		}
	}

	b.BufferSize.Set(int64(b.length()))
	return dropped
}

func (b *MemoryBuffer) Batch(batchSize int) []telegraf.Metric {
	b.Lock()
	defer b.Unlock()

	outLen := min(b.size, batchSize)
	out := make([]telegraf.Metric, outLen)
	if outLen == 0 {
		return out
	}

	b.batchFirst = b.first
	b.batchSize = outLen

	batchIndex := b.batchFirst
	for i := range out {
		out[i] = b.buf[batchIndex]
		b.buf[batchIndex] = nil
		batchIndex = b.next(batchIndex)
	}

	b.first = b.nextby(b.first, b.batchSize)
	b.size -= outLen
	return out
}

func (b *MemoryBuffer) Accept(batch []telegraf.Metric) {
	b.Lock()
	defer b.Unlock()

	for _, m := range batch {
		b.metricWritten(m)
	}

	b.resetBatch()
	b.BufferSize.Set(int64(b.length()))
}

func (b *MemoryBuffer) Reject(batch []telegraf.Metric) {
	b.Lock()
	defer b.Unlock()

	if len(batch) == 0 {
		return
	}

	free := b.cap - b.size
	restore := min(len(batch), free)
	skip := len(batch) - restore

	b.first = b.prevby(b.first, restore)
	b.size = min(b.size+restore, b.cap)

	re := b.first

	// Copy metrics from the batch back into the buffer
	for i := range batch {
		if i < skip {
			b.metricDropped(batch[i])
		} else {
			b.buf[re] = batch[i]
			re = b.next(re)
		}
	}

	b.resetBatch()
	b.BufferSize.Set(int64(b.length()))
}

func (b *MemoryBuffer) Stats() BufferStats {
	return b.BufferStats
}

// next returns the next index with wrapping.
func (b *MemoryBuffer) next(index int) int {
	index++
	if index == b.cap {
		return 0
	}
	return index
}

// nextby returns the index that is count newer with wrapping.
func (b *MemoryBuffer) nextby(index, count int) int {
	index += count
	index %= b.cap
	return index
}

// prevby returns the index that is count older with wrapping.
func (b *MemoryBuffer) prevby(index, count int) int {
	index -= count
	for index < 0 {
		index += b.cap
	}

	index %= b.cap
	return index
}

func (b *MemoryBuffer) resetBatch() {
	b.batchFirst = 0
	b.batchSize = 0
}
