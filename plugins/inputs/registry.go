package inputs

import "github.com/influxdata/telegraf/plugins"

type Creator func() plugins.Input

var Inputs = map[string]Creator{}

func Add(name string, creator Creator) {
	Inputs[name] = creator
}
