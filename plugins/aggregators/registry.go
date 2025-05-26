package aggregators

import "github.com/influxdata/telegraf"

// Creator is the function to create a new aggregator
type Creator func() telegraf.Aggregator

// Aggregators contains the registry of all known aggregators
var Aggregators = make(map[string]Creator)

// Add adds an aggregator to the registry. Usually this function is called in the plugin's init function
func Add(name string, creator Creator) {
	Aggregators[name] = creator
}
