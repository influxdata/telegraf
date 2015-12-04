// +build darwin
// +build !cgo

package cpu

import "github.com/shirou/gopsutil/internal/common"

func perCPUTimes() ([]CPUTimesStat, error) {
	return []CPUTimesStat{}, common.NotImplementedError
}

func allCPUTimes() ([]CPUTimesStat, error) {
	return []CPUTimesStat{}, common.NotImplementedError
}
