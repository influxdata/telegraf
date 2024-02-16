//go:build linux

package systemd_units

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
)

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

type unitInfo struct {
	name           string
	state          dbus.UnitStatus
	properties     map[string]interface{}
	unitFileState  string
	unitFilePreset string
}

type client interface {
	Connected() bool
	Close()

	ListUnitFilesByPatternsContext(ctx context.Context, states, pattern []string) ([]dbus.UnitFile, error)
	ListUnitsByNamesContext(ctx context.Context, units []string) ([]dbus.UnitStatus, error)
	GetUnitTypePropertiesContext(ctx context.Context, unit, unitType string) (map[string]interface{}, error)
	GetUnitPropertyContext(ctx context.Context, unit, propertyName string) (*dbus.Property, error)
	ListUnitsContext(ctx context.Context) ([]dbus.UnitStatus, error)
}

type archParams struct {
	client  client
	pattern []string
	filter  filter.Filter
}

func (s *SystemdUnits) Init() error {
	// Set default pattern
	if s.Pattern == "" {
		s.Pattern = "*"
	}

	// Check unit-type and convert the first letter to uppercase as this is
	// what dbus expects.
	switch s.UnitType {
	case "":
		s.UnitType = "service"
	case "service", "socket", "target", "device", "mount", "automount", "swap",
		"timer", "path", "slice", "scope":
	default:
		return fmt.Errorf("invalid 'unittype' %q", s.UnitType)
	}
	s.UnitType = strings.ToUpper(s.UnitType[0:1]) + strings.ToLower(s.UnitType[1:])

	// Check the sub-command
	switch s.SubCommand {
	case "":
		s.SubCommand = "list-units"
	case "list-units", "show":
	default:
		return fmt.Errorf("invalid 'subcommand' %q", s.SubCommand)
	}

	s.pattern = strings.Split(s.Pattern, " ")
	f, err := filter.Compile(s.pattern)
	if err != nil {
		return fmt.Errorf("compiling filter failed: %w", err)
	}
	s.filter = f

	return nil
}

func (s *SystemdUnits) Start(telegraf.Accumulator) error {
	ctx := context.Background()
	client, err := dbus.NewSystemConnectionContext(ctx)
	if err != nil {
		return err
	}
	s.client = client

	return nil
}

func (s *SystemdUnits) Stop() {
	if s.client != nil && s.client.Connected() {
		s.client.Close()
	}
}

