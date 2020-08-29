package telegraf

// Processor is a processor plugin interface for defining new inline processors.
// these are extremely efficient and should be used over StreamingProcessor if
// you do not need asynchronous metric writes.
type Processor interface {
	PluginDescriber

	// Apply the filter to the given metric.
	Apply(in ...Metric) []Metric
}

// StreamingProcessor is a processor that can take in a stream of messages
type StreamingProcessor interface {
	PluginDescriber

	// Start is the initializer for the processor
	// Start is only called once per plugin instance, and never in parallel.
	// Start should exit immediately after setup
	Start(acc Accumulator) error

	// Add is called for each metric to be processed.
	Add(metric Metric, acc Accumulator) error

	// Stop gives you a callback to free resources.
	// by the time Stop is called, the input stream will have already been closed
	// and Add will not be called anymore.
	// When stop returns, you should no longer be writing metrics to the
	// accumulator.
	Stop() error
}
