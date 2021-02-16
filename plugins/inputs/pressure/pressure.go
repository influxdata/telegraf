package pressure

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Pressure struct {
	Log telegraf.Logger `toml:"-"`
}

func (_ *Pressure) SampleConfig() string { return "" }

func (_ *Pressure) Description() string {
	return "Gather system metrics about Pressure Stall (PSI)"
}

func init() {
	inputs.Add("pressure", func() telegraf.Input {
		return &Pressure{}
	})
}

