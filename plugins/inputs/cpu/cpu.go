//go:generate ../../../tools/readme_config_includer/generator
package cpu

import (
	_ "embed"
	"encoding/json"
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
	lastStats  map[string]stats
	cpuInfo    map[string]cpu.InfoStat
	coreID     bool
	physicalID bool

	PerCPU         bool `toml:"percpu"`
	TotalCPU       bool `toml:"totalcpu"`
	CollectCPUTime bool `toml:"collect_cpu_time"`
	ReportActive   bool `toml:"report_active"`
	CoreTags       bool `toml:"core_tags"`

	Log telegraf.Logger `toml:"-"`
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
				c.Log.Tracef("CPU #%d: %+v", ci.CPU, ci)
			}
		} else {
			c.Log.Warnf("Failed to gather info about CPUs: %s", err)
		}
	}

	return nil
}

func (c *CPU) Gather(acc telegraf.Accumulator) error {
	statistics, err := cpuTimes(c.PerCPU, c.TotalCPU)
	if err != nil {
		return fmt.Errorf("error getting CPU info: %w", err)
	}
	now := time.Now()

	var report bool
	for _, s := range statistics {
		tags := map[string]string{
			"cpu": s.CPU,
		}
		if c.coreID {
			tags["core_id"] = c.cpuInfo[s.CPU].CoreID
		}
		if c.physicalID {
			tags["physical_id"] = c.cpuInfo[s.CPU].PhysicalID
		}

		// Add cpu time metrics
		if c.CollectCPUTime {
			fields := s.cycleFields(c.ReportActive)
			acc.AddCounter("cpu", fields, tags, now)
		}

		// Add in percentage
		if len(c.lastStats) == 0 {
			// If it's the 1st gather, can't get CPU Usage stats yet
			continue
		}

		last, ok := c.lastStats[s.CPU]
		if !ok {
			continue
		}

		var fields map[string]interface{}
		if fields, err = s.percentageFields(last, c.ReportActive); err != nil {
			// Break the loop here to update the last statistics cache for the
			// next cycle. The error will be propagated.
			report = true
			break
		} else if fields == nil {
			// We don't have a delta in time so we can't emit statistics here
			report = true
			continue
		}
		acc.AddGauge("cpu", fields, tags, now)

		// Set a report marker if any of the fields is invalid
		report = report || !valid(fields)
	}

	// Debug print the raw data for invalid values
	if c.Log.Level().Includes(telegraf.Trace) && report {
		c.Log.Trace("Detected invalid field values!")
		for _, s := range statistics {
			curr, err := json.Marshal(s)
			if err == nil {
				last, ok := c.lastStats[s.CPU]
				if ok {
					if prev, err := json.Marshal(last); err == nil {
						c.Log.Tracef("Invalid raw values %d CPU %q: %s; %s", now.UnixNano(), s.CPU, string(curr), string(prev))
					}
				} else {
					c.Log.Tracef("Invalid raw values %d CPU %q: %s; {}", now.UnixNano(), s.CPU, string(curr))
				}
			}
		}
	}

	// Update the last-value cache for computing the percentages
	c.lastStats = make(map[string]stats, len(statistics))
	for _, s := range statistics {
		c.lastStats[s.CPU] = s
	}

	return err
}

func valid(fields map[string]interface{}) bool {
	for _, raw := range fields {
		v := raw.(float64)
		if v < 0.0 || v > 100.0 {
			return false
		}
	}
	return true
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
