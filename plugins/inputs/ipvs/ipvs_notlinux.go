//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package ipvs

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Ipvs struct {
	Log telegraf.Logger `toml:"-"`
}

func (*Ipvs) SampleConfig() string { return sampleConfig }

func (i *Ipvs) Init() error {
	i.Log.Warn("Current platform is not supported")
	return nil
}

func (*Ipvs) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("ipvs", func() telegraf.Input {
		return &Ipvs{}
	})
}