func (s *SystemdUnits) Gather(acc telegraf.Accumulator) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.Timeout))
	defer cancel()

	// List all loaded units to handle multi-instance units correctly
	loaded, err := s.client.ListUnitsContext(ctx)
	if err != nil {
		return fmt.Errorf("listing loaded units failed: %w", err)
	}

	// List all unit files matching the pattern to also get disabled units
	list := []string{"enabled", "disabled", "static"}
	files, err := s.client.ListUnitFilesByPatternsContext(ctx, list, s.pattern)
	if err != nil {
		return fmt.Errorf("listing unit files failed: %w", err)
	}

	// Collect all matching units, the loaded ones and the disabled ones
	states := make([]dbus.UnitStatus, 0, len(files))

	// Match all loaded units first
	multiUnits := make(map[string]bool)
	for _, u := range loaded {
		if s.filter.Match(u.Name) {
			states = append(states, u)

			// Remember multi-instance units to remove duplicates from files
			if strings.Contains(u.Name, "@") {
				prefix, _, _ := strings.Cut(u.Name, "@")
				suffix := path.Ext(u.Name)
				multiUnits[prefix+"@"+suffix] = true
			}
		}
	}

	// Now split the unit-files into disabled ones and static ones, ignore
	// enabled units as those are already contained in the "loaded" list.
	disabled := make([]string, 0, len(files))
	static := make([]string, 0, len(files))
	for _, f := range files {
		name := path.Base(f.Path)

		switch f.Type {
		case "disabled":
			disabled = append(disabled, name)
		case "static":
			// Make sure we filter already loaded static multi-instance units
			if !strings.Contains(name, "@") {
				static = append(static, name)
			}
			prefix, _, _ := strings.Cut(name, "@")
			suffix := path.Ext(name)
			if !multiUnits[prefix+"@"+suffix] {
				static = append(static, name)
			}
			multiUnits[prefix+"@"+suffix] = true
		}
	}

	// Resolve the disabled and remaining static units
	disabledStates, err := s.client.ListUnitsByNamesContext(ctx, disabled)
	if err != nil {
		return fmt.Errorf("listing unit states failed: %w", err)
	}
	states = append(states, disabledStates...)

	// Merge the unit information into one struct
	units := make([]unitInfo, 0, len(states))
	for _, state := range states {
		// Filter units of the wrong type
		props, err := s.client.GetUnitTypePropertiesContext(ctx, state.Name, s.UnitType)
		if err != nil {
			continue
		}

		u := unitInfo{
			name:       state.Name,
			state:      state,
			properties: props,
		}

		// Get required unit file properties
		if v, err := s.client.GetUnitPropertyContext(ctx, state.Name, "UnitFileState"); err == nil {
			u.unitFileState = strings.Trim(v.Value.String(), `'"`)
		}
		if v, err := s.client.GetUnitPropertyContext(ctx, state.Name, "UnitFilePreset"); err == nil {
			u.unitFilePreset = strings.Trim(v.Value.String(), `'"`)
		}

		units = append(units, u)
	}

	// Add special information about unused static units
	for _, name := range static {
		if !strings.EqualFold(strings.TrimPrefix(path.Ext(name), "."), s.UnitType) {
			continue
		}

		units = append(units, unitInfo{
			name: name,
			state: dbus.UnitStatus{
				Name:        name,
				LoadState:   "stub",
				ActiveState: "inactive",
				SubState:    "dead",
			},
		})
	}

	// Create the metrics
	switch s.SubCommand {
	case "list-units":
		for _, u := range units {
			// Map the state names to numerical values
			load, ok := loadMap[u.state.LoadState]
			if !ok {
				acc.AddError(fmt.Errorf("parsing field 'load' failed, value not in map: %s", u.state.LoadState))
				continue
			}
			active, ok := activeMap[u.state.ActiveState]
			if !ok {
				acc.AddError(fmt.Errorf("parsing field field 'active' failed, value not in map: %s", u.state.ActiveState))
				continue
			}
			subState, ok := subMap[u.state.SubState]
			if !ok {
				acc.AddError(fmt.Errorf("parsing field field 'sub' failed, value not in map: %s", u.state.SubState))
				continue
			}

			// Create the metric
			tags := map[string]string{
				"name":   u.name,
				"load":   u.state.LoadState,
				"active": u.state.ActiveState,
				"sub":    u.state.SubState,
			}

			fields := map[string]interface{}{
				"load_code":   load,
				"active_code": active,
				"sub_code":    subState,
			}
			acc.AddFields("systemd_units", fields, tags)
		}
	case "show":
		for _, u := range units {
			// Map the state names to numerical values
			load, ok := loadMap[u.state.LoadState]
			if !ok {
				acc.AddError(fmt.Errorf("parsing field 'load' failed, value not in map: %s", u.state.LoadState))
				continue
			}
			active, ok := activeMap[u.state.ActiveState]
			if !ok {
				acc.AddError(fmt.Errorf("parsing field field 'active' failed, value not in map: %s", u.state.ActiveState))
				continue
			}
			subState, ok := subMap[u.state.SubState]
			if !ok {
				acc.AddError(fmt.Errorf("parsing field field 'sub' failed, value not in map: %s", u.state.SubState))
				continue
			}

			// Create the metric
			tags := map[string]string{
				"name":      u.name,
				"load":      u.state.LoadState,
				"active":    u.state.ActiveState,
				"sub":       u.state.SubState,
				"uf_state":  u.unitFileState,
				"uf_preset": u.unitFilePreset,
			}
			fields := map[string]interface{}{
				"load_code":    load,
				"active_code":  active,
				"sub_code":     subState,
				"status_errno": u.properties["StatusErrno"],
				"restarts":     u.properties["NRestarts"],
				"mem_current":  u.properties["MemoryCurrent"],
				"mem_peak":     u.properties["MemoryPeak"],
				"swap_current": u.properties["MemorySwapCurrent"],
				"swap_peak":    u.properties["MemorySwapPeak"],
				"mem_avail":    u.properties["MemoryAvailable"],
				"pid":          u.properties["MainPID"],
			}
			acc.AddFields("systemd_units", fields, tags)
		}
	}

	return nil
}
