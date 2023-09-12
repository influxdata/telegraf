//go:build !linux

package cgroup

import (
	"context"

	"github.com/influxdata/telegraf"
)

func (*CGroup) Gather(_ context.Context, _ telegraf.Accumulator) error {
	return nil
}
