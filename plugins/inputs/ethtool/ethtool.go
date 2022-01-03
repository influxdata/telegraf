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

	// Normalization on the key names
	NormalizeKeys []string `toml:"normalize_keys"`

	Log telegraf.Logger `toml:"-"`

	// the ethtool command
	command Command
}

const (
	pluginName       = "ethtool"
	tagInterface     = "interface"
	tagDriverName    = "driver"
	fieldInterfaceUp = "interface_up"

	sampleConfig = `
  ## List of interfaces to pull metrics for
  # interface_include = ["eth0"]

  ## List of interfaces to ignore when pulling metrics.
  # interface_exclude = ["eth1"]

  ## Some drivers declare statistics with extra whitespace, different spacing,
  ## and mix cases. This list, when enabled, can be used to clean the keys.
  ## Here are the current possible normalizations:
  ##  * snakecase: converts fooBarBaz to foo_bar_baz
  ##  * trim: removes leading and trailing whitespace
  ##  * lower: changes all capitalized letters to lowercase
  ##  * underscore: replaces spaces with underscores
  # normalize_keys = ["snakecase", "trim", "lower", "underscore"]
`
)

func (e *Ethtool) SampleConfig() string {
	return sampleConfig
}

// Description returns a one-sentence description on the Input
func (e *Ethtool) Description() string {
	return "Returns ethtool statistics for given interfaces"
}
