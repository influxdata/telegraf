package system

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
)

type SystemStats struct {
	Log telegraf.Logger
}

func (s *SystemStats) Gather(acc telegraf.Accumulator) error {
	loadavg, err := load.Avg()
	if err != nil && !strings.Contains(err.Error(), "not implemented") {
		return err
	}

	numCPUs, err := cpu.Counts(true)
	if err != nil {
		return err
	}

	fields := map[string]interface{}{
		"load1":  loadavg.Load1,
		"load5":  loadavg.Load5,
		"load15": loadavg.Load15,
		"n_cpus": numCPUs,
	}

	users, err := host.Users()
	if err == nil {
		fields["n_users"] = len(users)
	} else if os.IsNotExist(err) {
		s.Log.Debugf("Reading users: %s", err.Error())
	} else if os.IsPermission(err) {
		s.Log.Debug(err.Error())
	}

	now := time.Now()
	acc.AddGauge("system", fields, nil, now)

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
		// This will always succeed, so skip checking the error
		//nolint:errcheck,revive
		fmt.Fprintf(w, "%d day%s, ", days, s)
	}

	minutes := uptime / 60
	hours := minutes / 60
	hours %= 24
	minutes %= 60

	// This will always succeed, so skip checking the error
	//nolint:errcheck,revive
	fmt.Fprintf(w, "%2d:%02d", hours, minutes)

	// This will always succeed, so skip checking the error
	//nolint:errcheck,revive
	w.Flush()
	return buf.String()
}

func init() {
	inputs.Add("system", func() telegraf.Input {
		return &SystemStats{}
	})
}
