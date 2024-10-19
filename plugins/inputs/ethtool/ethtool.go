//go:generate ../../../tools/readme_config_includer/generator
package ethtool

import (
	_ "embed"
)

//go:embed sample.conf
var sampleConfig string

const pluginName = "ethtool"

type command interface {
	init() error
	driverName(intf namespacedInterface) (string, error)
	interfaces(includeNamespaces bool) ([]namespacedInterface, error)
	stats(intf namespacedInterface) (map[string]uint64, error)
	get(intf namespacedInterface) (map[string]uint64, error)
}

func (*Ethtool) SampleConfig() string {
	return sampleConfig
}
