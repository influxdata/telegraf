package system

import (
	"fmt"

	"github.com/influxdb/tivan/plugins"
)

type CPUStats struct {
	ps PS
}

func (_ *CPUStats) Description() string {
	return "Read metrics about cpu usage"
}

func (_ *CPUStats) SampleConfig() string { return "" }

func (s *CPUStats) Gather(acc plugins.Accumulator) error {
	times, err := s.ps.CPUTimes()
	if err != nil {
		return fmt.Errorf("error getting CPU info: %s", err)
	}

	for _, cts := range times {
		tags := map[string]string{
			"cpu": cts.CPU,
		}

		add(acc, "user", cts.User, tags)
		add(acc, "system", cts.System, tags)
		add(acc, "idle", cts.Idle, tags)
		add(acc, "nice", cts.Nice, tags)
		add(acc, "iowait", cts.Iowait, tags)
		add(acc, "irq", cts.Irq, tags)
		add(acc, "softirq", cts.Softirq, tags)
		add(acc, "steal", cts.Steal, tags)
		add(acc, "guest", cts.Guest, tags)
		add(acc, "guestNice", cts.GuestNice, tags)
		add(acc, "stolen", cts.Stolen, tags)
	}

	return nil
}

func init() {
	plugins.Add("cpu", func() plugins.Plugin {
		return &CPUStats{ps: &systemPS{}}
	})
}
