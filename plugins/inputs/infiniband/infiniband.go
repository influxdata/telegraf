package infiniband

import (
	"github.com/influxdata/telegraf"
)

// Stores the configuration values for the infiniband plugin - as there are no
// config values, this is intentionally empty
type Infiniband struct {
	Log telegraf.Logger `toml:"-"`
}

// Sample configuration for plugin
var InfinibandConfig = ``

func (i *Infiniband) SampleConfig() string {
	return InfinibandConfig
}

func (i *Infiniband) Description() string {
	return "Gets counters from all InfiniBand cards and ports installed"
}
