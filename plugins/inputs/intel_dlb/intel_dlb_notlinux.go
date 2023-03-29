//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package intel_dlb

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type IntelDLB struct {
	Log telegraf.Logger `toml:"-"`
}

func (i *IntelDLB) Init() error {
	i.Log.Warn("current platform is not supported")
	return nil
}
func (*IntelDLB) SampleConfig() string                { return sampleConfig }
func (*IntelDLB) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("intel_dlb", func() telegraf.Input {
		return &IntelDLB{}
	})
}
