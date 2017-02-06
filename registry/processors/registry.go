package processors

import (
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/registry"
)

type Creator func() telegraf.Processor

var Processors = map[string]Creator{}

func Add(name string, creator Creator) {
	if override := registry.GetName(); override != "" {
		name = override
	}
	log.Println("D! Loading plugin: [[processors." + name + "]]")
	Processors[name] = creator
}
