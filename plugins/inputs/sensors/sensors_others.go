//go:build !linux && !openbsd

package sensors

import "github.com/influxdata/telegraf"

func (s *Sensors) Init() error {
	s.Log.Warn("Current platform is not supported")
	return nil
}

func (*Sensors) Gather(telegraf.Accumulator) error { return nil }
