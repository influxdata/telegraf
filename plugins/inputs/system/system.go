//go:generate ../../../tools/readme_config_includer/generator
package system

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

var availableCollectors = []string{"load", "uptime"}

type System struct {
	Collect []string `toml:"collect"`

	Log telegraf.Logger `toml:"-"`

	collectLoad   bool
	collectUptime bool
}

func (*System) SampleConfig() string {
	return sampleConfig
}

func (s *System) Init() error {
	if len(s.Collect) == 0 {
		s.Collect = availableCollectors
	}
	if err := choice.CheckSlice(s.Collect, availableCollectors); err != nil {
		return fmt.Errorf("config option 'collect': %w", err)
	}

	s.collectLoad = choice.Contains("load", s.Collect)
	s.collectUptime = choice.Contains("uptime", s.Collect)

	return nil
}

func (s *System) Gather(acc telegraf.Accumulator) error {
	now := time.Now()

	if s.collectLoad {
		fields := make(map[string]interface{})

		loadavg, err := load.Avg()
		if err != nil {
			if !strings.Contains(err.Error(), "not implemented") {
				return err
			}
		} else {
			fields["load1"] = loadavg.Load1
			fields["load5"] = loadavg.Load5
			fields["load15"] = loadavg.Load15
		}

		numLogicalCPUs, err := cpu.Counts(true)
		if err != nil {
			return err
		}
		numPhysicalCPUs, err := cpu.Counts(false)
		if err != nil {
			return err
		}
		fields["n_cpus"] = numLogicalCPUs
		fields["n_physical_cpus"] = numPhysicalCPUs

		users, err := host.Users()
		if err == nil {
			fields["n_users"] = len(users)
			fields["n_unique_users"] = findUniqueUsers(users)
		} else if os.IsNotExist(err) {
			s.Log.Debugf("Reading users: %s", err.Error())
		} else if os.IsPermission(err) {
			s.Log.Debug(err.Error())
		}

		acc.AddGauge("system", fields, nil, now)
	}

	if s.collectUptime {
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

func findUniqueUsers(userStats []host.UserStat) int {
	uniqueUsers := make(map[string]bool)
	for _, userstat := range userStats {
		if _, ok := uniqueUsers[userstat.User]; !ok {
			uniqueUsers[userstat.User] = true
		}
	}
	return len(uniqueUsers)
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
		return &System{}
	})
}
