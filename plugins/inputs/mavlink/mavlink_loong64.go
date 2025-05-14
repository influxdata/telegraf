//go:generate ../../../tools/readme_config_includer/generator
//go:build loong64

package mavlink

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Mavlink struct {
	Log telegraf.Logger `toml:"-"`
}

func (*Mavlink) SampleConfig() string { return sampleConfig }

func (m *Mavlink) Init() error {
	m.Log.Warn("Current platform is not supported")
	return nil
}

func (*Mavlink) Gather(telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("mavlink", func() telegraf.Input { return &Mavlink{} })
}
