//go:build linux

package procstat

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/shirou/gopsutil/v3/process"
)

func processName(p *process.Process) (string, error) {
	return p.Exe()
}

func queryPidWithWinServiceName(_ string) (uint32, error) {
	return 0, errors.New("os not supporting win_service option")
}

func collectMemmap(proc Process, prefix string, fields map[string]any) {
	memMapStats, err := proc.MemoryMaps(true)
	if err == nil && len(*memMapStats) == 1 {
		memMap := (*memMapStats)[0]
		fields[prefix+"memory_size"] = memMap.Size
		fields[prefix+"memory_pss"] = memMap.Pss
		fields[prefix+"memory_shared_clean"] = memMap.SharedClean
		fields[prefix+"memory_shared_dirty"] = memMap.SharedDirty
		fields[prefix+"memory_private_clean"] = memMap.PrivateClean
		fields[prefix+"memory_private_dirty"] = memMap.PrivateDirty
		fields[prefix+"memory_referenced"] = memMap.Referenced
		fields[prefix+"memory_anonymous"] = memMap.Anonymous
		fields[prefix+"memory_swap"] = memMap.Swap
	}
}

func findBySystemdUnits(units []string) ([]processGroup, error) {
	ctx := context.Background()
	conn, err := dbus.NewSystemConnectionContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to systemd: %w", err)
	}
	defer conn.Close()

	sdunits, err := conn.ListUnitsByPatternsContext(ctx, []string{"enabled", "disabled", "static"}, units)
	if err != nil {
		return nil, fmt.Errorf("failed to list units: %w", err)
	}

	groups := make([]processGroup, 0, len(sdunits))
	for _, u := range sdunits {
		prop, err := conn.GetUnitTypePropertyContext(ctx, u.Name, "Service", "MainPID")
		if err != nil {
			// This unit might not be a service or similar
			continue
		}
		raw := prop.Value.Value()
		pid, ok := raw.(uint32)
		if !ok {
			return nil, fmt.Errorf("failed to parse PID %v of unit %q: invalid type %T", raw, u, raw)
		}
		p, err := process.NewProcess(int32(pid))
		if err != nil {
			return nil, fmt.Errorf("failed to find process for PID %d of unit %q: %w", pid, u, err)
		}
		groups = append(groups, processGroup{
			processes: []*process.Process{p},
			tags:      map[string]string{"systemd_unit": u.Name},
		})
	}

	return groups, nil
}

func findByWindowsServices(_ []string) ([]processGroup, error) {
	return nil, nil
}
