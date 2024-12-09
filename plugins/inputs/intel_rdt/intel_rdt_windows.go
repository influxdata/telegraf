//go:build windows

package intel_rdt

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type IntelRDT struct {
	Log telegraf.Logger `toml:"-"`
}

func (*IntelRDT) SampleConfig() string { return sampleConfig }

func (i *IntelRDT) Init() error {
	i.Log.Warn("Current platform is not supported")
	return nil
}

func (*IntelRDT) Start(_ telegraf.Accumulator) error { return nil }

func (*IntelRDT) Gather(_ telegraf.Accumulator) error { return nil }

func (*IntelRDT) Stop() {}

func init() {
	inputs.Add("intel_rdt", func() telegraf.Input {
		return &IntelRDT{}
	})
}
