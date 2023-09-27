//go:generate ../../../tools/readme_config_includer/generator
//go:build !linux

package hugepages

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Hugepages struct {
	Log telegraf.Logger `toml:"-"`
}

func (h *Hugepages) Init() error {
	h.Log.Warn("current platform is not supported")
	return nil
}
func (*Hugepages) SampleConfig() string                { return sampleConfig }
func (*Hugepages) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("hugepages", func() telegraf.Input {
		return &Hugepages{}
	})
}
