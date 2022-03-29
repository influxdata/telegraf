//go:generate go run ../../../scripts/generate_plugindata/main.go
//go:generate go run ../../../scripts/generate_plugindata/main.go --clean
package wireless

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Wireless is used to store configuration values.
type Wireless struct {
	HostProc string          `toml:"host_proc"`
	Log      telegraf.Logger `toml:"-"`
}

// SampleConfig displays configuration instructions.
func (w *Wireless) SampleConfig() string {
	return `{{ .SampleConfig }}`
}

func init() {
	inputs.Add("wireless", func() telegraf.Input {
		return &Wireless{}
	})
}
