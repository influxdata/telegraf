package telegraf

type Input interface {
	// SampleConfig returns the default configuration of the Input
	SampleConfig() string

	// Description returns a one-sentence description on the Input
	Description() string

	// Gather takes in an accumulator and adds the metrics that the Input
	// gathers. This is called every "interval"
	Gather(Accumulator) error
}

type ServiceInput interface {
	Input

	// Start starts the ServiceInput's service, whatever that may be
	Start(Accumulator) error

	// Stop stops the services and closes any necessary channels and connections
	Stop()
}
