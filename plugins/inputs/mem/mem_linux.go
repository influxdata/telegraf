//go:build linux

package mem

import (
	"github.com/shirou/gopsutil/v4/mem"
)

const extendedMemorySupported = true

func getExtendedMemoryFields() (map[string]any, error) {
	exVM, err := mem.NewExLinux().VirtualMemory()
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"active_file":   exVM.ActiveFile,
		"inactive_file": exVM.InactiveFile,
		"active_anon":   exVM.ActiveAnon,
		"inactive_anon": exVM.InactiveAnon,
		"unevictable":   exVM.Unevictable,
		"percpu":        exVM.Percpu,
	}, nil
}
