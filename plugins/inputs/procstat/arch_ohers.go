//go:build !linux && !windows

package procstat

import (
	"errors"

	"github.com/shirou/gopsutil/v3/process"
)

func processName(p *process.Process) (string, error) {
	return p.Exe()
}

func queryPidWithWinServiceName(_ string) (uint32, error) {
	return 0, errors.New("os not supporting win_service option")
}

func collectMemmap(Process, string, map[string]any) {}
