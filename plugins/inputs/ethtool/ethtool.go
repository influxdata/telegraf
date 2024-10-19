//go:generate ../../../tools/readme_config_includer/generator
package ethtool

import (
	_ "embed"
)

//go:embed sample.conf
var sampleConfig string

const pluginName = "ethtool"

func (*Ethtool) SampleConfig() string {
	return sampleConfig
}
