// +build !linux,!freebsd

package zfs

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/registry/inputs"
)

func (z *Zfs) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("zfs", func() telegraf.Input {
		return &Zfs{}
	})
}
