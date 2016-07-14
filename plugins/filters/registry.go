package filters

import "github.com/influxdata/telegraf"

type Creator func() telegraf.Filter

var Filters = map[string]Creator{}

func Add(name string, creator Creator) {
	Filters[name] = creator
}
