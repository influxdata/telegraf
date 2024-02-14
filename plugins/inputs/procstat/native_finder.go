package procstat

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

// NativeFinder uses gopsutil to find processes
type NativeFinder struct {
	procs []*process.Process
}

// Uid will return all pids for the given user
func (pg *NativeFinder) Init() error {
	var err error
	pg.procs, err = process.Processes()
	if err != nil {
		return err
	}
	return nil
}

// Uid will return all pids for the given user
func (pg *NativeFinder) UID(user string) error {
	var final []*process.Process
	for _, p := range pg.procs {
		username, err := p.Username()
		if err != nil {
			//skip, this can happen if we don't have permissions or
			//the pid no longer exists
			continue
		}
		if username == user {
			final = append(final, p)
		}
	}
	pg.procs = final
	return nil
}

// PidFile returns the pid from the pid file given.
func (pg *NativeFinder) PidFile(path string) ([]PID, error) {
	var pids []PID
	pidString, err := os.ReadFile(path)
	if err != nil {
		return pids, fmt.Errorf("failed to read pidfile %q: %w", path, err)
	}
	pid, err := strconv.ParseInt(strings.TrimSpace(string(pidString)), 10, 32)
	if err != nil {
		return pids, err
	}
	pids = append(pids, PID(pid))
	return pids, nil
}

// FullPattern matches on the command line when the process was executed
func (pg *NativeFinder) FullPattern(pattern string) error {
	var final []*process.Process
	regxPattern, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	for _, p := range pg.procs {
		cmd, err := p.Cmdline()
		if err != nil {
			//skip, this can happen if we don't have permissions or
			//the pid no longer exists
			continue
		}
		if regxPattern.MatchString(cmd) {
			final = append(final, p)
		}
	}
	pg.procs = final
	return nil
}

// Children matches children pids on the command line when the process was executed
func (pg *NativeFinder) Children(pid PID) ([]PID, error) {
	// Get all running processes
	p, err := process.NewProcess(int32(pid))
	if err != nil {
		return nil, fmt.Errorf("getting process %d failed: %w", pid, err)
	}

	// Get all children of the current process
	children, err := p.Children()
	if err != nil {
		return nil, fmt.Errorf("unable to get children of process %d: %w", p.Pid, err)
	}
	pids := make([]PID, 0, len(children))
	for _, child := range children {
		pids = append(pids, PID(child.Pid))
	}

	return pids, err
}

// Pattern matches on the process name
func (pg *NativeFinder) Pattern(pattern string) error {
	var final []*process.Process
	regxPattern, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	for _, p := range pg.procs {
		name, err := processName(p)
		if err != nil {
			//skip, this can happen if we don't have permissions or
			//the pid no longer exists
			continue
		}
		if regxPattern.MatchString(name) {
			final = append(final, p)
		}
	}
	pg.procs = final
	return nil
}

// Pattern matches on the process name
func (pg *NativeFinder) GetResult() ([]PID, error) {
	var dst []PID
	for _, p := range pg.procs {
		dst = append(dst, PID(p.Pid))
	}
	return dst, nil
}
