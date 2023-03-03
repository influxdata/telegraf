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

func (i *Ipvs) Init() error {
	i.Log.Warn("current platform is not supported")
	return nil
}
func (i *Ipvs) SampleConfig() string                { return sampleConfig }
func (i *Ipvs) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("ipvs", func() telegraf.Input {
		return &Ipvs{}
	})
}
