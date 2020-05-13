package agent

import (
	"sync"

	"github.com/influxdata/telegraf"
)

// streamingAccumulator implements the StreamingAccumulator interface
type streamingAccumulator struct {
	sync.Locker

	// in pipe is closed and there are no more pending messages if this is true.
	inClosed bool

	cachedMetric telegraf.Metric
	in           <-chan telegraf.Metric
	out          chan<- telegraf.Metric
}

// NewStreamingAccumulator creates a new streaming accumulator that is safe to
// use across multiple goroutines. It conceptually wraps a incoming and outgoing
// channels so that the user doesn't have to deal with them directly.
// It also handles metric filtering and modifications so that plugins don't need
// to deal with it.
func NewStreamingAccumulator(
	inMetrics <-chan telegraf.Metric,
	outMetrics chan<- telegraf.Metric,
) telegraf.StreamingAccumulator {
	return &streamingAccumulator{
		in:  inMetrics,
		out: outMetrics,
	}
}

func (sa *streamingAccumulator) PassMetric(m telegraf.Metric) {
	sa.out <- m
}

// GetNextMetric returns a metric or blocks until one is available.
func (sa *streamingAccumulator) GetNextMetric() telegraf.Metric {
	sa.Lock()
	if sa.cachedMetric != nil {
		m := sa.cachedMetric
		sa.cachedMetric = nil
		sa.Unlock()
		return m
	}
	sa.Unlock()

	m, ok := <-sa.in
	if !ok {
		sa.Lock()
		sa.inClosed = true
		sa.Unlock()
		return nil
	}
	return m
}

// IsMetricAvailable returns true if a metric is available to be read.
// it returns false if GetNextMetric would block or if the stream is closed
func (sa *streamingAccumulator) IsMetricAvailable() bool {
	sa.Lock()
	defer sa.Unlock()
	if sa.cachedMetric != nil {
		return true
	}
	if sa.inClosed {
		return false
	}

	select {
	case m, ok := <-sa.in:
		if !ok {
			sa.inClosed = true
		}
		sa.cachedMetric = m
		return true
	default:
		return false
	}
}

// IsStreamClosed returns true when the stream is closed and there are no more
// metrics to read.
func (sa *streamingAccumulator) IsStreamClosed() bool {
	sa.Lock()
	defer sa.Unlock()
	return sa.cachedMetric == nil && sa.inClosed
}
