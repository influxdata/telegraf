// +build !linux

package cgroup

import (
	"github.com/masami10/telegraf"
)

func (g *CGroup) Gather(acc telegraf.Accumulator) error {
	return nil
}
