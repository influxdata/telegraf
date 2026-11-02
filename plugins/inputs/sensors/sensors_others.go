//go:build !linux && !openbsd

package sensors

import "github.com/influxdata/telegraf"

func (s *Sensors) Init() error {
	s.Log.Warn("Current platform is not supported")
	// Used on Linux and OpenBSD only; referenced here so the unused linter
	// is quiet. A //nolint:unused on the fields cannot be used instead, as
	// nolintlint would then flag it as unnecessary on the other platforms.
	_ = s.path
	_ = s.deviceFilter
	return nil
}

func (*Sensors) Gather(telegraf.Accumulator) error { return nil }
