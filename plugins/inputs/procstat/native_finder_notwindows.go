// +build !windows

package procstat

import (
	"regexp"

	"github.com/shirou/gopsutil/process"
)

//Pattern matches on the process name
func (pg *NativeFinder) Pattern(pattern string) ([]PID, error) {
	var pids []PID
	regxPattern, err := regexp.Compile(pattern)
	if err != nil {
		return pids, err
	}
	procs, err := process.Processes()
	if err != nil {
		return pids, err
	}
	for _, p := range procs {
		name, err := p.Exe()
		if err != nil {
			//skip, this can be caused by the pid no longer existing
			//or you having no permissions to access it
			continue
		}
		if regxPattern.MatchString(name) {
			pids = append(pids, PID(p.Pid))
		}
	}
	return pids, err
}

//FullPattern matches on the command line when the process was executed
func (pg *NativeFinder) FullPattern(pattern string) ([]PID, error) {
	var pids []PID
	regxPattern, err := regexp.Compile(pattern)
	if err != nil {
		return pids, err
	}
	procs, err := process.Processes()
	if err != nil {
		return pids, err
	}
	for _, p := range procs {
		cmd, err := p.Cmdline()
		if err != nil {
			//skip, this can be caused by the pid no longer existing
			//or you having no permissions to access it
			continue
		}
		if regxPattern.MatchString(cmd) {
			pids = append(pids, PID(p.Pid))
		}
	}
	return pids, err
}
