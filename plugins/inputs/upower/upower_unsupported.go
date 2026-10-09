//go:generate ../../../tools/readme_config_includer/generator
//go:build windows || darwin

package upower

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type UPower struct {
	Log telegraf.Logger `toml:"-"`
}

func (*UPower) SampleConfig() string { return sampleConfig }

func (u *UPower) Init() error {
	u.Log.Warn("Current platform is not supported")
	return nil
}

func (*UPower) Gather(telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("upower", func() telegraf.Input {
		return &UPower{}
	})
}
