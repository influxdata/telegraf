package procstat

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf/internal"
)

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

func (pg *Pgrep) Pattern(pattern string) ([]PID, error) {
	args := []string{pattern}
	return find(pg.path, args)
}

func (pg *Pgrep) Uid(user string) ([]PID, error) {
	args := []string{"-u", user}
	return find(pg.path, args)
}

func (pg *Pgrep) FullPattern(pattern string) ([]PID, error) {
	args := []string{"-f", pattern}
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

	//if exit code 1, ie no processes found, do not return error
	if i, _ := internal.ExitStatus(err); i == 1 {
		return "", nil
	}

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
