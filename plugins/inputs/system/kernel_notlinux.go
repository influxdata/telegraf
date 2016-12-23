// +build !linux

package system

import (
	"github.com/influxdata/telegraf/plugins"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Kernel struct {
}

func (k *Kernel) Description() string {
	return "Get kernel statistics from /proc/stat"
}

func (k *Kernel) SampleConfig() string { return "" }

func (k *Kernel) Gather(acc plugins.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("kernel", func() plugins.Input {
		return &Kernel{}
	})
}
