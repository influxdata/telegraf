package system

import (
	"bufio"
	"bytes"
	"fmt"

	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"

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
	loadavg, err := load.LoadAvg()
	if err != nil {
		return err
	}

	hostinfo, err := host.HostInfo()
	if err != nil {
		return err
	}

	acc.Add("load1", loadavg.Load1, nil)
	acc.Add("load5", loadavg.Load5, nil)
	acc.Add("load15", loadavg.Load15, nil)
	acc.Add("uptime", hostinfo.Uptime, nil)
	acc.Add("uptime_format", format_uptime(hostinfo.Uptime), nil)

	return nil
}

func format_uptime(uptime uint64) string {
	buf := new(bytes.Buffer)
	w := bufio.NewWriter(buf)

	days := uptime / (60 * 60 * 24)

	if days != 0 {
		s := ""
		if days > 1 {
			s = "s"
		}
		fmt.Fprintf(w, "%d day%s, ", days, s)
	}

	minutes := uptime / 60
	hours := minutes / 60
	hours %= 24
	minutes %= 60

	fmt.Fprintf(w, "%2d:%02d", hours, minutes)

	w.Flush()
	return buf.String()
}

func init() {
	plugins.Add("system", func() plugins.Plugin {
		return &SystemStats{}
	})
}
