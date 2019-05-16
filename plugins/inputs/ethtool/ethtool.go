package ethtool

import (
	"net"
	"sync"
)

type Command interface {
	DriverName(intf string) (string, error)
	Interfaces() ([]net.Interface, error)
	Stats(intf string) (map[string]uint64, error)
}

type Ethtool struct {
	// This is the list of interface names to include
	InterfaceInclude []string `toml:"interface_include"`

	// This is the list of interface names to ignore
	InterfaceExclude []string `toml:"interface_exclude"`

	// Whether to include the driver name in the tag
	DriverName bool `toml:"driver_name_tag"`

	// the ethtool command
	command Command

	// Will parallelize the ethtool call in event of many interfaces
	// so using this to sync
	wg sync.WaitGroup
}

const (
	pluginName    = "ethtool"
	tagInterface  = "interface"
	tagDriverName = "driver"

	sampleConfig = `
  ## List of interfaces to pull metrics for
  # interface_include = ["eth0"]

  ## List of interfaces to ignore when pulling metrics.  
  # interface_exclude = ["eth1"]

  ## Add driver information as tag
  # driver_name_tag = true
`
)

func (e *Ethtool) SampleConfig() string {
	return sampleConfig
}

// Description returns a one-sentence description on the Input
func (e *Ethtool) Description() string {
	return "Returns ethtool statistics for given interfaces"
}
