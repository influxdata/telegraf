//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package timex

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Timex struct {
	Log telegraf.Logger `toml:"-"`
}

func (*Timex) SampleConfig() string {
	return sampleConfig
}

func (tx *Timex) Init() error {
	tx.Log.Warn("Current platform is not supported")
	return nil
}

func (*Timex) Gather(telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("timex", func() telegraf.Input {
		return &Timex{}
	})
}
