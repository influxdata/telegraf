//go:generate ../../../tools/readme_config_includer/generator
package ethtool

import (
	_ "embed"
	"net"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
)

//go:embed sample.conf
var sampleConfig string

var downInterfacesBehaviors = []string{"expose", "skip"}

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

	// Behavior regarding metrics for downed interfaces
	DownInterfaces string `toml:" down_interfaces"`

	// Normalization on the key names
	NormalizeKeys []string `toml:"normalize_keys"`

	Log telegraf.Logger `toml:"-"`

	interfaceFilter filter.Filter

	// the ethtool command
	command Command
}

func (*Ethtool) SampleConfig() string {
	return sampleConfig
}

const (
	pluginName       = "ethtool"
	tagInterface     = "interface"
	tagDriverName    = "driver"
	fieldInterfaceUp = "interface_up"
)
