package procstat

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/process"
)

//NativeFinder uses gopsutil to find processes
type NativeFinder struct {
}

//NewNativeFinder ...
func NewNativeFinder() (PIDFinder, error) {
	return &NativeFinder{}, nil
}

//Uid will return all pids for the given user
func (pg *NativeFinder) Uid(user string) ([]PID, error) {
	var dst []PID
	procs, err := process.Processes()
	if err != nil {
		return dst, err
	}
	for _, p := range procs {
		username, err := p.Username()
		if err != nil {
			//skip, this can happen if we don't have permissions or
			//the pid no longer exists
			continue
		}
		if username == user {
			dst = append(dst, PID(p.Pid))
		}
	}
	return dst, nil
}

//PidFile returns the pid from the pid file given.
func (pg *NativeFinder) PidFile(path string) ([]PID, error) {
	var pids []PID
	pidString, err := ioutil.ReadFile(path)
	if err != nil {
		return pids, fmt.Errorf("Failed to read pidfile '%s'. Error: '%s'",
			path, err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidString)))
	if err != nil {
		return pids, err
	}
	pids = append(pids, PID(pid))
	return pids, nil

}
