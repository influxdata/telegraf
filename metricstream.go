package telegraf

// MetricStream provides a way to interact with the metric stream,
// both being notified of new metrics (GetNextMetric), and being able to pass
// them on downstream (PassMetric).
// (presumably after you've changed it)
type MetricStream interface {
	// PassMetric adds an metric to the accumulator.
	PassMetric(Metric)

	// GetNextMetric returns the next metric availabe to process.
	// If no metric is available yet, it blocks until a metric is ready for processing.
	// See also IsMetricAvailable()
	// GetNextMetric will return nil when the stream is trying to shut down to
	// prevent this call from blocking forever
	GetNextMetric() Metric

	// IsMetricAvailable is false when GetNextMetric() would block.
	// Use this if you need to check if a metric is ready to be processed without
	// blocking.
	IsMetricAvailable() bool

	// IsStreamClosed is true only when Telegraf is shutting down.
	// You can use this to gracefully stop processing on quit and restart.
	IsStreamClosed() bool
}
