//go:generate ../../../tools/readme_config_includer/generator
//go:build !windows

package win_services

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type WinServices struct {
	Log telegraf.Logger `toml:"-"`
}

func (w *WinServices) Init() error {
	w.Log.Warn("current platform is not supported")
	return nil
}
func (w *WinServices) SampleConfig() string                { return sampleConfig }
func (w *WinServices) Gather(_ telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("win_services", func() telegraf.Input {
		return &WinServices{}
	})
}
