//go:build linux

package mem

import (
	"github.com/shirou/gopsutil/v4/mem"
)

type linuxExtendedMemoryStats struct{}

func newExtendedMemoryStats() extendedMemoryStats {
	return &linuxExtendedMemoryStats{}
}

// getFields returns extended virtual memory statistics from /proc/meminfo.
func (*linuxExtendedMemoryStats) getFields() (map[string]interface{}, error) {
	exVM, err := mem.NewExLinux().VirtualMemory()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"active_file":   exVM.ActiveFile,
		"inactive_file": exVM.InactiveFile,
		"active_anon":   exVM.ActiveAnon,
		"inactive_anon": exVM.InactiveAnon,
		"unevictable":   exVM.Unevictable,
		"percpu":        exVM.Percpu,
	}, nil
}
