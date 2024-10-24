//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux || !amd64

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

func (*IntelPowerstat) SampleConfig() string { return sampleConfig }

func (i *IntelPowerstat) Init() error {
	i.Log.Warn("Current platform is not supported")
	return nil
}

func (*IntelPowerstat) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("intel_powerstat", func() telegraf.Input {
		return &IntelPowerstat{}
	})
}
