package system

import (
	"fmt"

	"github.com/influxdb/telegraf/plugins"
	"github.com/shirou/gopsutil/cpu"
)

type CPUStats struct {
	ps        PS
	lastStats []cpu.CPUTimesStat

	PerCPU   bool `toml:"percpu"`
	TotalCPU bool `toml:"totalcpu"`
}

func NewCPUStats(ps PS) *CPUStats {
	return &CPUStats{
		ps: ps,
	}
}

func (_ *CPUStats) Description() string {
	return "Read metrics about cpu usage"
}

var sampleConfig = `
  # Whether to report per-cpu stats or not
  percpu = true
  # Whether to report total system cpu stats or not
  totalcpu = true
  # Comment this line if you want the raw CPU time metrics
  drop = ["cpu_time"]
`

func (_ *CPUStats) SampleConfig() string {
	return sampleConfig
}

func (s *CPUStats) Gather(acc plugins.Accumulator) error {
	times, err := s.ps.CPUTimes(s.PerCPU, s.TotalCPU)
	if err != nil {
		return fmt.Errorf("error getting CPU info: %s", err)
	}

	for i, cts := range times {
		tags := map[string]string{
			"cpu": cts.CPU,
		}

		total := totalCpuTime(cts)

		// Add total cpu numbers
		add(acc, "time_user", cts.User, tags)
		add(acc, "time_system", cts.System, tags)
		add(acc, "time_idle", cts.Idle, tags)
		add(acc, "time_nice", cts.Nice, tags)
		add(acc, "time_iowait", cts.Iowait, tags)
		add(acc, "time_irq", cts.Irq, tags)
		add(acc, "time_softirq", cts.Softirq, tags)
		add(acc, "time_steal", cts.Steal, tags)
		add(acc, "time_guest", cts.Guest, tags)
		add(acc, "time_guest_nice", cts.GuestNice, tags)

		// Add in percentage
		if len(s.lastStats) == 0 {
			// If it's the 1st gather, can't get CPU stats yet
			continue
		}
		lastCts := s.lastStats[i]
		lastTotal := totalCpuTime(lastCts)
		totalDelta := total - lastTotal

		if totalDelta < 0 {
			return fmt.Errorf("Error: current total CPU time is less than previous total CPU time")
		}

		if totalDelta == 0 {
			continue
		}

		add(acc, "usage_user", 100*(cts.User-lastCts.User)/totalDelta, tags)
		add(acc, "usage_system", 100*(cts.System-lastCts.System)/totalDelta, tags)
		add(acc, "usage_idle", 100*(cts.Idle-lastCts.Idle)/totalDelta, tags)
		add(acc, "usage_nice", 100*(cts.Nice-lastCts.Nice)/totalDelta, tags)
		add(acc, "usage_iowait", 100*(cts.Iowait-lastCts.Iowait)/totalDelta, tags)
		add(acc, "usage_irq", 100*(cts.Irq-lastCts.Irq)/totalDelta, tags)
		add(acc, "usage_softirq", 100*(cts.Softirq-lastCts.Softirq)/totalDelta, tags)
		add(acc, "usage_steal", 100*(cts.Steal-lastCts.Steal)/totalDelta, tags)
		add(acc, "usage_guest", 100*(cts.Guest-lastCts.Guest)/totalDelta, tags)
		add(acc, "usage_guest_nice", 100*(cts.GuestNice-lastCts.GuestNice)/totalDelta, tags)

	}

	s.lastStats = times

	return nil
}

func totalCpuTime(t cpu.CPUTimesStat) float64 {
	total := t.User + t.System + t.Nice + t.Iowait + t.Irq + t.Softirq + t.Steal +
		t.Guest + t.GuestNice + t.Idle
	return total
}

func init() {
	plugins.Add("cpu", func() plugins.Plugin {
		return &CPUStats{ps: &systemPS{}}
	})
}
