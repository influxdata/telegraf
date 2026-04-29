//go:generate ../../../tools/readme_config_includer/generator
package system

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type System struct {
	Include []string        `toml:"include"`
	Log     telegraf.Logger `toml:"-"`
}

func (*System) SampleConfig() string {
	return sampleConfig
}

func (s *System) Init() error {
	// Suppress deprecation warnings for default-only configs.
	userSupplied := len(s.Include) > 0
	if !userSupplied {
		s.Include = []string{"load", "users", "legacy_cpus", "legacy_uptime"}
	}

	enabled := make(map[string]bool, len(s.Include))
	deduped := make([]string, 0, len(s.Include))
	for _, incl := range s.Include {
		if enabled[incl] {
			continue
		}
		switch incl {
		case "load", "users", "cpus", "uptime":
		case "legacy_cpus":
			if userSupplied {
				config.PrintOptionValueDeprecationNotice(
					"inputs.system",
					"include",
					"legacy_cpus",
					telegraf.DeprecationInfo{
						Since:     "1.39.0",
						RemovalIn: "1.45.0",
						Notice:    "use 'cpus' instead",
					},
				)
			}
		case "legacy_uptime":
			if userSupplied {
				config.PrintOptionValueDeprecationNotice(
					"inputs.system",
					"include",
					"legacy_uptime",
					telegraf.DeprecationInfo{
						Since:     "1.39.0",
						RemovalIn: "1.45.0",
						Notice:    "use 'uptime' instead",
					},
				)
			}
		default:
			return fmt.Errorf("invalid 'include' option %q", incl)
		}
		enabled[incl] = true
		deduped = append(deduped, incl)
	}
	s.Include = deduped

	if enabled["cpus"] && enabled["legacy_cpus"] {
		return errors.New(`"cpus" and "legacy_cpus" are mutually exclusive`)
	}
	if enabled["uptime"] && enabled["legacy_uptime"] {
		return errors.New(`"uptime" and "legacy_uptime" are mutually exclusive`)
	}

	return nil
}

func (s *System) Gather(acc telegraf.Accumulator) error {
	now := time.Now()
	fields := make(map[string]interface{}, 8)

	for _, incl := range s.Include {
		switch incl {
		case "load":
			loadavg, err := load.Avg()
			if err != nil {
				if !strings.Contains(err.Error(), "not implemented") {
					acc.AddError(fmt.Errorf("reading load averages: %w", err))
				}
				continue
			}
			fields["load1"] = loadavg.Load1
			fields["load5"] = loadavg.Load5
			fields["load15"] = loadavg.Load15
		case "users":
			users, err := host.Users()
			if err == nil {
				fields["n_users"] = len(users)
				fields["n_unique_users"] = findUniqueUsers(users)
			} else if os.IsNotExist(err) {
				s.Log.Debugf("Reading users: %s", err.Error())
			} else if os.IsPermission(err) {
				s.Log.Debug(err.Error())
			} else {
				s.Log.Warnf("Reading users: %s", err.Error())
			}
		case "cpus", "legacy_cpus":
			numLogicalCPUs, err := cpu.Counts(true)
			if err != nil {
				acc.AddError(fmt.Errorf("reading logical CPU count: %w", err))
				continue
			}
			numPhysicalCPUs, err := cpu.Counts(false)
			if err != nil {
				acc.AddError(fmt.Errorf("reading physical CPU count: %w", err))
				continue
			}
			if incl == "cpus" {
				fields["n_virtual_cpus"] = numLogicalCPUs
			} else {
				fields["n_cpus"] = numLogicalCPUs
			}
			fields["n_physical_cpus"] = numPhysicalCPUs
		case "uptime":
			uptime, err := host.Uptime()
			if err != nil {
				acc.AddError(fmt.Errorf("reading uptime: %w", err))
				continue
			}
			fields["uptime"] = uptime
		case "legacy_uptime":
			uptime, err := host.Uptime()
			if err != nil {
				acc.AddError(fmt.Errorf("reading uptime: %w", err))
				continue
			}
			acc.AddCounter("system", map[string]interface{}{
				"uptime": uptime,
			}, nil, now)
			acc.AddFields("system", map[string]interface{}{
				"uptime_format": formatUptime(uptime),
			}, nil, now)
		}
	}

	if len(fields) > 0 {
		acc.AddGauge("system", fields, nil, now)
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
