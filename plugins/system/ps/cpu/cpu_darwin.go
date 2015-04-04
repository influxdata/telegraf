// +build darwin

package cpu

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	common "github.com/shirou/gopsutil/common"
)

// sys/resource.h
const (
	CPUser    = 0
	CPNice    = 1
	CPSys     = 2
	CPIntr    = 3
	CPIdle    = 4
	CPUStates = 5
)

// time.h
const (
	ClocksPerSec = 128
)

func CPUTimes(percpu bool) ([]CPUTimesStat, error) {
	var ret []CPUTimesStat

	var sysctlCall string
	var ncpu int
	if percpu {
		sysctlCall = "kern.cp_times"
		ncpu, _ = CPUCounts(true)
	} else {
		sysctlCall = "kern.cp_time"
		ncpu = 1
	}

	cpuTimes, err := common.DoSysctrl(sysctlCall)
	if err != nil {
		return ret, err
	}

	for i := 0; i < ncpu; i++ {
		offset := CPUStates * i
		user, err := strconv.ParseFloat(cpuTimes[CPUser+offset], 64)
		if err != nil {
			return ret, err
		}
		nice, err := strconv.ParseFloat(cpuTimes[CPNice+offset], 64)
		if err != nil {
			return ret, err
		}
		sys, err := strconv.ParseFloat(cpuTimes[CPSys+offset], 64)
		if err != nil {
			return ret, err
		}
		idle, err := strconv.ParseFloat(cpuTimes[CPIdle+offset], 64)
		if err != nil {
			return ret, err
		}
		intr, err := strconv.ParseFloat(cpuTimes[CPIntr+offset], 64)
		if err != nil {
			return ret, err
		}

		c := CPUTimesStat{
			User:   float64(user / ClocksPerSec),
			Nice:   float64(nice / ClocksPerSec),
			System: float64(sys / ClocksPerSec),
			Idle:   float64(idle / ClocksPerSec),
			Irq:    float64(intr / ClocksPerSec),
		}
		if !percpu {
			c.CPU = "cpu-total"
		} else {
			c.CPU = fmt.Sprintf("cpu%d", i)
		}

		ret = append(ret, c)
	}

	return ret, nil
}

// Returns only one CPUInfoStat on FreeBSD
func CPUInfo() ([]CPUInfoStat, error) {
	var ret []CPUInfoStat

	out, err := exec.Command("/usr/sbin/sysctl", "machdep.cpu").Output()
	if err != nil {
		return ret, err
	}

	c := CPUInfoStat{}
	for _, line := range strings.Split(string(out), "\n") {
		values := strings.Fields(line)
		if len(values) < 1 {
			continue
		}

		t, err := strconv.ParseInt(values[1], 10, 64)
		// err is not checked here because some value is string.
		if strings.HasPrefix(line, "machdep.cpu.brand_string") {
			c.ModelName = strings.Join(values[1:], " ")
		} else if strings.HasPrefix(line, "machdep.cpu.family") {
			c.Family = values[1]
		} else if strings.HasPrefix(line, "machdep.cpu.model") {
			c.Model = values[1]
		} else if strings.HasPrefix(line, "machdep.cpu.stepping") {
			if err != nil {
				return ret, err
			}
			c.Stepping = int32(t)
		} else if strings.HasPrefix(line, "machdep.cpu.features") {
			for _, v := range values[1:] {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if strings.HasPrefix(line, "machdep.cpu.leaf7_features") {
			for _, v := range values[1:] {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if strings.HasPrefix(line, "machdep.cpu.extfeatures") {
			for _, v := range values[1:] {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if strings.HasPrefix(line, "machdep.cpu.core_count") {
			if err != nil {
				return ret, err
			}
			c.Cores = int32(t)
		} else if strings.HasPrefix(line, "machdep.cpu.cache.size") {
			if err != nil {
				return ret, err
			}
			c.CacheSize = int32(t)
		} else if strings.HasPrefix(line, "machdep.cpu.vendor") {
			c.VendorID = values[1]
		}

		// TODO:
		// c.Mhz = mustParseFloat64(values[1])
	}

	return append(ret, c), nil
}
