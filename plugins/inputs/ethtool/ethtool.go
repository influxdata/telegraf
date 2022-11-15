package ethtool

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
)

//go:embed sample.conf
var sampleConfig string

var downInterfacesBehaviors = []string{"expose", "skip"}

type Command interface {
	Init() error
	DriverName(intf NamespacedInterface) (string, error)
	Interfaces(includeNamespaces bool) ([]NamespacedInterface, error)
	Stats(intf NamespacedInterface) (map[string]uint64, error)
}

type Ethtool struct {
	// This is the list of interface names to include
	InterfaceInclude []string `toml:"interface_include"`

	// This is the list of interface names to ignore
	InterfaceExclude []string `toml:"interface_exclude"`

	// Behavior regarding metrics for downed interfaces
	DownInterfaces string `toml:" down_interfaces"`

	// This is the list of namespace names to include
	NamespaceInclude []string `toml:"namespace_include"`

	// This is the list of namespace names to ignore
	NamespaceExclude []string `toml:"namespace_exclude"`

	// Normalization on the key names
	NormalizeKeys []string `toml:"normalize_keys"`

	Log telegraf.Logger `toml:"-"`

	interfaceFilter   filter.Filter
	namespaceFilter   filter.Filter
	includeNamespaces bool

	// the ethtool command
	command Command
}

func (*Ethtool) SampleConfig() string {
	return sampleConfig
}

const (
	pluginName       = "ethtool"
	tagInterface     = "interface"
	tagNamespace     = "namespace"
	tagDriverName    = "driver"
	fieldInterfaceUp = "interface_up"
)
