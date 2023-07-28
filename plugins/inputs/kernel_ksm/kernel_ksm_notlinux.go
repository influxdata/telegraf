//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package kernel_ksm

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type KernelKsm struct {
	Log telegraf.Logger `toml:"-"`
}

func (k *KernelKsm) Init() error {
	k.Log.Warn("current platform is not supported")
	return nil
}
func (*KernelKsm) SampleConfig() string                { return sampleConfig }
func (*KernelKsm) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("kernel_ksm", func() telegraf.Input {
		return &KernelKsm{}
	})
}
