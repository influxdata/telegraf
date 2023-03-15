//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package dpdk

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Dpdk struct {
	Log telegraf.Logger `toml:"-"`
}

func (d *Dpdk) Init() error {
	d.Log.Warn("current platform is not supported")
	return nil
}
func (*Dpdk) SampleConfig() string                { return sampleConfig }
func (*Dpdk) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("dpdk", func() telegraf.Input {
		return &Dpdk{}
	})
}
