//go:build !linux

package psi

import "github.com/influxdata/telegraf"

func (psi *Psi) Gather(_ telegraf.Accumulator) error {
	psi.Log.Warn("Pressure Stall Information is only supported on Linux")
	return nil
}
