package inputs

import (
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/registry"
)

type Creator func() telegraf.Input

var Inputs = map[string]Creator{}

func Add(name string, creator Creator) {
	if override := registry.GetName(); override != "" {
		name = override
	}
	log.Println("D! Loading plugin: [[inputs." + name + "]]")
	Inputs[name] = creator
}
