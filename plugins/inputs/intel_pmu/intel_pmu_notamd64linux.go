//go:build !linux || !amd64

package intel_pmu

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type IntelPMU struct {
	Log telegraf.Logger `toml:"-"`
}

func (i *IntelPMU) Init() error {
	i.Log.Warn("current platform is not supported")
	return nil
}
func (i *IntelPMU) SampleConfig() string                { return sampleConfig }
func (i *IntelPMU) Gather(_ telegraf.Accumulator) error { return nil }
func (i *IntelPMU) Start(_ telegraf.Accumulator) error  { return nil }
func (i *IntelPMU) Stop()                               {}

func init() {
	inputs.Add("intel_pmu", func() telegraf.Input {
		return &IntelPMU{}
	})
}
