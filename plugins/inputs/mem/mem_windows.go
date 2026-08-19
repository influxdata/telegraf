//go:build windows

package mem

import (
	"github.com/shirou/gopsutil/v4/mem"
)

const extendedMemorySupported = true

func getExtendedMemoryFields() (map[string]interface{}, error) {
	exVM, err := mem.NewExWindows().VirtualMemory()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"commit_limit":    exVM.CommitLimit,
		"commit_total":    exVM.CommitTotal,
		"virtual_total":   exVM.VirtualTotal,
		"virtual_avail":   exVM.VirtualAvail,
		"phys_total":      exVM.PhysTotal,
		"phys_avail":      exVM.PhysAvail,
		"page_file_total": exVM.PageFileTotal,
		"page_file_avail": exVM.PageFileAvail,
	}, nil
}
