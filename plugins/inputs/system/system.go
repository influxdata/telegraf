package system

import (
	"bufio"
	"bytes"
	"fmt"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type SystemStats struct{}

func (_ *SystemStats) Description() string {
	return "Read metrics about system load & uptime"
}

func (_ *SystemStats) SampleConfig() string { return "" }

func (_ *SystemStats) Gather(acc telegraf.Accumulator) error {
	loadavg, err := load.Avg()
	if err != nil && !strings.Contains(err.Error(), "not implemented") {
		return err
	}

	hostinfo, err := host.Info()
	if err != nil {
		return err
	}

	users, err := host.Users()
	if err != nil {
		return err
	}

	acc.AddGauge("system", map[string]interface{}{
		"load1":   loadavg.Load1,
		"load5":   loadavg.Load5,
		"load15":  loadavg.Load15,
		"n_users": len(users),
		"n_cpus":  runtime.NumCPU(),
	}, nil)
	acc.AddCounter("system", map[string]interface{}{
		"uptime":        hostinfo.Uptime,
		"uptime_format": format_uptime(hostinfo.Uptime),
	}, nil)

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
	inputs.Add("system", func() telegraf.Input {
		return &SystemStats{}
	})
}
