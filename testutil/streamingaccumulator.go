package testutil

import "github.com/influxdata/telegraf"

type testMetricStream struct {
	queue                     []telegraf.Metric
	ProcessedMetrics          []telegraf.Metric
	closeStreamWhenQueueEmpty bool
}

func NewTestMetricStream(closeOnEmpty bool) *testMetricStream {
	return &testMetricStream{
		queue:                     []telegraf.Metric{},
		closeStreamWhenQueueEmpty: closeOnEmpty,
	}
}

func (a *testMetricStream) Enqueue(m ...telegraf.Metric) {
	a.queue = append(a.queue, m...)
}

func (a *testMetricStream) PassMetric(m telegraf.Metric) {
	a.ProcessedMetrics = append(a.ProcessedMetrics, m)
}

func (a *testMetricStream) GetNextMetric() telegraf.Metric {
	if !a.IsMetricAvailable() {
		return nil
	}
	m := a.queue[0]
	a.queue = a.queue[1:]
	return m
}

func (a *testMetricStream) IsMetricAvailable() bool {
	return len(a.queue) > 0
}

func (a *testMetricStream) IsStreamClosed() bool {
	// approximate queue closure here if we need to.
	if a.closeStreamWhenQueueEmpty {
		return len(a.queue) == 0
	}
	return false
}
