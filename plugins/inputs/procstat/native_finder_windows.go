package procstat

import (
	"regexp"
)

// Pattern matches on the process name
func (pg *NativeFinder) Pattern(pattern string) ([]PID, error) {
	var pids []PID
	regxPattern, err := regexp.Compile(pattern)
	if err != nil {
		return pids, err
	}
	procs, err := pg.FastProcessList()
	if err != nil {
		return pids, err
	}
	for _, p := range procs {
		name, err := p.Name()
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
