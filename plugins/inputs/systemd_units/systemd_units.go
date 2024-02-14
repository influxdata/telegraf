//go:generate ../../../tools/readme_config_includer/generator
package systemd_units

import (
	_ "embed"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// SystemdUnits is a telegraf plugin to gather systemd unit status
type SystemdUnits struct {
	Pattern    string          `toml:"pattern"`
	UnitType   string          `toml:"unittype"`
	SubCommand string          `toml:"subcommand"`
	Timeout    config.Duration `toml:"timeout"`
	Log        telegraf.Logger `toml:"-"`

	client client
}

func (*SystemdUnits) SampleConfig() string {
	return sampleConfig
}

func init() {
	inputs.Add("systemd_units", func() telegraf.Input {
		return &SystemdUnits{Timeout: config.Duration(time.Second)}
	})
}
