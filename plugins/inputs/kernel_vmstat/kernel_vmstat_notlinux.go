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

func (k *KernelVmstat) Init() error {
	k.Log.Warn("current platform is not supported")
	return nil
}
func (k *KernelVmstat) SampleConfig() string                { return sampleConfig }
func (k *KernelVmstat) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("kernel_vmstat", func() telegraf.Input {
		return &KernelVmstat{}
	})
}
