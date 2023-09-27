//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux || !amd64

package intel_baseband

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Baseband struct {
	Log telegraf.Logger `toml:"-"`
}

func (b *Baseband) Init() error {
	b.Log.Warn("current platform is not supported")
	return nil
}
func (*Baseband) SampleConfig() string                { return sampleConfig }
func (*Baseband) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("intel_baseband", func() telegraf.Input {
		return &Baseband{}
	})
}
