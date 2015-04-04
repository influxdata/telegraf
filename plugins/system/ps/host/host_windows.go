// +build windows

package host

import (
	"fmt"
	"os"
	"strings"
	"time"

	common "github.com/shirou/gopsutil/common"
	process "github.com/shirou/gopsutil/process"
)

var (
	procGetSystemTimeAsFileTime = common.Modkernel32.NewProc("GetSystemTimeAsFileTime")
)

func HostInfo() (*HostInfoStat, error) {
	ret := &HostInfoStat{}
	hostname, err := os.Hostname()
	if err != nil {
		return ret, err
	}

	ret.Hostname = hostname
	uptime, err := BootTime()
	if err == nil {
		ret.Uptime = uptime
	}

	procs, err := process.Pids()
	if err != nil {
		return ret, err
	}

	ret.Procs = uint64(len(procs))

	return ret, nil
}

func BootTime() (uint64, error) {
	lines, err := common.GetWmic("os", "get", "LastBootUpTime")
	if err != nil {
		return 0, err
	}
	if len(lines) == 0 || len(lines[0]) != 2 {
		return 0, fmt.Errorf("could not get LastBootUpTime")
	}
	format := "20060102150405"
	t, err := time.Parse(format, strings.Split(lines[0][1], ".")[0])
	if err != nil {
		return 0, err
	}
	now := time.Now()
	return uint64(now.Sub(t).Seconds()), nil
}

func Users() ([]UserStat, error) {

	var ret []UserStat

	return ret, nil
}
