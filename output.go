package telegraf

type Output interface {
	// Connect to the Output
	Connect() error
	// Close any connections to the Output
	Close() error
	// Description returns a one-sentence description on the Output
	Description() string
	// SampleConfig returns the default configuration of the Output
	SampleConfig() string
	// Write takes in group of points to be written to the Output
	Write(metrics []Metric) error
}

// AggregatingOutput adds aggregating functionality to an Output.  May be used
// if the Output only accepts a fixed set of aggregations over a time period.
// These functions may be called concurrently to the Write function.
type AggregatingOutput interface {
	// Connect to the Output
	Connect() error
	// Close any connections to the Output
	Close() error
	// Description returns a one-sentence description on the Output
	Description() string
	// SampleConfig returns the default configuration of the Output
	SampleConfig() string
	// Write takes in group of points to be written to the Output
	Write(metrics []Metric) error

	// Add the metric to the aggregator
	Add(in Metric)
	// Push returns the aggregated metrics and is called every flush interval.
	Push() []Metric
	// Reset signals the the aggregator period is completed.
	Reset()
}
