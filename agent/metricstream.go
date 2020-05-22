package agent

import (
	"github.com/influxdata/telegraf"
)

// metricStream implements the MetricStreamAccumulator interface
type metricStream struct {
	out chan<- telegraf.Metric
}

// NewMetricStream creates a new streaming accumulator that is safe to
// use across multiple goroutines. It conceptually wraps a incoming and outgoing
// channels so that the user doesn't have to deal with them directly.
// It also handles metric filtering and modifications so that plugins don't need
// to deal with it.
func NewMetricStreamAccumulator(
	outMetrics chan<- telegraf.Metric,
) telegraf.MetricStreamAccumulator {
	return &metricStream{
		out: outMetrics,
	}
}

func (sa *metricStream) PassMetric(m telegraf.Metric) {
	sa.out <- m
}

func (sa *metricStream) DropMetric(m telegraf.Metric) {

}
