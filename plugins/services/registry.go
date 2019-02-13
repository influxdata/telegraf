package services

import (
	"fmt"

	"github.com/influxdata/telegraf"
)

type Creator func() telegraf.Service

var Services = map[string]Creator{}

func Add(name string, creator Creator) {
	fmt.Println(name, "start")
	Services[name] = creator
	fmt.Println(name, "finish")
}
