package outputs

import (
	"github.com/influxdata/telegraf/plugins"
)

type Creator func() plugins.Output

var Outputs = map[string]Creator{}

func Add(name string, creator Creator) {
	Outputs[name] = creator
}
