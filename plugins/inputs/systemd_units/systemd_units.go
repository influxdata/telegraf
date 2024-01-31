//go:generate ../../../tools/readme_config_includer/generator
package systemd_units

import (
	"bytes"
	_ "embed"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const measurement = "systemd_units"

// Below are mappings of systemd state tables as defined in
// https://github.com/systemd/systemd/blob/c87700a1335f489be31cd3549927da68b5638819/src/basic/unit-def.c
// Duplicate strings are removed from this list.
// This map is used by `subcommand_show` and `subcommand_list`. Changes must be
// compatible with both subcommands.
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
	"condition":     0x000e,
	"cleaning":      0x000f,

	// automount_state_table, offset 0x0010
	// continuation of service_state_table
	"waiting":                    0x0010,
	"reload-signal":              0x0011,
	"reload-notify":              0x0012,
	"final-watchdog":             0x0013,
	"dead-before-auto-restart":   0x0014,
	"failed-before-auto-restart": 0x0015,
	"dead-resources-pinned":      0x0016,
	"auto-restart-queued":        0x0017,

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

// SystemdUnits is a telegraf plugin to gather systemd unit status
type SystemdUnits struct {
	Timeout       config.Duration `toml:"timeout"`
	SubCommand    string          `toml:"subcommand"`
	UnitType      string          `toml:"unittype"`
	Pattern       string          `toml:"pattern"`
	resultParser  parseResultFunc
	commandParams *[]string
}

type getParametersFunc func(*SystemdUnits) *[]string
type parseResultFunc func(telegraf.Accumulator, *bytes.Buffer)

type subCommandInfo struct {
	getParameters getParametersFunc
	parseResult   parseResultFunc
}

var (
	defaultSubCommand = "list-units"
	defaultTimeout    = config.Duration(time.Second)
	defaultUnitType   = "service"
	defaultPattern    = ""
)

func (*SystemdUnits) SampleConfig() string {
	return sampleConfig
}

func (s *SystemdUnits) Init() error {
	var subCommandInfo *subCommandInfo

	switch s.SubCommand {
	case "show":
		subCommandInfo = initSubcommandShow()
	case "list-units":
		subCommandInfo = initSubcommandListUnits()
	default:
		return fmt.Errorf("invalid value for 'subcommand': %s", s.SubCommand)
	}

	// Save the parsing function for later and pre-compute the
	// command line because it will not change.
	s.resultParser = subCommandInfo.parseResult
	s.commandParams = subCommandInfo.getParameters(s)

	return nil
}

func (s *SystemdUnits) Gather(acc telegraf.Accumulator) error {
	// is systemctl available ?
	systemctlPath, err := exec.LookPath("systemctl")
	if err != nil {
		return err
	}

	cmd := exec.Command(systemctlPath, *s.commandParams...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = internal.RunTimeout(cmd, time.Duration(s.Timeout))
	if err != nil {
		return fmt.Errorf("error running systemctl %q: %w", strings.Join(*s.commandParams, " "), err)
	}

	s.resultParser(acc, &out)

	return nil
}

func init() {
	inputs.Add("systemd_units", func() telegraf.Input {
		return &SystemdUnits{
			Timeout:    defaultTimeout,
			UnitType:   defaultUnitType,
			Pattern:    defaultPattern,
			SubCommand: defaultSubCommand,
		}
	})
}
