package system

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/shirou/gopsutil/cpu"
)

// cpuCoreUsageStats is a CPU core usage in percents up to 100
type cpuCoreUsageStats struct {
	User      float64
	System    float64
	Idle      float64
	Nice      float64
	Iowait    float64
	Irq       float64
	Softirq   float64
	Steal     float64
	Guest     float64
	GuestNice float64
}

// Total returns total CPU usage: 100% - idle
func (s *cpuCoreUsageStats) Total() float64 {
	return s.User + s.System + s.Nice + s.Iowait + s.Irq + s.Softirq + s.Steal + s.Guest + s.GuestNice
}

type CPUStats struct {
	ps        PS
	lastStats []cpu.TimesStat

	PerCPU                 bool `toml:"percpu"`
	TotalCPU               bool `toml:"totalcpu"`
	CollectCPUTime         bool `toml:"collect_cpu_time"`
	CollectSummaryCPUUsage bool `toml:"collect_summary_cpu_usage"`
}

func NewCPUStats(ps PS) *CPUStats {
	return &CPUStats{
		ps:             ps,
		CollectCPUTime: true,
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
  ## If true, collect raw CPU time metrics.
  collect_cpu_time = false
  ## If true, collect summary CPU usage
  collect_summary_cpu_usage = true
`

func (_ *CPUStats) SampleConfig() string {
	return sampleConfig
}

func (s *CPUStats) Gather(acc telegraf.Accumulator) error {
	cpuCoreUsage := &cpuCoreUsageStats{}
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

		if s.CollectCPUTime {
			// Add cpu time metrics
			fieldsC := map[string]interface{}{
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
			acc.AddCounter("cpu", fieldsC, tags, now)
		}

		// Add in percentage
		if len(s.lastStats) == 0 {
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

		cpuCoreUsage.User = 100 * (cts.User - lastCts.User - (cts.Guest - lastCts.Guest)) / totalDelta
		cpuCoreUsage.System = 100 * (cts.System - lastCts.System) / totalDelta
		cpuCoreUsage.Idle = 100 * (cts.Idle - lastCts.Idle) / totalDelta
		cpuCoreUsage.Nice = 100 * (cts.Nice - lastCts.Nice - (cts.GuestNice - lastCts.GuestNice)) / totalDelta
		cpuCoreUsage.Iowait = 100 * (cts.Iowait - lastCts.Iowait) / totalDelta
		cpuCoreUsage.Irq = 100 * (cts.Irq - lastCts.Irq) / totalDelta
		cpuCoreUsage.Softirq = 100 * (cts.Softirq - lastCts.Softirq) / totalDelta
		cpuCoreUsage.Steal = 100 * (cts.Steal - lastCts.Steal) / totalDelta
		cpuCoreUsage.Guest = 100 * (cts.Guest - lastCts.Guest) / totalDelta
		cpuCoreUsage.GuestNice = 100 * (cts.GuestNice - lastCts.GuestNice) / totalDelta

		fieldsG := map[string]interface{}{
			"usage_user":       cpuCoreUsage.User,
			"usage_system":     cpuCoreUsage.System,
			"usage_idle":       cpuCoreUsage.Idle,
			"usage_nice":       cpuCoreUsage.Nice,
			"usage_iowait":     cpuCoreUsage.Iowait,
			"usage_irq":        cpuCoreUsage.Irq,
			"usage_softirq":    cpuCoreUsage.Softirq,
			"usage_steal":      cpuCoreUsage.Steal,
			"usage_guest":      cpuCoreUsage.Guest,
			"usage_guest_nice": cpuCoreUsage.GuestNice,
		}
		acc.AddGauge("cpu", fieldsG, tags, now)

		if s.CollectSummaryCPUUsage {
			fieldsSummary := map[string]interface{}{
				"usage_total": cpuCoreUsage.Total(),
			}
			acc.AddGauge("cpu", fieldsSummary, tags, now)
		}
	}

	s.lastStats = times

	return nil
}

func totalCpuTime(t cpu.TimesStat) float64 {
	total := t.User + t.System + t.Nice + t.Iowait + t.Irq + t.Softirq + t.Steal +
		t.Idle
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
