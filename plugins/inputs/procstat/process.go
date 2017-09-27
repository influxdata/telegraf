package procstat

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/process"
)

type Process interface {
	PID() PID
	Tags() map[string]string

	IOCounters() (*process.IOCountersStat, error)
	MemoryInfo() (*process.MemoryInfoStat, error)
	Name() (string, error)
	NumCtxSwitches() (*process.NumCtxSwitchesStat, error)
	NumFDs() (int32, error)
	NumThreads() (int32, error)
	Percent(interval time.Duration) (float64, error)
	Times() (*cpu.TimesStat, error)
	RlimitUsage(bool) ([]process.RlimitStat, error)
}

type Proc struct {
	hasCPUTimes bool
	tags        map[string]string
	*process.Process
}

func NewProc(pid PID) (Process, error) {
	process, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, err
	}

	proc := &Proc{
		Process:     process,
		hasCPUTimes: false,
		tags:        make(map[string]string),
	}
	return proc, nil
}

func (p *Proc) Tags() map[string]string {
	return p.tags
}

func (p *Proc) PID() PID {
	return PID(p.Process.Pid)
}

func (p *Proc) Percent(interval time.Duration) (float64, error) {
	cpu_perc, err := p.Process.Percent(time.Duration(0))
	if !p.hasCPUTimes && err == nil {
		p.hasCPUTimes = true
		return 0, fmt.Errorf("Must call Percent twice to compute percent cpu.")
	}
	return cpu_perc, err
}
