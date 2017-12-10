package procstat

import (
	"errors"

	"github.com/shirou/gopsutil/process"
)

var ErrorNotImplemented = errors.New("not implemented in windows")

// Implemention of PIDGatherer that execs pgrep to find processes
type Pgrep struct {
}

func NewPgrep() (PIDFinder, error) {
	return &Pgrep{}, nil
}

func (pg *Pgrep) PidFile(path string) ([]PID, error) {
	return nil, ErrorNotImplemented
}

func (pg *Pgrep) Pattern(pattern string) ([]PID, error) {
	pids := make([]PID, 0)
	procs, err := process.GetWin32ProcsByName(pattern)
	if err != nil {
		return pids, err
	}
	for _, p := range procs {
		pids = append(pids, PID(p.ProcessID))
	}
	return pids, nil
}

func (pg *Pgrep) FullPattern(pattern string) ([]PID, error) {
	pids := make([]PID, 0)
	procs, err := process.GetWin32ProcsByCmdLine(pattern)
	if err != nil {
		return pids, err
	}
	for _, p := range procs {
		pids = append(pids, PID(p.ProcessID))
	}
	return pids, nil
}

func (pg *Pgrep) Uid(user string) ([]PID, error) {
	return nil, ErrorNotImplemented
}
