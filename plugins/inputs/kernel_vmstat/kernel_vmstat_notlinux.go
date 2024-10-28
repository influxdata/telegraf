//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package kernel_vmstat

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type KernelVmstat struct {
	Log telegraf.Logger `toml:"-"`
}

func (*KernelVmstat) SampleConfig() string { return sampleConfig }

func (k *KernelVmstat) Init() error {
	k.Log.Warn("Current platform is not supported")
	return nil
}

func (*KernelVmstat) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("kernel_vmstat", func() telegraf.Input {
		return &KernelVmstat{}
	})
}
