//go:generate go run ../../../tools/generate_plugindata/main.go
//go:generate go run ../../../tools/nerate_plugindata/main.go --clean
package dmcache

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type DMCache struct {
	PerDevice        bool `toml:"per_device"`
	getCurrentStatus func() ([]string, error)
}

func (c *DMCache) SampleConfig() string {
	return `{{ .SampleConfig }}`
}

func init() {
	inputs.Add("dmcache", func() telegraf.Input {
		return &DMCache{
			PerDevice:        true,
			getCurrentStatus: dmSetupStatus,
		}
	})
}
