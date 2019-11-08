package infiniband

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Stores the configuration values for the infiniband plugin - as there are no
// config values, this is intentionally empty
type Infiniband struct {
}

// Sample configuration for plugin
var InfinibandConfig = `
  ## no config required
`

func (s *Infiniband) SampleConfig() string {
	return InfinibandConfig
}

func (s *Infiniband) Description() string {
	return "Gets counters from all InfiniBand cards and ports installed"
}

// Initialise plugin
func init() {
	inputs.Add("infiniband", func() telegraf.Input { return &Infiniband{} })
}
