//go:generate ../../../tools/readme_config_includer/generator
package zfs

import (
	_ "embed"

	"github.com/influxdata/telegraf"
)

//go:embed sample.conf
var sampleConfig string

type Zfs struct {
	KstatPath      string          `toml:"kstatPath"`
	KstatMetrics   []string        `toml:"kstatMetrics"`
	PoolMetrics    bool            `toml:"poolMetrics"`
	DatasetMetrics bool            `toml:"datasetMetrics"`
	UseNativeTypes bool            `toml:"useNativeTypes"`
	Log            telegraf.Logger `toml:"-"`

	helper //nolint:unused // for OS-specific usage
}

func (*Zfs) SampleConfig() string {
	return sampleConfig
}
