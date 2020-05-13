package testutil

import "github.com/influxdata/telegraf"

type testStreamingAccumulator struct {
	queue                     []telegraf.Metric
	ProcessedMetrics          []telegraf.Metric
	closeStreamWhenQueueEmpty bool
}

func NewTestStreamingAccumulator(closeOnEmpty bool) *testStreamingAccumulator {
	return &testStreamingAccumulator{
		queue:                     []telegraf.Metric{},
		closeStreamWhenQueueEmpty: closeOnEmpty,
	}
}

func (a *testStreamingAccumulator) Enqueue(m ...telegraf.Metric) {
	a.queue = append(a.queue, m...)
}

func (a *testStreamingAccumulator) PassMetric(m telegraf.Metric) {
	a.ProcessedMetrics = append(a.ProcessedMetrics, m)
}

func (a *testStreamingAccumulator) GetNextMetric() telegraf.Metric {
	if !a.IsMetricAvailable() {
		return nil
	}
	m := a.queue[0]
	a.queue = a.queue[1:]
	return m
}

func (a *testStreamingAccumulator) IsMetricAvailable() bool {
	return len(a.queue) > 0
}

func (a *testStreamingAccumulator) IsStreamClosed() bool {
	// approximate queue closure here if we need to.
	if a.closeStreamWhenQueueEmpty {
		return len(a.queue) == 0
	}
	return false
}
