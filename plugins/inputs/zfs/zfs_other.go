// +build !linux,!freebsd

package zfs

import (
	"github.com/influxdata/telegraf/plugins"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func (z *Zfs) Gather(acc plugins.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("zfs", func() plugins.Input {
		return &Zfs{}
	})
}
