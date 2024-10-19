package shim

import "github.com/influxdata/telegraf"

// inputShim implements the MetricMaker interface.
type inputShim struct {
	Input telegraf.Input
}

// LogName satisfies the MetricMaker interface
func (i inputShim) LogName() string {
	return ""
}

// MakeMetric satisfies the MetricMaker interface
func (i inputShim) MakeMetric(m telegraf.Metric) telegraf.Metric {
	return m // don't need to do anything to it.
}

// Log satisfies the MetricMaker interface
func (i inputShim) Log() telegraf.Logger {
	return nil
}
