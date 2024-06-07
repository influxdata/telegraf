//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux || !amd64

package intel_pmt

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type IntelPMT struct {
	Log telegraf.Logger `toml:"-"`
}

func (p *IntelPMT) Init() error {
	p.Log.Warn("Current platform is not supported")
	return nil
}
func (*IntelPMT) SampleConfig() string                { return sampleConfig }
func (*IntelPMT) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("intel_pmt", func() telegraf.Input {
		return &IntelPMT{}
	})
}
