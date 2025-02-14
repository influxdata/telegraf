//go:generate ../../../tools/readme_config_includer/generator
//go:build !windows

package win_eventlog

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type WinEventLog struct {
	Log telegraf.Logger `toml:"-"`
}

func (*WinEventLog) SampleConfig() string { return sampleConfig }

func (w *WinEventLog) Init() error {
	w.Log.Warn("Current platform is not supported")
	return nil
}

func (*WinEventLog) Gather(telegraf.Accumulator) error { return nil }

func init() {
	inputs.Add("win_eventlog", func() telegraf.Input {
		return &WinEventLog{}
	})
}
