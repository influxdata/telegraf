package inputs

import (
	"fmt"

	"github.com/influxdata/telegraf"
)

type Creator func() telegraf.Input

var Inputs = map[string]Creator{}

func Add(name string, creator Creator) {
	fmt.Println(name, "start")
	Inputs[name] = creator
	fmt.Println(name, "finish")
}
