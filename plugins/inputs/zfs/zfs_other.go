//go:build !linux && !freebsd

package zfs

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type helper struct{} //nolint:unused // not used for "other" OSes, needed for Zfs struct

func (z *Zfs) Init() error {
	z.Log.Warn("Current platform is not supported")
	return nil
}

func (*Zfs) Gather(telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("zfs", func() telegraf.Input {
		return &Zfs{}
	})
}
