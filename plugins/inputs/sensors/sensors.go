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

const measurement = "sensors"

var defaultTimeout = config.Duration(5 * time.Second)

const defaultBinary = "/sbin/sysctl"

type Sensors struct {
	RemoveNumbers bool            `toml:"remove_numbers"`
	MetricVersion int             `toml:"metric_version"`
	Binary        string          `toml:"binary"`
	Devices       []string        `toml:"devices"`
	Timeout       config.Duration `toml:"timeout"`
	Log           telegraf.Logger `toml:"-"`

	path         string
	deviceFilter filter.Filter
	run          runner
}

func (*Sensors) SampleConfig() string {
	return sampleConfig
}

func init() {
	inputs.Add("sensors", func() telegraf.Input {
		return &Sensors{
			RemoveNumbers: true,
			Timeout:       defaultTimeout,
			Binary:        defaultBinary,
		}
	})
}
