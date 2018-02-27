package aggregators

import "github.com/influxdata/telegraf"

type Creator func() telegraf.Aggregator

var Aggregators = map[string]Creator{}

func Add(name string, creator Creator) {
	Aggregators[name] = creator
}
