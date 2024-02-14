//go:generate ../../../tools/readme_config_includer/generator
package systemd_units

import (
	"context"
	_ "embed"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type client interface {
	Connected() bool
	Close()

	ListUnitFilesByPatternsContext(ctx context.Context, states, pattern []string) ([]dbus.UnitFile, error)
	ListUnitsByNamesContext(ctx context.Context, units []string) ([]dbus.UnitStatus, error)
	GetUnitTypePropertiesContext(ctx context.Context, unit, unitType string) (map[string]interface{}, error)
	GetUnitPropertyContext(ctx context.Context, unit, propertyName string) (*dbus.Property, error)
	ListUnitsContext(ctx context.Context) ([]dbus.UnitStatus, error)
}

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
