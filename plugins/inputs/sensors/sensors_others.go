//go:build !linux && !openbsd

package sensors

import "github.com/influxdata/telegraf"

func (s *Sensors) Init() error {
	s.Log.Warn("Current platform is not supported")
	// Used on Linux only; referenced here so the unused linter is quiet.
	_ = measurement
	_ = s.path
	return nil
}

func (*Sensors) Gather(telegraf.Accumulator) error { return nil }
