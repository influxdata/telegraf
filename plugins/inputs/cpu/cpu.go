//go:generate ../../../tools/readme_config_includer/generator
package cpu

import (
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/psutil"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type CPU struct {
	ps         psutil.PS
	lastStats  map[string]cpu.TimesStat
	cpuInfo    map[string]cpu.InfoStat
	coreID     bool
	physicalID bool

	PerCPU         bool `toml:"percpu"`
	TotalCPU       bool `toml:"totalcpu"`
	CollectCPUTime bool `toml:"collect_cpu_time"`
	ReportActive   bool `toml:"report_active"`
	CoreTags       bool `toml:"core_tags"`
	ClampPercent   bool `toml:"clamp_percentages"`

	Log telegraf.Logger `toml:"-"`
}

func usagePercent(delta, totalDelta float64, clamp bool) float64 {
	if totalDelta <= 0 {
		return 0
	}

	if clamp {
		if delta < 0 {
			delta = 0
		} else if delta > totalDelta {
			delta = totalDelta
		}
	}

	return 100 * delta / totalDelta
}

func (*CPU) SampleConfig() string {
	return sampleConfig
}

func (c *CPU) Init() error {
	if c.CoreTags {
		cpuInfo, err := cpu.Info()
		if err == nil {
			c.coreID = cpuInfo[0].CoreID != ""
			c.physicalID = cpuInfo[0].PhysicalID != ""

			c.cpuInfo = make(map[string]cpu.InfoStat)
			for _, ci := range cpuInfo {
				c.cpuInfo[fmt.Sprintf("cpu%d", ci.CPU)] = ci
			}
		} else {
			c.Log.Warnf("Failed to gather info about CPUs: %s", err)
		}
	}

	return nil
}

func (c *CPU) Gather(acc telegraf.Accumulator) error {
	times, err := c.ps.CPUTimes(c.PerCPU, c.TotalCPU)
	if err != nil {
		return fmt.Errorf("error getting CPU info: %w", err)
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
			err = errors.New("current total CPU time is less than previous total CPU time")
			break
		}

		if totalDelta == 0 {
			continue
		}

		fieldsG := map[string]interface{}{}

		fieldsG["usage_user"] = usagePercent((cts.User-lastCts.User)-(cts.Guest-lastCts.Guest), totalDelta, c.ClampPercent)
		fieldsG["usage_system"] = usagePercent(cts.System-lastCts.System, totalDelta, c.ClampPercent)
		fieldsG["usage_idle"] = usagePercent(cts.Idle-lastCts.Idle, totalDelta, c.ClampPercent)
		fieldsG["usage_nice"] = usagePercent((cts.Nice-lastCts.Nice)-(cts.GuestNice-lastCts.GuestNice), totalDelta, c.ClampPercent)
		fieldsG["usage_iowait"] = usagePercent(cts.Iowait-lastCts.Iowait, totalDelta, c.ClampPercent)
		fieldsG["usage_irq"] = usagePercent(cts.Irq-lastCts.Irq, totalDelta, c.ClampPercent)
		fieldsG["usage_softirq"] = usagePercent(cts.Softirq-lastCts.Softirq, totalDelta, c.ClampPercent)
		fieldsG["usage_steal"] = usagePercent(cts.Steal-lastCts.Steal, totalDelta, c.ClampPercent)
		fieldsG["usage_guest"] = usagePercent(cts.Guest-lastCts.Guest, totalDelta, c.ClampPercent)
		fieldsG["usage_guest_nice"] = usagePercent(cts.GuestNice-lastCts.GuestNice, totalDelta, c.ClampPercent)
		if c.ReportActive {
			fieldsG["usage_active"] = usagePercent(active-lastActive, totalDelta, c.ClampPercent)
		}
		acc.AddGauge("cpu", fieldsG, tags, now)
	}

	c.lastStats = make(map[string]cpu.TimesStat)
	for _, cts := range times {
		c.lastStats[cts.CPU] = cts
	}

	return err
}

func totalCPUTime(t cpu.TimesStat) float64 {
	total := t.User + t.System + t.Nice + t.Iowait + t.Irq + t.Softirq + t.Steal + t.Idle
	return total
}

func activeCPUTime(t cpu.TimesStat) float64 {
	active := totalCPUTime(t) - t.Idle
	return active
}

func init() {
	inputs.Add("cpu", func() telegraf.Input {
		return &CPU{
			PerCPU:   true,
			TotalCPU: true,
			ps:       psutil.NewSystemPS(),
		}
	})
}
