package procstat

import (
	"fmt"
	GOPS "github.com/keybase/go-ps"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type PIDFinder interface {
	PidFile(path string) ([]PID, error)
	Pattern(pattern string) ([]PID, error)
	Uid(user string) ([]PID, error)
	FullPattern(path string) ([]PID, error)
}

// Implemention of PIDGatherer that execs pgrep to find processes
type Pgrep struct {
	path string
}

func NewPgrep() (PIDFinder, error) {
	path, err := exec.LookPath("pgrep")
	if err != nil {
		return nil, fmt.Errorf("Could not find pgrep binary: %s", err)
	}
	return &Pgrep{path}, nil
}

func (pg *Pgrep) PidFile(path string) ([]PID, error) {
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

func (pg Pgrep) Pattern(pattern string) (pids []PID, err error) {
	pids = make([]PID, 0, 0)
	procs, err := GOPS.Processes()
	if err != nil {
		return pids, err
	}

	regx, err := regexp.Compile(pattern)
	if err != nil {
		return pids, err
	}
	for _, p := range procs {
		path, err := p.Path()
		if err != nil {
			// ignore errors, when this fails
			// it means you don't have the permissions to see the process
		}
		if regx.MatchString(path) {
			pids = append(pids, PID(p.Pid()))
		}
	}
	return pids, nil
}

func (pg Pgrep) FullPattern(pattern string) ([]PID, error) {
	return pg.Pattern(pattern)
}

func (pg *Pgrep) Uid(user string) ([]PID, error) {
	args := []string{"-u", user}
	return find(pg.path, args)
}

func find(path string, args []string) ([]PID, error) {
	out, err := run(path, args)
	if err != nil {
		return nil, err
	}

	return parseOutput(out)
}

func run(path string, args []string) (string, error) {
	out, err := exec.Command(path, args...).Output()
	if err != nil {
		return "", fmt.Errorf("Error running %s: %s", path, err)
	}
	return string(out), err
}

func parseOutput(out string) ([]PID, error) {
	pids := []PID{}
	fields := strings.Fields(out)
	for _, field := range fields {
		pid, err := strconv.Atoi(field)
		if err != nil {
			return nil, err
		}
		if err == nil {
			pids = append(pids, PID(pid))
		}
	}
	return pids, nil
}
