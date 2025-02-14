//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package bcache

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Bcache struct {
	Log telegraf.Logger `toml:"-"`
}

func (*Bcache) SampleConfig() string { return sampleConfig }

func (b *Bcache) Init() error {
	b.Log.Warn("Current platform is not supported")
	return nil
}

func (*Bcache) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("bcache", func() telegraf.Input {
		return &Bcache{}
	})
}
