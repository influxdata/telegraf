// +build !linux

package cgroup

import (
	"github.com/influxdata/telegraf"
)

func (g *CGroup) Gather(acc telegraf.Accumulator) error {
	return nil
}
