package systemd_units

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// SystemdUnits is a telegraf plugin to gather systemd unit status
type SystemdUnits struct {
	Timeout   internal.Duration
	UnitType  string `toml:"unittype"`
	systemctl systemctl
}

type systemctl func(Timeout internal.Duration, UnitType string) (*bytes.Buffer, error)

const measurement = "systemd_units"

// Below are mappings of systemd state tables as defined at
// https://github.com/systemd/systemd/blob/c87700a1335f489be31cd3549927da68b5638819/src/basic/unit-def.
// Duplicate strings are removed from this list, zero values are
// intentionally left free to catch string-not-found errors
// On lookup, the values are decreased, so that string-not-found maps to -1
var load_map = map[string]int{
	"loaded":      1,
	"stub":        2,
	"not-found":   3,
	"bad-setting": 4,
	"error":       5,
	"merged":      6,
	"masked":      7,
}

var active_map = map[string]int{
	"active":       1,
	"reloading":    2,
	"inactive":     3,
	"failed":       4,
	"activating":   5,
	"deactivating": 6,
}

var sub_map = map[string]int{
	// service_state_table, offset 0x0000
	"running":       0x0001,
	"dead":          0x0002,
	"start-pre":     0x0003,
	"start":         0x0004,
	"exited":        0x0005,
	"reload":        0x0006,
	"stop":          0x0007,
	"stop-watchdog": 0x0008,
	"stop-sigterm":  0x0009,
	"stop-sigkill":  0x000a,
	"stop-post":     0x000b,
	"final-sigterm": 0x000c,
	"failed":        0x000d,
	"auto-restart":  0x000e,

	// automount_state_table, offset 0x0010
	"waiting": 0x0011,

	// device_state_table, offset 0x0020
	"tentative": 0x0021,
	"plugged":   0x0022,

	// mount_state_table, offset 0x0030
	"mounting":           0x0031,
	"mounting-done":      0x0032,
	"mounted":            0x0033,
	"remounting":         0x0034,
	"unmounting":         0x0035,
	"remounting-sigterm": 0x0036,
	"remounting-sigkill": 0x0037,
	"unmounting-sigterm": 0x0038,
	"unmounting-sigkill": 0x0039,

	// path_state_table, offset 0x0040

	// scope_state_table, offset 0x0050
	"abandoned": 0x0051,

	// slice_state_table, offset 0x0060
	"active": 0x0061,

	// socket_state_table, offset 0x0070
	"start-chown":      0x0071,
	"start-post":       0x0072,
	"listening":        0x0073,
	"stop-pre":         0x0074,
	"stop-pre-sigterm": 0x0075,
	"stop-pre-sigkill": 0x0076,
	"final-sigkill":    0x0077,

	// swap_state_table, offset 0x0080
	"activating":           0x0081,
	"activating-done":      0x0082,
	"deactivating":         0x0083,
	"deactivating-sigterm": 0x0084,
	"deactivating-sigkill": 0x0085,

	// target_state_table, offset 0x0090

	// timer_state_table, offset 0x00a0
	"elapsed": 0x00a1,
}

var (
	defaultTimeout  = internal.Duration{Duration: time.Second}
	defaultUnitType = "service"
)

// Description returns a short description of the plugin
func (systemd_units *SystemdUnits) Description() string {
	return "Gather systemd units state"
}

// SampleConfig returns sample configuration options.
func (systemd_units *SystemdUnits) SampleConfig() string {
	return `
  ## Set timeout for systemctl execution
  # timeout = "1s"
  #
  ## Filter for a specific unit types, default is "service":
  # unittype = "service"
`
}

// Gather parses systemctl outputs and adds counters to the Accumulator
func (systemd_units *SystemdUnits) Gather(acc telegraf.Accumulator) error {
	out, err := systemd_units.systemctl(systemd_units.Timeout, systemd_units.UnitType)
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
		tags := map[string]string{
			"name": data[0],
		}
		load := load_map[data[1]] - 1
		active := active_map[data[2]] - 1
		sub := sub_map[data[3]] - 1

		fields := map[string]interface{}{
			"load":   load,
			"active": active,
			"sub":    sub,
		}
		acc.AddCounter(measurement, fields, tags)
	}

	return nil
}

func setSystemctl(Timeout internal.Duration, UnitType string) (*bytes.Buffer, error) {
	// is systemctl available ?
	systemctlPath, err := exec.LookPath("systemctl")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(systemctlPath, "list-units", "--all", fmt.Sprintf("--type=%s", UnitType), "--no-legend")

	var out bytes.Buffer
	cmd.Stdout = &out
	err = internal.RunTimeout(cmd, Timeout.Duration)
	if err != nil {
		return &out, fmt.Errorf("error running systemctl list-units --all --type=%s --no-legend: %s", UnitType, err)
	}

	return &out, nil
}

func init() {
	inputs.Add("systemd_units", func() telegraf.Input {
		return &SystemdUnits{
			systemctl: setSystemctl,
			Timeout:   defaultTimeout,
			UnitType:  defaultUnitType,
		}
	})
}
