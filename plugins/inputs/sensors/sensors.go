//go:generate ../../../tools/readme_config_includer/generator
package sensors

import (
	_ "embed"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Sensors struct {
	RemoveNumbers       bool            `toml:"remove_numbers"`
	LinuxLegacyTagNames bool            `toml:"linux_legacy_tag_names"`
	Devices             []string        `toml:"devices"`
	Timeout             config.Duration `toml:"timeout"`
	Log                 telegraf.Logger `toml:"-"`

	path         string
	deviceFilter filter.Filter
}

func (*Sensors) SampleConfig() string {
	return sampleConfig
}

func init() {
	inputs.Add("sensors", func() telegraf.Input {
		return &Sensors{
			RemoveNumbers:       true,
			LinuxLegacyTagNames: true,
			Timeout:             config.Duration(5 * time.Second),
		}
	})
}
