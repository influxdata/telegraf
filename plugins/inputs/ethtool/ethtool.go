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
)
