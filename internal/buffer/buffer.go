package buffer

import (
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
)

var (
	MetricsWritten = selfstat.Register("agent", "metrics_written", map[string]string{})
	MetricsDropped = selfstat.Register("agent", "metrics_dropped", map[string]string{})
)

// Buffer is an object for storing metrics in a circular buffer.
type Buffer struct {
	sync.Mutex
	buf   []telegraf.Metric
	first int
	last  int
	size  int
	empty bool
}

// NewBuffer returns a Buffer
//   size is the maximum number of metrics that Buffer will cache. If Add is
//   called when the buffer is full, then the oldest metric(s) will be dropped.
func NewBuffer(size int) *Buffer {
	return &Buffer{
		buf:   make([]telegraf.Metric, size),
		first: 0,
		last:  0,
		size:  size,
		empty: true,
	}
}

// Len returns the current length of the buffer.
func (b *Buffer) Len() int {
	b.Lock()
	defer b.Unlock()
	return b.length()
}

func (b *Buffer) length() int {
	if b.empty {
		return 0
	} else if b.first <= b.last {
		return b.last - b.first + 1
	}
	// Spans the end of array.
	// size - gap in the middle
	return b.size - (b.first - b.last - 1) // size - gap
}

func (b *Buffer) push(m telegraf.Metric) {
	// Empty
	if b.empty {
		b.last = b.first // Reset
		b.buf[b.last] = m
		b.empty = false
		return
	}

	b.last++
	b.last %= b.size

	// Full
	if b.first == b.last {
		MetricsDropped.Incr(1)
		b.first++
		b.first %= b.size
	}
	b.buf[b.last] = m
}

// Add adds metrics to the buffer.
func (b *Buffer) Add(metrics ...telegraf.Metric) {
	b.Lock()
	defer b.Unlock()
	for i := range metrics {
		MetricsWritten.Incr(1)
		b.push(metrics[i])
	}
}

// Batch returns a batch of metrics of size batchSize.
// the batch will be of maximum length batchSize. It can be less than batchSize,
// if the length of Buffer is less than batchSize.
func (b *Buffer) Batch(batchSize int) []telegraf.Metric {
	b.Lock()
	defer b.Unlock()
	outLen := min(b.length(), batchSize)
	out := make([]telegraf.Metric, outLen)
	if outLen == 0 {
		return out
	}

	// We copy everything right of first up to last, count or end
	// b.last >= rightInd || b.last < b.first
	// therefore wont copy past b.last
	rightInd := min(b.size, b.first+outLen) - 1

	copyCount := copy(out, b.buf[b.first:rightInd+1])

	// We've emptied the ring
	if rightInd == b.last {
		b.empty = true
	}
	b.first = rightInd + 1
	b.first %= b.size

	// We circle back for the rest
	if copyCount < outLen {
		right := min(b.last, outLen-copyCount)
		copy(out[copyCount:], b.buf[b.first:right+1])
		// We've emptied the ring
		if right == b.last {
			b.empty = true
		}
		b.first = right + 1
		b.first %= b.size
	}
	return out
}

func min(a, b int) int {
	if b < a {
		return b
	}
	return a
}
