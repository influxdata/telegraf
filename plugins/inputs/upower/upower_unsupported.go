//go:build windows || darwin

package upower

import (
	"github.com/influxdata/telegraf"
)

func (u *UPower) Init() error {
	u.Log.Warn("Current platform is not supported")
	return nil
}

func (*UPower) Gather(telegraf.Accumulator) error { return nil }
