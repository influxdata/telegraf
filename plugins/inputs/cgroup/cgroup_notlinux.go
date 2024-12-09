//go:build !linux

package cgroup

import (
	"github.com/influxdata/telegraf"
)

func (*CGroup) Gather(_ telegraf.Accumulator) error {
	return nil
}
