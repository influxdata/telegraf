// +build freebsd

package cpu

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/internal/common"
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

var ClocksPerSec = float64(128)

func init() {
	out, err := exec.Command("/usr/bin/getconf", "CLK_TCK").Output()
	// ignore errors
	if err == nil {
		i, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
		if err == nil {
			ClocksPerSec = float64(i)
		}
	}
}

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
	filename := "/var/run/dmesg.boot"
	lines, _ := common.ReadLines(filename)

	var ret []CPUInfoStat

	c := CPUInfoStat{}
	for _, line := range lines {
		if matches := regexp.MustCompile(`CPU:\s+(.+) \(([\d.]+).+\)`).FindStringSubmatch(line); matches != nil {
			c.ModelName = matches[1]
			t, err := strconv.ParseFloat(matches[2], 64)
			if err != nil {
				return ret, nil
			}
			c.Mhz = t
		} else if matches := regexp.MustCompile(`Origin = "(.+)"  Id = (.+)  Family = (.+)  Model = (.+)  Stepping = (.+)`).FindStringSubmatch(line); matches != nil {
			c.VendorID = matches[1]
			c.Family = matches[3]
			c.Model = matches[4]
			t, err := strconv.ParseInt(matches[5], 10, 32)
			if err != nil {
				return ret, nil
			}
			c.Stepping = int32(t)
		} else if matches := regexp.MustCompile(`Features=.+<(.+)>`).FindStringSubmatch(line); matches != nil {
			for _, v := range strings.Split(matches[1], ",") {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if matches := regexp.MustCompile(`Features2=[a-f\dx]+<(.+)>`).FindStringSubmatch(line); matches != nil {
			for _, v := range strings.Split(matches[1], ",") {
				c.Flags = append(c.Flags, strings.ToLower(v))
			}
		} else if matches := regexp.MustCompile(`Logical CPUs per core: (\d+)`).FindStringSubmatch(line); matches != nil {
			// FIXME: no this line?
			t, err := strconv.ParseInt(matches[1], 10, 32)
			if err != nil {
				return ret, nil
			}
			c.Cores = int32(t)
		}

	}

	return append(ret, c), nil
}
