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

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type SystemStats struct {
	// Publishes all metrics at once, opposed to three separate metrics
	// Enabling this will lose metric type support (used in outputs such as prometheus)
	MergeMetrics bool `toml:"merge_metrics"`

	Log telegraf.Logger
}

func (*SystemStats) SampleConfig() string {
	return sampleConfig
}

func (s *SystemStats) addFields(acc telegraf.Accumulator, loadFields map[string]interface{}, uptimeFields map[string]interface{}, uptimeFormatFields map[string]interface{}) {
	now := time.Now()

	if !s.MergeMetrics {
		acc.AddGauge("system", loadFields, nil, now)
		acc.AddCounter("system", uptimeFields, nil, now)
		acc.AddFields("system", uptimeFormatFields, nil, now)
		return
	}

	mergedFields := map[string]interface{}{}

	for k, v := range loadFields {
		mergedFields[k] = v
	}
	for k, v := range uptimeFields {
		mergedFields[k] = v
	}
	for k, v := range uptimeFormatFields {
		mergedFields[k] = v
	}

	acc.AddFields("system", mergedFields, nil, now)
}

func (s *SystemStats) Gather(acc telegraf.Accumulator) error {
	loadavg, err := load.Avg()
	if err != nil && !strings.Contains(err.Error(), "not implemented") {
		return err
	}

	loadFields := map[string]interface{}{}
	users, err := host.Users()
	if err == nil {
		loadFields["n_users"] = len(users)
	} else if os.IsNotExist(err) {
		s.Log.Debugf("Reading users: %s", err.Error())
	} else if os.IsPermission(err) {
		s.Log.Debug(err.Error())
	}

	numCPUs, err := cpu.Counts(true)
	if err != nil {
		return err
	}

	loadFields["load1"] = loadavg.Load1
	loadFields["load5"] = loadavg.Load5
	loadFields["load15"] = loadavg.Load15
	loadFields["n_cpus"] = numCPUs

	uptime, err := host.Uptime()
	if err != nil {
		return err
	}
	uptimeFields := map[string]interface{}{
		"uptime": uptime,
	}

	uptimeFormatFields := map[string]interface{}{
		"uptime_format": formatUptime(uptime),
	}

	s.addFields(acc, loadFields, uptimeFields, uptimeFormatFields)

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
		fmt.Fprintf(w, "%d day%s, ", days, s)
	}

	minutes := uptime / 60
	hours := minutes / 60
	hours %= 24
	minutes %= 60

	// This will always succeed, so skip checking the error
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
