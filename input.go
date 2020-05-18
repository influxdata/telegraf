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

	// Start the ServiceInput.  The Accumulator may be retained and used until
	// Stop returns.
	Start(Accumulator) error

	// Stop stops the services and closes any necessary channels and connections
	Stop()
}

// StatefulInput allow a ServiceInput to store any arbitrary information that can be restored
// in the event that Telegraf or the plugin is restarted
type StatefulInput interface {
	ServiceInput

	// Sync writes the state of the plugin to the StateStore
	Sync() interface{}

	// Load
	Load(interface{}) error
}
