package aggregators

import (
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/registry"
)

type Creator func() telegraf.Aggregator

var Aggregators = map[string]Creator{}

func Add(name string, creator Creator) {
	if override := registry.GetName(); override != "" {
		name = override
	}
	log.Println("D! Loading plugin: [[aggregators." + name + "]]")
	Aggregators[name] = creator
}
