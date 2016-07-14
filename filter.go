package telegraf

type Filter interface {
	// SampleConfig returns the default configuration of the Input
	SampleConfig() string

	// Description returns a one-sentence description on the Input
	Description() string

	//create pipe for filter
	Pipe(in chan Metric) chan Metric

	// start the filter
	Start(shutdown chan struct{})
}
