//go:generate ../../../tools/readme_config_includer/generator
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
	Log telegraf.Logger `toml:"-"`
}

func (k *Kernel) Init() error {
	k.Log.Warn("current platform is not supported")
	return nil
}
func (*Kernel) SampleConfig() string                { return sampleConfig }
func (*Kernel) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("kernel", func() telegraf.Input {
		return &Kernel{}
	})
}
