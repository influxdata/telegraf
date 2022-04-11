package systemd_units

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// SystemdUnits is a telegraf plugin to gather systemd unit status
type SystemdUnits struct {
	Timeout   config.Duration
	UnitType  string `toml:"unittype"`
	Pattern   string `toml:"pattern"`
	systemctl systemctl
}

type systemctl func(timeout config.Duration, unitType string, pattern string) (*bytes.Buffer, error)

const measurement = "systemd_units"

// Below are mappings of systemd state tables as defined in
// https://github.com/systemd/systemd/blob/c87700a1335f489be31cd3549927da68b5638819/src/basic/unit-def.c
// Duplicate strings are removed from this list.
var loadMap = map[string]int{
	"loaded":      0,
	"stub":        1,
	"not-found":   2,
	"bad-setting": 3,
	"error":       4,
	"merged":      5,
	"masked":      6,
}

var activeMap = map[string]int{
	"active":       0,
	"reloading":    1,
	"inactive":     2,
	"failed":       3,
	"activating":   4,
	"deactivating": 5,
}

var subMap = map[string]int{
	// service_state_table, offset 0x0000
	"running":       0x0000,
	"dead":          0x0001,
	"start-pre":     0x0002,
	"start":         0x0003,
	"exited":        0x0004,
	"reload":        0x0005,
	"stop":          0x0006,
	"stop-watchdog": 0x0007,
	"stop-sigterm":  0x0008,
	"stop-sigkill":  0x0009,
	"stop-post":     0x000a,
	"final-sigterm": 0x000b,
	"failed":        0x000c,
	"auto-restart":  0x000d,

	// automount_state_table, offset 0x0010
	"waiting": 0x0010,

	// device_state_table, offset 0x0020
	"tentative": 0x0020,
	"plugged":   0x0021,

	// mount_state_table, offset 0x0030
	"mounting":           0x0030,
	"mounting-done":      0x0031,
	"mounted":            0x0032,
	"remounting":         0x0033,
	"unmounting":         0x0034,
	"remounting-sigterm": 0x0035,
	"remounting-sigkill": 0x0036,
	"unmounting-sigterm": 0x0037,
	"unmounting-sigkill": 0x0038,

	// path_state_table, offset 0x0040

	// scope_state_table, offset 0x0050
	"abandoned": 0x0050,

	// slice_state_table, offset 0x0060
	"active": 0x0060,

	// socket_state_table, offset 0x0070
	"start-chown":      0x0070,
	"start-post":       0x0071,
	"listening":        0x0072,
	"stop-pre":         0x0073,
	"stop-pre-sigterm": 0x0074,
	"stop-pre-sigkill": 0x0075,
	"final-sigkill":    0x0076,

	// swap_state_table, offset 0x0080
	"activating":           0x0080,
	"activating-done":      0x0081,
	"deactivating":         0x0082,
	"deactivating-sigterm": 0x0083,
	"deactivating-sigkill": 0x0084,

	// target_state_table, offset 0x0090

	// timer_state_table, offset 0x00a0
	"elapsed": 0x00a0,
}

var (
	defaultTimeout  = config.Duration(time.Second)
	defaultUnitType = "service"
	defaultPattern  = ""
)

// Gather parses systemctl outputs and adds counters to the Accumulator
func (s *SystemdUnits) Gather(acc telegraf.Accumulator) error {
	out, err := s.systemctl(s.Timeout, s.UnitType, s.Pattern)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		line := scanner.Text()

		data := strings.Fields(line)
		if len(data) < 4 {
			acc.AddError(fmt.Errorf("Error parsing line (expected at least 4 fields): %s", line))
			continue
		}
		name := data[0]
		load := data[1]
		active := data[2]
		sub := data[3]
		tags := map[string]string{
			"name":   name,
			"load":   load,
			"active": active,
			"sub":    sub,
		}

		var (
			loadCode   int
			activeCode int
			subCode    int
			ok         bool
		)
		if loadCode, ok = loadMap[load]; !ok {
			acc.AddError(fmt.Errorf("Error parsing field 'load', value not in map: %s", load))
			continue
		}
		if activeCode, ok = activeMap[active]; !ok {
			acc.AddError(fmt.Errorf("Error parsing field 'active', value not in map: %s", active))
			continue
		}
		if subCode, ok = subMap[sub]; !ok {
			acc.AddError(fmt.Errorf("Error parsing field 'sub', value not in map: %s", sub))
			continue
		}
		fields := map[string]interface{}{
			"load_code":   loadCode,
			"active_code": activeCode,
			"sub_code":    subCode,
		}

		acc.AddFields(measurement, fields, tags)
	}

	return nil
}

func setSystemctl(timeout config.Duration, unitType string, pattern string) (*bytes.Buffer, error) {
	// is systemctl available ?
	systemctlPath, err := exec.LookPath("systemctl")
	if err != nil {
		return nil, err
	}
	// build parameters for systemctl call
	params := []string{"list-units"}
	// create patterns parameters if provided in config
	if pattern != "" {
		psplit := strings.SplitN(pattern, " ", -1)
		for v := range psplit {
			params = append(params, psplit[v])
		}
	}
	params = append(params, "--all", "--plain")
	// add type as configured in config
	params = append(params, fmt.Sprintf("--type=%s", unitType))
	params = append(params, "--no-legend")
	cmd := exec.Command(systemctlPath, params...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = internal.RunTimeout(cmd, time.Duration(timeout))
	if err != nil {
		return &out, fmt.Errorf("error running systemctl %s: %s", strings.Join(params, " "), err)
	}
	return &out, nil
}

func init() {
	inputs.Add("systemd_units", func() telegraf.Input {
		return &SystemdUnits{
			systemctl: setSystemctl,
			Timeout:   defaultTimeout,
			UnitType:  defaultUnitType,
			Pattern:   defaultPattern,
		}
	})
}
