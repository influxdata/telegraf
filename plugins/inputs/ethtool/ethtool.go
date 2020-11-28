package ethtool

import (
	"net"

	"github.com/influxdata/telegraf"
)

type Command interface {
	Init() error
	DriverName(intf string) (string, error)
	Interfaces() ([]net.Interface, error)
	Stats(intf string) (map[string]uint64, error)
}

type Ethtool struct {
	// This is the list of interface names to include
	InterfaceInclude []string `toml:"interface_include"`

	// This is the list of interface names to ignore
	InterfaceExclude []string `toml:"interface_exclude"`

	Log telegraf.Logger `toml:"-"`

	// the ethtool command
	command Command
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
`
)

func (e *Ethtool) SampleConfig() string {
	return sampleConfig
}

// Description returns a one-sentence description on the Input
func (e *Ethtool) Description() string {
	return "Returns ethtool statistics for given interfaces"
}
