//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package intel_powerstat

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type IntelPowerstat struct {
	Log telegraf.Logger `toml:"-"`
}

func (i *IntelPowerstat) Init() error {
	i.Log.Warn("current platform is not supported")
	return nil
}
func (*IntelPowerstat) SampleConfig() string                { return sampleConfig }
func (*IntelPowerstat) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("intel_powerstat", func() telegraf.Input {
		return &IntelPowerstat{}
	})
}
