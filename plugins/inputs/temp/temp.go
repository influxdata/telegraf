//go:generate ../../../tools/readme_config_includer/generator
package temp

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Temperature struct {
	MetricFormat string          `toml:"metric_format"`
	DeviceTag    bool            `toml:"add_device_tag"`
	Log          telegraf.Logger `toml:"-"`
}

func (*Temperature) SampleConfig() string {
	return sampleConfig
}

func init() {
	inputs.Add("temp", func() telegraf.Input {
		return &Temperature{}
	})
}
