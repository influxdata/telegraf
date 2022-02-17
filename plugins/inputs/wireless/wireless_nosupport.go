//go:build !linux && !darwin
// +build !linux,!darwin

package wireless

import (
	"github.com/influxdata/telegraf"
)

func (w *Wireless) Init() error {
	w.Log.Warn("Current platform is not supported")
	return nil
}

func (w *Wireless) Gather(acc telegraf.Accumulator) error {
	return nil
}
