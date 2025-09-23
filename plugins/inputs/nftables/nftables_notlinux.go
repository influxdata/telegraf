//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package nftables

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Nftables struct {
	Log telegraf.Logger `toml:"-"`
}

func (*Nftables) SampleConfig() string { return sampleConfig }

func (i *Nftables) Init() error {
	i.Log.Warn("Current platform is not supported")
	return nil
}

func (*Nftables) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("nftables", func() telegraf.Input {
		return &Nftables{}
	})
}
