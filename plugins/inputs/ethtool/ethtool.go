//go:generate ../../../tools/readme_config_includer/generator
package ethtool

import (
	_ "embed"
)

const pluginName = "ethtool"

//go:embed sample.conf
var sampleConfig string

type Command interface {
	Init() error
	DriverName(intf NamespacedInterface) (string, error)
	Interfaces(includeNamespaces bool) ([]NamespacedInterface, error)
	Stats(intf NamespacedInterface) (map[string]uint64, error)
	Get(intf NamespacedInterface) (map[string]uint64, error)
}

func (*Ethtool) SampleConfig() string {
	return sampleConfig
}
