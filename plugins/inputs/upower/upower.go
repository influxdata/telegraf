//go:generate ../../../tools/readme_config_includer/generator
package upower

import (
	"context"
	_ "embed"
	"time"

	"github.com/godbus/dbus/v5"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type UPower struct {
	DeviceTypes []string        `toml:"device_types"`
	Timeout     config.Duration `toml:"timeout"`
	Log         telegraf.Logger `toml:"-"`

	client client //nolint:unused,nolintlint // Only used on supported platforms
}

type client interface { //nolint:unused,nolintlint // Only used on supported platforms
	connected() bool
	close() error
	enumerateDevices(ctx context.Context) ([]dbus.ObjectPath, error)
	deviceProperties(ctx context.Context, path dbus.ObjectPath) (map[string]dbus.Variant, error)
}

func (*UPower) SampleConfig() string {
	return sampleConfig
}

func init() {
	inputs.Add("upower", func() telegraf.Input {
		return &UPower{Timeout: config.Duration(5 * time.Second)}
	})
}
