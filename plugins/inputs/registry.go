package inputs

import "time"

type Accumulator interface {
	// Create a point with a value, decorating it with tags
	// NOTE: tags is expected to be owned by the caller, don't mutate
	// it after passing to Add.
	Add(measurement string,
		value interface{},
		tags map[string]string,
		t ...time.Time)

	AddFields(measurement string,
		fields map[string]interface{},
		tags map[string]string,
		t ...time.Time)
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
	// SampleConfig returns the default configuration of the Input
	SampleConfig() string

	// Description returns a one-sentence description on the Input
	Description() string

	// Gather takes in an accumulator and adds the metrics that the Input
	// gathers. This is called every "interval"
	Gather(Accumulator) error

	// Start starts the ServiceInput's service, whatever that may be
	Start() error

	// Stop stops the services and closes any necessary channels and connections
	Stop()
}

type Creator func() Input

var Inputs = map[string]Creator{}

func Add(name string, creator Creator) {
	Inputs[name] = creator
}
