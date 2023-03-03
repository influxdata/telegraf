//go:build !linux

package kernel

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Kernel struct {
}

func (k *Kernel) SampleConfig() string { return sampleConfig }
func (k *Kernel) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("kernel", func() telegraf.Input {
		return &Kernel{}
	})
}
