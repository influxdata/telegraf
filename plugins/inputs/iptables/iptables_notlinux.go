//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package iptables

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Iptables struct {
	Log telegraf.Logger `toml:"-"`
}

func (*Iptables) SampleConfig() string { return sampleConfig }

func (i *Iptables) Init() error {
	i.Log.Warn("Current platform is not supported")
	return nil
}

func (*Iptables) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("iptables", func() telegraf.Input {
		return &Iptables{}
	})
}
