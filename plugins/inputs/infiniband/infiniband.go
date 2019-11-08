package infiniband

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
