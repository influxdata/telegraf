// +build !linux

package cgroup

import (
	"github.com/influxdata/telegraf/plugins"
)

func (g *CGroup) Gather(acc plugins.Accumulator) error {
	return nil
}
