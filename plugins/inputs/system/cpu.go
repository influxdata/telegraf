package system

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/shirou/gopsutil/cpu"
)

type CPUStats struct {
	ps        PS
	lastStats []cpu.TimesStat

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
  ## Whether to report per-cpu stats or not
  percpu = true
  ## Whether to report total system cpu stats or not
  totalcpu = true
  ## Comment this line if you want the raw CPU time metrics
  fielddrop = ["time_*"]
`

func (_ *CPUStats) SampleConfig() string {
	return sampleConfig
}

func (s *CPUStats) Gather(acc telegraf.Accumulator) error {
	times, err := s.ps.CPUTimes(s.PerCPU, s.TotalCPU)
	if err != nil {
		return fmt.Errorf("error getting CPU info: %s", err)
	}
	now := time.Now()

	for i, cts := range times {
		tags := map[string]string{
			"cpu": cts.CPU,
		}

		total := totalCpuTime(cts)

		// Add cpu time metrics
		fields := map[string]interface{}{
			"time_user":       cts.User,
			"time_system":     cts.System,
			"time_idle":       cts.Idle,
			"time_nice":       cts.Nice,
			"time_iowait":     cts.Iowait,
			"time_irq":        cts.Irq,
			"time_softirq":    cts.Softirq,
			"time_steal":      cts.Steal,
			"time_guest":      cts.Guest,
			"time_guest_nice": cts.GuestNice,
		}

		// Add in percentage
		if len(s.lastStats) == 0 {
			acc.AddFields("cpu", fields, tags, now)
			// If it's the 1st gather, can't get CPU Usage stats yet
			continue
		}
		lastCts := s.lastStats[i]
		lastTotal := totalCpuTime(lastCts)
		totalDelta := total - lastTotal

		if totalDelta < 0 {
			s.lastStats = times
			return fmt.Errorf("Error: current total CPU time is less than previous total CPU time")
		}

		if totalDelta == 0 {
			continue
		}

		fields["usage_user"] = 100 * (cts.User - lastCts.User) / totalDelta
		fields["usage_system"] = 100 * (cts.System - lastCts.System) / totalDelta
		fields["usage_idle"] = 100 * (cts.Idle - lastCts.Idle) / totalDelta
		fields["usage_nice"] = 100 * (cts.Nice - lastCts.Nice) / totalDelta
		fields["usage_iowait"] = 100 * (cts.Iowait - lastCts.Iowait) / totalDelta
		fields["usage_irq"] = 100 * (cts.Irq - lastCts.Irq) / totalDelta
		fields["usage_softirq"] = 100 * (cts.Softirq - lastCts.Softirq) / totalDelta
		fields["usage_steal"] = 100 * (cts.Steal - lastCts.Steal) / totalDelta
		fields["usage_guest"] = 100 * (cts.Guest - lastCts.Guest) / totalDelta
		fields["usage_guest_nice"] = 100 * (cts.GuestNice - lastCts.GuestNice) / totalDelta
		acc.AddFields("cpu", fields, tags, now)
	}

	s.lastStats = times

	return nil
}

func totalCpuTime(t cpu.TimesStat) float64 {
	total := t.User + t.System + t.Nice + t.Iowait + t.Irq + t.Softirq + t.Steal +
		t.Guest + t.GuestNice + t.Idle
	return total
}

func init() {
	inputs.Add("cpu", func() telegraf.Input {
		return &CPUStats{
			PerCPU:   true,
			TotalCPU: true,
			ps:       &systemPS{},
		}
	})
}
