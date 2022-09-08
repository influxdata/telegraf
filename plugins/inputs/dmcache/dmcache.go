//go:generate ../../../tools/readme_config_includer/generator
package dmcache

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//
//go:embed sample.conf
var sampleConfig string

type DMCache struct {
	PerDevice        bool `toml:"per_device"`
	getCurrentStatus func() ([]string, error)
}

func (*DMCache) SampleConfig() string {
	return sampleConfig
}

func init() {
	inputs.Add("dmcache", func() telegraf.Input {
		return &DMCache{
			PerDevice:        true,
			getCurrentStatus: dmSetupStatus,
		}
	})
}
