//go:generate ../../../tools/readme_config_includer/generator
package cpu

import (
	_ "embed"
	"fmt"
	"time"

	cpuUtil "github.com/shirou/gopsutil/v3/cpu"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//
//go:embed sample.conf
var sampleConfig string

type CPUStats struct {
	ps         system.PS
	lastStats  map[string]cpuUtil.TimesStat
	cpuInfo    map[string]cpuUtil.InfoStat
	coreID     bool
	physicalID bool

	PerCPU         bool `toml:"percpu"`
	TotalCPU       bool `toml:"totalcpu"`
	CollectCPUTime bool `toml:"collect_cpu_time"`
	ReportActive   bool `toml:"report_active"`
	CoreTags       bool `toml:"core_tags"`

	Log telegraf.Logger `toml:"-"`
}

func NewCPUStats(ps system.PS) *CPUStats {
	return &CPUStats{
		ps:             ps,
		CollectCPUTime: true,
		ReportActive:   true,
	}
}

func (*CPUStats) SampleConfig() string {
	return sampleConfig
}

func (c *CPUStats) Gather(acc telegraf.Accumulator) error {
	times, err := c.ps.CPUTimes(c.PerCPU, c.TotalCPU)
	if err != nil {
		return fmt.Errorf("error getting CPU info: %s", err)
	}
	now := time.Now()

	for _, cts := range times {
		tags := map[string]string{
			"cpu": cts.CPU,
		}
		if c.coreID {
			tags["core_id"] = c.cpuInfo[cts.CPU].CoreID
		}
		if c.physicalID {
			tags["physical_id"] = c.cpuInfo[cts.CPU].PhysicalID
		}

		total := totalCPUTime(cts)
		active := activeCPUTime(cts)

		if c.CollectCPUTime {
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
			if c.ReportActive {
				fieldsC["time_active"] = activeCPUTime(cts)
			}
			acc.AddCounter("cpu", fieldsC, tags, now)
		}

		// Add in percentage
		if len(c.lastStats) == 0 {
			// If it's the 1st gather, can't get CPU Usage stats yet
			continue
		}

		lastCts, ok := c.lastStats[cts.CPU]
		if !ok {
			continue
		}
		lastTotal := totalCPUTime(lastCts)
		lastActive := activeCPUTime(lastCts)
		totalDelta := total - lastTotal

		if totalDelta < 0 {
			err = fmt.Errorf("current total CPU time is less than previous total CPU time")
			break
		}

		if totalDelta == 0 {
			continue
		}

		fieldsG := map[string]interface{}{
			"usage_user":       100 * (cts.User - lastCts.User - (cts.Guest - lastCts.Guest)) / totalDelta,
			"usage_system":     100 * (cts.System - lastCts.System) / totalDelta,
			"usage_idle":       100 * (cts.Idle - lastCts.Idle) / totalDelta,
			"usage_nice":       100 * (cts.Nice - lastCts.Nice - (cts.GuestNice - lastCts.GuestNice)) / totalDelta,
			"usage_iowait":     100 * (cts.Iowait - lastCts.Iowait) / totalDelta,
			"usage_irq":        100 * (cts.Irq - lastCts.Irq) / totalDelta,
			"usage_softirq":    100 * (cts.Softirq - lastCts.Softirq) / totalDelta,
			"usage_steal":      100 * (cts.Steal - lastCts.Steal) / totalDelta,
			"usage_guest":      100 * (cts.Guest - lastCts.Guest) / totalDelta,
			"usage_guest_nice": 100 * (cts.GuestNice - lastCts.GuestNice) / totalDelta,
		}
		if c.ReportActive {
			fieldsG["usage_active"] = 100 * (active - lastActive) / totalDelta
		}
		acc.AddGauge("cpu", fieldsG, tags, now)
	}

	c.lastStats = make(map[string]cpuUtil.TimesStat)
	for _, cts := range times {
		c.lastStats[cts.CPU] = cts
	}

	return err
}

func (c *CPUStats) Init() error {
	if c.CoreTags {
		cpuInfo, err := cpuUtil.Info()
		if err == nil {
			c.coreID = cpuInfo[0].CoreID != ""
			c.physicalID = cpuInfo[0].PhysicalID != ""

			c.cpuInfo = make(map[string]cpuUtil.InfoStat)
			for _, ci := range cpuInfo {
				c.cpuInfo[fmt.Sprintf("cpu%d", ci.CPU)] = ci
			}
		} else {
			c.Log.Warnf("Failed to gather info about CPUs: %s", err)
		}
	}

	return nil
}

func totalCPUTime(t cpuUtil.TimesStat) float64 {
	total := t.User + t.System + t.Nice + t.Iowait + t.Irq + t.Softirq + t.Steal + t.Idle
	return total
}

func activeCPUTime(t cpuUtil.TimesStat) float64 {
	active := totalCPUTime(t) - t.Idle
	return active
}

func init() {
	inputs.Add("cpu", func() telegraf.Input {
		return &CPUStats{
			PerCPU:   true,
			TotalCPU: true,
			ps:       system.NewSystemPS(),
		}
	})
}
