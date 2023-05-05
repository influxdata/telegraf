//go:build !linux

package ethtool

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const pluginName = "ethtool"

type Ethtool struct {
	Log telegraf.Logger `toml:"-"`
}

func (*Ethtool) SampleConfig() string {
	return sampleConfig
}

func (e *Ethtool) Init() error {
	e.Log.Warn("Current platform is not supported")
	return nil
}

func (*Ethtool) Gather(_ telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add(pluginName, func() telegraf.Input {
		return &Ethtool{}
	})
}
