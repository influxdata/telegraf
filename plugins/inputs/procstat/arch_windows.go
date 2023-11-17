//go:build windows

package procstat

import "github.com/shirou/gopsutil/v3/process"

func processName(p *process.Process) (string, error) {
	return p.Name()
}
