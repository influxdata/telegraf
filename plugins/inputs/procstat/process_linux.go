package procstat

import (
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/process"
)

type Process interface {
	PID() PID
	Tags() map[string]string

	IOCounters() (*process.IOCountersStat, error)
	MemoryInfo() (*process.MemoryInfoStat, error)
	MemoryMaps(bool) (*[]process.MemoryMapsStat, error)
	Name() (string, error)
	NumCtxSwitches() (*process.NumCtxSwitchesStat, error)
	NumFDs() (int32, error)
	NumThreads() (int32, error)
	Percent(interval time.Duration) (float64, error)
	Times() (*cpu.TimesStat, error)
	RlimitUsage(bool) ([]process.RlimitStat, error)
	Username() (string, error)
}
