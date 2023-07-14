//go:build !linux

package diskio

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

type DiskIO struct {
	ps system.PS

	Devices          []string
	DeviceTags       []string
	NameTemplates    []string
	SkipSerialNumber bool

	Log telegraf.Logger

	deviceFilter filter.Filter
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
