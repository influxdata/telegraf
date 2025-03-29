//go:generate ../../../tools/readme_config_includer/generator
package zfs

import (
	_ "embed"
)

//go:embed sample.conf
var sampleConfig string

func (*Zfs) SampleConfig() string {
	return sampleConfig
}
