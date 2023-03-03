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

func (i *Iptables) Init() error {
	i.Log.Warn("current platform is not supported")
	return nil
}
func (i *Iptables) SampleConfig() string                { return sampleConfig }
func (i *Iptables) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("iptables", func() telegraf.Input {
		return &Iptables{}
	})
}
