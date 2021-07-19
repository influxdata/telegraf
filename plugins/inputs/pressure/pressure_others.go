// +build !linux

package pressure

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func (p *Pressure) Init() error {
	p.Log.Warn("Only Linux platform is supported")
	return nil
}

func (p *Pressure) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("pressure", func() telegraf.Input {
		return &Pressure{}
	})
}
