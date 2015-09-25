package system

import (
	"github.com/cloudfoundry/gosigar"

	"github.com/influxdb/telegraf/plugins"
)

type SystemStats struct{}

func (_ *SystemStats) Description() string {
	return "Read metrics about system load & uptime"
}

func (_ *SystemStats) SampleConfig() string { return "" }

func (_ *SystemStats) add(acc plugins.Accumulator,
	name string, val float64, tags map[string]string) {
	if val >= 0 {
		acc.Add(name, val, tags)
	}
}

func (_ *SystemStats) Gather(acc plugins.Accumulator) error {
	loadavg := sigar.LoadAverage{}
	if err := loadavg.Get(); err != nil {
		return err
	}

	uptime := sigar.Uptime{}
	if err := uptime.Get(); err != nil {
		return err
	}

	acc.Add("load1", loadavg.One, nil)
	acc.Add("load5", loadavg.Five, nil)
	acc.Add("load15", loadavg.Fifteen, nil)
	acc.Add("uptime", uptime.Length, nil)
	acc.Add("uptime_format", uptime.Format(), nil)

	return nil
}

func init() {
	plugins.Add("system", func() plugins.Plugin {
		return &SystemStats{}
	})
}
