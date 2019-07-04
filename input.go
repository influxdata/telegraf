package telegraf

// Initializer is an interface that all plugin types: Inputs, Outputs,
// Processors, and Aggregators can optionally implement to initialize the
// plugin.
type Initializer interface {
	// Init performs one time setup of the plugin and returns an error if the
	// configuration is invalid.
	Init() error
}

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

	// Start the ServiceInput.  The Accumulator may be retained and used until
	// Stop returns.
	Start(Accumulator) error

	// Stop stops the services and closes any necessary channels and connections
	Stop()
}
