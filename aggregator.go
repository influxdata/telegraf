package telegraf

type Aggregator interface {
	// SampleConfig returns the default configuration of the Input
	SampleConfig() string

	// Description returns a one-sentence description on the Input
	Description() string

	// Apply the metric to the aggregator
	Apply(in Metric)

	// Start starts the service filter with the given accumulator
	Start(acc Accumulator) error
	Stop()
}
