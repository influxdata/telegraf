package buffer

import (
	"sync"

	"github.com/influxdata/telegraf"
)

// Buffer is an object for storing metrics in a circular buffer.
type Buffer struct {
	buf chan telegraf.Metric
	// total dropped metrics
	drops int
	// total metrics added
	total int

	sync.Mutex
}

// NewBuffer returns a Buffer
//   size is the maximum number of metrics that Buffer will cache. If Add is
//   called when the buffer is full, then the oldest metric(s) will be dropped.
func NewBuffer(size int) *Buffer {
	return &Buffer{
		buf: make(chan telegraf.Metric, size),
	}
}

// IsEmpty returns true if Buffer is empty.
func (b *Buffer) IsEmpty() bool {
	return len(b.buf) == 0
}

// Len returns the current length of the buffer.
func (b *Buffer) Len() int {
	return len(b.buf)
}

// Drops returns the total number of dropped metrics that have occured in this
// buffer since instantiation.
func (b *Buffer) Drops() int {
	return b.drops
}

// Total returns the total number of metrics that have been added to this buffer.
func (b *Buffer) Total() int {
	return b.total
}

// Add adds metrics to the buffer.
func (b *Buffer) Add(metrics ...telegraf.Metric) {
	for i, _ := range metrics {
		b.total++
		select {
		case b.buf <- metrics[i]:
		default:
			b.drops++
			<-b.buf
			b.buf <- metrics[i]
		}
	}
}

// Batch returns a batch of metrics of size batchSize.
// the batch will be of maximum length batchSize. It can be less than batchSize,
// if the length of Buffer is less than batchSize.
func (b *Buffer) Batch(batchSize int) []telegraf.Metric {
	b.Lock()
	n := min(len(b.buf), batchSize)
	out := make([]telegraf.Metric, n)
	for i := 0; i < n; i++ {
		out[i] = <-b.buf
	}
	b.Unlock()
	return out
}

func min(a, b int) int {
	if b < a {
		return b
	}
	return a
}
