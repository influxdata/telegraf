package plugins

import "time"

type Accumulator interface {
	// Create a named point with a value, decorating it with named tags
	// NOTE: tags is expected to be owned by the caller, don't mutate
	// it after passing to Add.
	Add(name string, value interface{}, tags map[string]string)

	// Create a named point with a set of values, decorating it with named tags
	// NOTE: tags and values are expected to be owned by the caller, don't mutate
	// them after passing to AddValuesWithTime.
	AddValuesWithTime(
		name string,
		values map[string]interface{},
		tags map[string]string,
		timestamp time.Time,
	)
}

type Plugin interface {
	SampleConfig() string
	Description() string
	Gather(Accumulator) error
}

type Creator func() Plugin

var Plugins = map[string]Creator{}

func Add(name string, creator Creator) {
	Plugins[name] = creator
}
