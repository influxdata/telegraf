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

// VirtualMemory returns VirtualmemoryStat.
func VirtualMemory() (*VirtualMemoryStat, error) {
	p, err := getPageSize()
	if err != nil {
		return nil, err
	}

	total, err := common.DoSysctrl("hw.memsize")
	if err != nil {
		return nil, err
	}
	free, err := common.DoSysctrl("vm.page_free_count")
	if err != nil {
		return nil, err
	}
	parsed := make([]uint64, 0, 7)
	vv := []string{
		total[0],
		free[0],
	}
	for _, target := range vv {
		t, err := strconv.ParseUint(target, 10, 64)
		if err != nil {
			return nil, err
		}
		parsed = append(parsed, t)
	}

	ret := &VirtualMemoryStat{
		Total: parsed[0],
		Free:  parsed[1] * p,
	}

	// TODO: platform independent (worked freebsd?)
	ret.Available = ret.Free + ret.Buffers + ret.Cached

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
