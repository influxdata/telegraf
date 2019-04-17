package system

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
)

type SystemStats struct {
	SkipUsers  bool
	SkipUptime bool
}

var SystemsampleConfig = `

## Uncomment to remove deprecated metrics.
# fielddrop = ["uptime_format"]
## Uncomment the following line to disable uptime collection
# skip_uptime = false
## Uncomment the following line to disable uptime collection
# skip_users = false
`

func (_ *SystemStats) Description() string {
	return "Read metrics about system load & uptime"
}

func (_ *SystemStats) SampleConfig() string {
	return SystemsampleConfig
}

func (s *SystemStats) Gather(acc telegraf.Accumulator) error {
	loadavg, err := load.Avg()
	if err != nil && !strings.Contains(err.Error(), "not implemented") {
		return err
	}

	fields := map[string]interface{}{
		"load1":  loadavg.Load1,
		"load5":  loadavg.Load5,
		"load15": loadavg.Load15,
		"n_cpus": runtime.NumCPU(),
	}

	if !s.SkipUsers {
		users, err := host.Users()
		if err == nil {
			fields["n_users"] = len(users)
		} else if !os.IsPermission(err) {
			return err
		}
	}

	now := time.Now()
	acc.AddGauge("system", fields, nil, now)

	if !s.SkipUptime {
		uptime, err := host.Uptime()
		if err != nil {
			return err
		}

		acc.AddCounter("system", map[string]interface{}{
			"uptime": uptime,
		}, nil, now)
		acc.AddFields("system", map[string]interface{}{
			"uptime_format": formatUptime(uptime),
		}, nil, now)
	}

	return nil
}

func formatUptime(uptime uint64) string {
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
