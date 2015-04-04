package process

import (
	"encoding/json"
	"runtime"
	"time"

	cpu "github.com/shirou/gopsutil/cpu"
)

type Process struct {
	Pid            int32 `json:"pid"`
	name           string
	status         string
	numCtxSwitches *NumCtxSwitchesStat
	uids           []int32
	gids           []int32
	numThreads     int32
	memInfo        *MemoryInfoStat

	lastCPUTimes *cpu.CPUTimesStat
	lastCPUTime  time.Time
}

type OpenFilesStat struct {
	Path string `json:"path"`
	Fd   uint64 `json:"fd"`
}

type MemoryInfoStat struct {
	RSS  uint64 `json:"rss"`  // bytes
	VMS  uint64 `json:"vms"`  // bytes
	Swap uint64 `json:"swap"` // bytes
}

type RlimitStat struct {
	Resource int32 `json:"resource"`
	Soft     int32 `json:"soft"`
	Hard     int32 `json:"hard"`
}

type IOCountersStat struct {
	ReadCount  uint64 `json:"read_count"`
	WriteCount uint64 `json:"write_count"`
	ReadBytes  uint64 `json:"read_bytes"`
	WriteBytes uint64 `json:"write_bytes"`
}

type NumCtxSwitchesStat struct {
	Voluntary   int64 `json:"voluntary"`
	Involuntary int64 `json:"involuntary"`
}

func (p Process) String() string {
	s, _ := json.Marshal(p)
	return string(s)
}

func (o OpenFilesStat) String() string {
	s, _ := json.Marshal(o)
	return string(s)
}

func (m MemoryInfoStat) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}

func (r RlimitStat) String() string {
	s, _ := json.Marshal(r)
	return string(s)
}

func (i IOCountersStat) String() string {
	s, _ := json.Marshal(i)
	return string(s)
}

func (p NumCtxSwitchesStat) String() string {
	s, _ := json.Marshal(p)
	return string(s)
}

func PidExists(pid int32) (bool, error) {
	pids, err := Pids()
	if err != nil {
		return false, err
	}

	for _, i := range pids {
		if i == pid {
			return true, err
		}
	}

	return false, err
}

// If interval is 0, return difference from last call(non-blocking).
// If interval > 0, wait interval sec and return diffrence between start and end.
func (p *Process) CPUPercent(interval time.Duration) (float64, error) {
	numcpu := runtime.NumCPU()
	calculate := func(t1, t2 *cpu.CPUTimesStat, delta float64) float64 {
		if delta == 0 {
			return 0
		}
		delta_proc := (t2.User - t1.User) + (t2.System - t1.System)
		overall_percent := ((delta_proc / delta) * 100) * float64(numcpu)
		return overall_percent
	}

	cpuTimes, err := p.CPUTimes()
	if err != nil {
		return 0, err
	}

	if interval > 0 {
		p.lastCPUTimes = cpuTimes
		p.lastCPUTime = time.Now()
		time.Sleep(interval)
		cpuTimes, err = p.CPUTimes()
		if err != nil {
			return 0, err
		}
	} else {
		if p.lastCPUTimes == nil {
			// invoked first time
			p.lastCPUTimes, err = p.CPUTimes()
			if err != nil {
				return 0, err
			}
			p.lastCPUTime = time.Now()
			return 0, nil
		}
	}

	delta := (time.Now().Sub(p.lastCPUTime).Seconds()) * float64(numcpu)
	ret := calculate(p.lastCPUTimes, cpuTimes, float64(delta))
	p.lastCPUTimes = cpuTimes
	p.lastCPUTime = time.Now()
	return ret, nil
}
