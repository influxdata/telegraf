//go:generate ../../../tools/readme_config_includer/generator
//go:build !windows

package win_wmi

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Wmi struct {
	Log telegraf.Logger `toml:"-"`
}

func (w *Wmi) Init() error {
	w.Log.Warn("current platform is not supported")
	return nil
}
func (w *Wmi) SampleConfig() string                { return sampleConfig }
func (w *Wmi) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("win_wmi", func() telegraf.Input {
		return &Wmi{}
	})
}
