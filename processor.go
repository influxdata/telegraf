package telegraf

type Processor interface {
	// SampleConfig returns the default configuration of the Input
	SampleConfig() string

	// Description returns a one-sentence description on the Input
	Description() string

	// Apply the filter to the given metric.
	Apply(in ...Metric) []Metric
}
