package wireless_mac

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Wireless is used to store configuration values.
type Wireless struct {
	HostProc string          `toml:"host_proc"`
	Log      telegraf.Logger `toml:"-"`
}

var sampleConfig = `
  [[inputs.wireless_mac]]
  ## dump metrics with 0 values too
  dump_zeros       = true
`


// Description returns information about the plugin.
func (w *Wireless_Mac) Description() string {
	return "Monitor wifi signal strength and quality on macOS"
}

// SampleConfig displays configuration instructions.
func (w *Wireless_Mac) SampleConfig() string {
	return sampleConfig
}

func init() {
	inputs.Add("wireless_mac", func() telegraf.Input {
		return &Wireless_Mac{}
	})
}
