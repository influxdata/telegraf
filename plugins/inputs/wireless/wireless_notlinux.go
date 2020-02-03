// +build !linux

package wireless

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func (w *Wireless) Init() error {
	w.Log.Warn("Current platform is not supported")
	return nil
}

func (w *Wireless) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("wireless", func() telegraf.Input {
		return &Wireless{}
	})
}
