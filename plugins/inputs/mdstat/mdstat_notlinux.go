//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package mdstat

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Mdstat struct {
	Log telegraf.Logger `toml:"-"`
}

func (m *Mdstat) Init() error {
	m.Log.Warn("current platform is not supported")
	return nil
}
func (*Mdstat) SampleConfig() string                { return sampleConfig }
func (*Mdstat) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("mdstat", func() telegraf.Input {
		return &Mdstat{}
	})
}
