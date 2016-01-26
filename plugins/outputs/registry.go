package outputs

import (
	"github.com/influxdata/telegraf/models"
)

type Creator func() models.Output

var Outputs = map[string]Creator{}

func Add(name string, creator Creator) {
	Outputs[name] = creator
}
