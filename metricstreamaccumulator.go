package telegraf

// MetricStreamAccumulator provides a way to pass processed metrics back to
// the stream, and to add new ones.
// if you don't plan to return a metric to the stream, you should call DropMetric
// so that any trackers know not to wait for it.
// (presumably after you've changed it)
type MetricStreamAccumulator interface {
	// PassMetric adds an metric to the accumulator, passing it downstream.
	PassMetric(Metric)

	// DropMetric indicates that you do not want this metric to continue being processed
	DropMetric(Metric)
}
