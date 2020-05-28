package telegraf

// MetricStreamAccumulator provides a way to pass processed metrics back to
// the stream, and to add new ones.
type MetricStreamAccumulator interface {
	// PassMetric adds an metric to the accumulator, passing it downstream.
	PassMetric(Metric)
}
