//go:build !linux

package diskio

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

type DiskIO struct {
	Devices          []string `toml:"devices"`
	DeviceTags       []string `toml:"device_tags"`
	NameTemplates    []string `toml:"name_templates"`
	SkipSerialNumber bool     `toml:"skip_serial_number"`

	Log telegraf.Logger `toml:"-"`

	deviceFilter filter.Filter

	ps system.PS
}

func (*DiskIO) diskInfo(_ string) (map[string]string, error) {
	return nil, nil
}

func resolveName(name string) string {
	return name
}

func getDeviceWWID(_ string) string {
	return ""
}
