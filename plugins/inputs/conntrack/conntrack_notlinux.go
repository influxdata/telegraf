//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package conntrack

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Conntrack struct {
	Log telegraf.Logger `toml:"-"`
}

func (c *Conntrack) Init() error {
	c.Log.Warn("current platform is not supported")
	return nil
}
func (*Conntrack) SampleConfig() string                { return sampleConfig }
func (*Conntrack) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("conntrack", func() telegraf.Input {
		return &Conntrack{}
	})
}
