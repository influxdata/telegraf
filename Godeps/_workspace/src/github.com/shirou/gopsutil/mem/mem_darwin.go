// +build darwin

package mem

import (
	"os/exec"
	"strconv"
	"strings"

	common "github.com/shirou/gopsutil/common"
)

func getPageSize() (uint64, error) {
	out, err := exec.Command("pagesize").Output()
	if err != nil {
		return 0, err
	}
	o := strings.TrimSpace(string(out))
	p, err := strconv.ParseUint(o, 10, 64)
	if err != nil {
		return 0, err
	}
	return p, nil
}

// Runs vm_stat and returns Free and inactive pages
func getVmStat(pagesize uint64, vms *VirtualMemoryStat) error {
	out, err := exec.Command("vm_stat").Output()
	if err != nil {
		return err
	}
	return parseVmStat(string(out), pagesize, vms)
}

func parseVmStat(out string, pagesize uint64, vms *VirtualMemoryStat) error {
	var err error

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.Trim(fields[1], " .")
		switch key {
		case "Pages free":
			free, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = e
			}
			vms.Free = free * pagesize
		case "Pages inactive":
			inactive, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = e
			}
			vms.Cached += inactive * pagesize
			vms.Inactive = inactive * pagesize
		case "Pages active":
			active, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = e
			}
			vms.Active = active * pagesize
		case "Pages wired down":
			wired, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = e
			}
			vms.Wired = wired * pagesize
		case "Pages purgeable":
			purgeable, e := strconv.ParseUint(value, 10, 64)
			if e != nil {
				err = e
			}
			vms.Cached += purgeable * pagesize
		}
	}
	return err
}

// VirtualMemory returns VirtualmemoryStat.
func VirtualMemory() (*VirtualMemoryStat, error) {
	ret := &VirtualMemoryStat{}

	p, err := getPageSize()
	if err != nil {
		return nil, err
	}
	t, err := common.DoSysctrl("hw.memsize")
	if err != nil {
		return nil, err
	}
	total, err := strconv.ParseUint(t[0], 10, 64)
	if err != nil {
		return nil, err
	}
	err = getVmStat(p, ret)
	if err != nil {
		return nil, err
	}

	ret.Available = ret.Free + ret.Cached
	ret.Total = total

	ret.Used = ret.Total - ret.Free
	ret.UsedPercent = float64(ret.Total-ret.Available) / float64(ret.Total) * 100.0

	return ret, nil
}

// SwapMemory returns swapinfo.
func SwapMemory() (*SwapMemoryStat, error) {
	var ret *SwapMemoryStat

	swapUsage, err := common.DoSysctrl("vm.swapusage")
	if err != nil {
		return ret, err
	}

	total := strings.Replace(swapUsage[2], "M", "", 1)
	used := strings.Replace(swapUsage[5], "M", "", 1)
	free := strings.Replace(swapUsage[8], "M", "", 1)

	total_v, err := strconv.ParseFloat(total, 64)
	if err != nil {
		return nil, err
	}
	used_v, err := strconv.ParseFloat(used, 64)
	if err != nil {
		return nil, err
	}
	free_v, err := strconv.ParseFloat(free, 64)
	if err != nil {
		return nil, err
	}

	u := float64(0)
	if total_v != 0 {
		u = ((total_v - free_v) / total_v) * 100.0
	}

	// vm.swapusage shows "M", multiply 1000
	ret = &SwapMemoryStat{
		Total:       uint64(total_v * 1000),
		Used:        uint64(used_v * 1000),
		Free:        uint64(free_v * 1000),
		UsedPercent: u,
	}

	return ret, nil
}
