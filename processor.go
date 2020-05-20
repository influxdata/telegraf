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

	// Start is called when the processor should start.
	// The MetricStream may be retained and used until Stop returns.
	// Start is only called once per plugin instance, and never in parallel.
	// Start should exit when acc.IsStreamClosed() returns true.
	// Start should not exit until the processor is ready to quit and the stream
	// is empty.
	Start(acc MetricStream) error

	// Stop is called when the plugin should stop processing.
	// at this point no new metrics will be coming in to the MetricStream,
	// you can finish up processing the remaining metrics until IsStreamClosed()
	// returns true. Wait for this to happen, then return from Stop. After Stop()
	// returns, the reference to the MetricStream should not be used.
	Stop()
}
