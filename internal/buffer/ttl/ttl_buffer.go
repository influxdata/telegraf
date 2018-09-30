package ttl

import (
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
)

var now = time.Now

var (
	MetricsWritten = selfstat.Register("agent", "metrics_written", map[string]string{})
	MetricsDropped = selfstat.Register("agent", "metrics_dropped", map[string]string{})
)

type TTLBufferItem struct {
	metric    *telegraf.Metric
	timestamp time.Time
	next      *TTLBufferItem
}

func (i *TTLBufferItem) isExpired(deadline time.Time) bool {
	return i.timestamp.Before(deadline)
}

type TTLBuffer struct {
	sync.Mutex
	head *TTLBufferItem
	tail *TTLBufferItem
	ttl  time.Duration
	size int
}

func NewTTLBuffer(ttl time.Duration) *TTLBuffer {
	return &TTLBuffer{
		head: nil,
		tail: nil,
		ttl:  ttl,
		size: 0,
	}
}

func (b *TTLBuffer) IsEmpty() bool {
	return b.Len() == 0
}

func (b *TTLBuffer) Len() int {
	return b.size
}

func (b *TTLBuffer) push(m *telegraf.Metric, timestamp time.Time) {
	item := &TTLBufferItem{
		metric:    m,
		timestamp: timestamp,
	}

	if b.IsEmpty() {
		b.head, b.tail = item, item
	} else {
		prevHead := b.head
		b.head = item
		prevHead.next = b.head
	}

	b.size++
}

func (b *TTLBuffer) pop() *telegraf.Metric {
	popped := b.tail
	b.tail = b.tail.next
	b.size--
	return popped.metric
}

func (b *TTLBuffer) dropExpired() {
	if b.IsEmpty() {
		return
	}

	deadline := now().Add(-b.ttl)
	for b.tail.isExpired(deadline) {
		b.pop()
		MetricsDropped.Incr(1)
	}
}

func (b *TTLBuffer) Add(metrics ...telegraf.Metric) {
	b.Lock()
	defer b.Unlock()

	now := now()
	for _, m := range metrics {
		MetricsWritten.Incr(1)
		b.push(&m, now)
	}

	b.dropExpired()
}

func (b *TTLBuffer) Batch(batchSize int) []telegraf.Metric {
	b.Lock()
	defer b.Unlock()

	outLen := min(b.Len(), batchSize)
	out := make([]telegraf.Metric, outLen)
	if outLen == 0 {
		return out
	}

	for i := 0; i < outLen; i++ {
		out[i] = *b.pop()
	}
	return out
}

func min(a, b int) int {
	if b < a {
		return b
	}
	return a
}
