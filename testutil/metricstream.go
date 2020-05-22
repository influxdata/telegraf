package testutil

import "github.com/influxdata/telegraf"

type testMetricStream struct {
	ProcessedMetrics []telegraf.Metric
}

func NewTestMetricStreamAccumulator() *testMetricStream {
	return &testMetricStream{}
}

func (a *testMetricStream) PassMetric(m telegraf.Metric) {
	a.ProcessedMetrics = append(a.ProcessedMetrics, m)
}

func (a *testMetricStream) DropMetric(m telegraf.Metric) {
}
