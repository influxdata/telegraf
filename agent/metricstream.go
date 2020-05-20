package agent

import (
	"sync"

	"github.com/influxdata/telegraf"
)

// metricStream implements the MetricStream interface
type metricStream struct {
	l sync.Mutex

	// in pipe is closed and there are no more pending messages if this is true.
	inChClosed bool

	cachedMetric telegraf.Metric
	in           <-chan telegraf.Metric
	out          chan<- telegraf.Metric
}

// NewMetricStream creates a new streaming accumulator that is safe to
// use across multiple goroutines. It conceptually wraps a incoming and outgoing
// channels so that the user doesn't have to deal with them directly.
// It also handles metric filtering and modifications so that plugins don't need
// to deal with it.
func NewMetricStream(
	inMetrics <-chan telegraf.Metric,
	outMetrics chan<- telegraf.Metric,
) telegraf.MetricStream {
	return &metricStream{
		in:  inMetrics,
		out: outMetrics,
	}
}

func (sa *metricStream) PassMetric(m telegraf.Metric) {
	sa.out <- m
}

// GetNextMetric returns a metric or blocks until one is available.
func (sa *metricStream) GetNextMetric() telegraf.Metric {
	sa.l.Lock()
	if sa.cachedMetric != nil {
		m := sa.cachedMetric
		sa.cachedMetric = nil
		sa.l.Unlock()
		return m
	}
	sa.l.Unlock()

	m, ok := <-sa.in
	if !ok {
		sa.l.Lock()
		sa.inChClosed = true
		sa.l.Unlock()
		return nil
	}
	return m
}

// IsMetricAvailable returns true if a metric is available to be read.
// it returns false if GetNextMetric would block or if the stream is closed
func (sa *metricStream) IsMetricAvailable() bool {
	sa.l.Lock()
	defer sa.l.Unlock()
	if sa.cachedMetric != nil {
		return true
	}
	if sa.inChClosed {
		return false
	}

	select {
	case m, ok := <-sa.in:
		if !ok {
			sa.inChClosed = true
			return false
		}
		sa.cachedMetric = m
		return true
	default:
		return false
	}
}

// IsStreamClosed returns true when the stream is closed and there are no more
// metrics to read.
func (sa *metricStream) IsStreamClosed() bool {
	sa.l.Lock()
	defer sa.l.Unlock()
	return sa.cachedMetric == nil && sa.inChClosed
}
