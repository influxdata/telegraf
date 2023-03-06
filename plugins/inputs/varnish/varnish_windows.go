//go:build windows

package varnish

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Varnish struct {
	Log telegraf.Logger `toml:"-"`
}

func (v *Varnish) Init() error {
	v.Log.Warn("current platform is not supported")
	return nil
}
func (*Varnish) SampleConfig() string                { return sampleConfig }
func (*Varnish) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("varnish", func() telegraf.Input {
		return &Varnish{}
	})
}
