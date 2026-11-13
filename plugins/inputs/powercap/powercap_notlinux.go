//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package powercap

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Powercap struct {
	Log telegraf.Logger `toml:"-"`
}

func (*Powercap) SampleConfig() string {
	return sampleConfig
}

func (p *Powercap) Init() error {
	p.Log.Warn("Current platform is not supported")
	return nil
}

func (*Powercap) Gather(telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("powercap", func() telegraf.Input {
		return &Powercap{}
	})
}
