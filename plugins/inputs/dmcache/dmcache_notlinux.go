//go:build !linux

package dmcache

import (
	"context"

	"github.com/influxdata/telegraf"
)

func (*DMCache) Gather(_ context.Context, _ telegraf.Accumulator) error {
	return nil
}

func dmSetupStatus() ([]string, error) {
	return []string{}, nil
}
