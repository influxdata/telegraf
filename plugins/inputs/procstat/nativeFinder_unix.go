// +build !windows

package procstat

func (pg *NativeFinder) Pattern(pattern string) ([]PID, error) {
	var pids []pid
	procs, err := GetWin32ProcsByName(pattern, Timeout)
	if err != nil {
		return pids, err
	}
	for _, p := range procs {
		pids = append(pids, PID(p.ProcessID))
	}
	return pids, nil
}

func (pg *NativeFinder) FullPattern(pattern string) ([]PID, error) {
	var pids []pid
	procs, err := GetWin32ProcsByCmdLine(pattern, Timeout)
	if err != nil {
		return pids, err
	}
	for _, p := range procs {
		pids = append(pids, PID(p.ProcessID))
	}
	return pids, nil
}

func (pg *NativeFinder) Uid(user string) ([]PID, error) {
	return nil, ErrorNotImplemented
}
