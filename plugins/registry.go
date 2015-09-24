package plugins

import "time"

type Accumulator interface {
	// Create a point with a value, decorating it with tags
	// NOTE: tags is expected to be owned by the caller, don't mutate
	// it after passing to Add.
	Add(measurement string, value interface{}, tags map[string]string)

	// Create a point with a set of values, decorating it with tags
	// NOTE: tags and values are expected to be owned by the caller, don't mutate
	// them after passing to AddFieldsWithTime.
	AddFieldsWithTime(
		measurement string,
		values map[string]interface{},
		tags map[string]string,
		timestamp time.Time,
	)
}

type Plugin interface {
	// SampleConfig returns the default configuration of the Plugin
	SampleConfig() string

	// Description returns a one-sentence description on the Plugin
	Description() string

	// Gather takes in an accumulator and adds the metrics that the Plugin
	// gathers. This is called every "interval"
	Gather(Accumulator) error
}

type ServicePlugin interface {
	// SampleConfig returns the default configuration of the Plugin
	SampleConfig() string

	// Description returns a one-sentence description on the Plugin
	Description() string

	// Gather takes in an accumulator and adds the metrics that the Plugin
	// gathers. This is called every "interval"
	Gather(Accumulator) error

	// Start starts the ServicePlugin's service, whatever that may be
	Start() error

	// Stop stops the services and closes any necessary channels and connections
	Stop()
}

type Creator func() Plugin

var Plugins = map[string]Creator{}

func Add(name string, creator Creator) {
	Plugins[name] = creator
}
