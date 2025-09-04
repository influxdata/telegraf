//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux || !amd64

package turbostat

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Turbostat struct {
	Log telegraf.Logger `toml:"-"`
}

func (*Turbostat) SampleConfig() string { return sampleConfig }

func (s *Turbostat) Init() error {
	s.Log.Warn("Current platform is not supported")
	return nil
}

func (*Turbostat) Gather(telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("turbostat", func() telegraf.Input {
		return &Turbostat{}
	})
}
