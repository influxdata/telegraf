package procstat

import (
	"regexp"
	"sync"

	"github.com/influxdata/telegraf/plugins/inputs/procstat/like2regexp"
	"github.com/shirou/gopsutil/process"
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

var patternCache = make(map[string]*regexp.Regexp)
var pcmut sync.RWMutex

func likeToRegexp(p string) (*regexp.Regexp, error) {
	pcmut.RLock()
	re, ok := patternCache[p]
	pcmut.RUnlock()
	if ok {
		return re, nil
	}

	pattern := like2regexp.WMILikeToRegexp(p)
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	pcmut.Lock()
	patternCache[p] = re
	pcmut.Unlock()
	return re, nil
}

//FullPattern matches on the command line when the process was executed
func (pg *NativeFinder) FullPattern(pattern string) ([]PID, error) {
	var pids []PID

	regxPattern, err := likeToRegexp(pattern)
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