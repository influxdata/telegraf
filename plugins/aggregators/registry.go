package aggregators

import "github.com/influxdata/telegraf/plugins"

type Creator func() plugins.Aggregator

var Aggregators = map[string]Creator{}

func Add(name string, creator Creator) {
	Aggregators[name] = creator
}
