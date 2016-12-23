package processors

import "github.com/influxdata/telegraf/plugins"

type Creator func() plugins.Processor

var Processors = map[string]Creator{}

func Add(name string, creator Creator) {
	Processors[name] = creator
}
