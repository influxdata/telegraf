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

// Buffer stores metrics in a circular buffer.
type Buffer struct {
	sync.Mutex
	buf   []telegraf.Metric
	first int // index of the metric first added
	last  int // one after the index of the metric added last
	size  int // number of metrics currently in the buffer
	end   int // one after the index of the end of the buffer

	batchFirst int
	batchLast  int
	batchSize  int
	batchDrop  int64
}

// NewBuffer returns a Buffer
//   size is the maximum number of metrics that Buffer will cache. If Add is
//   called when the buffer is full, then the oldest metric(s) will be dropped.
func NewBuffer(capacity int) *Buffer {
	return &Buffer{
		buf:   make([]telegraf.Metric, capacity),
		first: 0,
		last:  0,
		size:  0,
		end:   capacity,
	}
}

// Len returns the number of metrics currently in the buffer.
func (b *Buffer) Len() int {
	b.Lock()
	defer b.Unlock()

	return b.size
}

func (b *Buffer) add(m telegraf.Metric) {
	b.buf[b.last] = m
	b.last++
	b.last %= b.end

	// full; overwrite first metric
	if b.size == b.end {
		b.first++
		b.first %= b.end

		// If there is an outstanding batch, we can't be sure if metrics from
		// it are dropped until it is acked or rejected.
		if b.batchSize == 0 {
			MetricsDropped.Incr(1)
		} else {
			b.batchDrop++
		}
	}

	b.size = min(b.size+1, b.end)
}

// Add adds metrics to the stack and returns the size of the stack.
func (b *Buffer) Add(metrics ...telegraf.Metric) {
	b.Lock()
	defer b.Unlock()

	for i := range metrics {
		MetricsWritten.Incr(1)
		b.add(metrics[i])
	}
}

// Batch returns a slice containing up to batchSize of the most recently added
// metrics.
//
// The metrics contained in the batch are not removed from the buffer, instead
// the last batch is recorded and removed only if Ack is called.  This is
// meant to avoid the requirement to reorder the buffer if an error occurs.
func (b *Buffer) Batch(batchSize int) []telegraf.Metric {
	b.Lock()
	defer b.Unlock()

	b.Reject()

	outLen := min(b.size, batchSize)
	out := make([]telegraf.Metric, outLen)
	if outLen == 0 {
		return out
	}

	b.batchFirst = b.first
	b.batchLast = b.first + outLen
	b.batchLast %= b.end
	b.batchSize = outLen

	until := min(b.end, b.first+outLen)

	n := copy(out, b.buf[b.first:until])
	if n < outLen {
		copy(out[n:], b.buf[:outLen-n])
	}
	return out
}

// Ack removes the metrics contained in the last batch.
func (b *Buffer) Ack() {
	if b.batchSize != 0 {
		b.first = b.batchLast
		b.size -= b.batchSize
		b.resetBatch()
	}
}

// Reject clears the current batch record so that calls to Ack will have no
// effect.
func (b *Buffer) Reject() {
	MetricsDropped.Incr(b.batchDrop)
	b.resetBatch()
}

func (b *Buffer) resetBatch() {
	b.batchFirst = 0
	b.batchLast = 0
	b.batchSize = 0
	b.batchDrop = 0
}

func min(a, b int) int {
	if b < a {
		return b
	}
	return a
}
