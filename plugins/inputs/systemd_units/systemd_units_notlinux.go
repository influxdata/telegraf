//go:build !linux

package systemd_units

import "github.com/influxdata/telegraf"

func (s *SystemdUnits) Init() error {
	s.Log.Info("Skipping plugin as it is not supported by this platform!")

	// Required to remove linter-warning on unused struct member
	_ = s.client

	return nil
}

func (*SystemdUnits) Gather(_ telegraf.Accumulator) error {
	return nil
}
