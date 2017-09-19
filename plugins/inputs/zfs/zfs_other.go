// +build !linux,!freebsd

package zfs

import (
	"github.com/masami10/telegraf"
	"github.com/masami10/telegraf/plugins/inputs"
)

func (z *Zfs) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("zfs", func() telegraf.Input {
		return &Zfs{}
	})
}
