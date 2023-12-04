package procstat

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf/internal"
)

// Implementation of PIDGatherer that execs pgrep to find processes
type Pgrep struct {
	path string
}

func newPgrepFinder() (PIDFinder, error) {
	path, err := exec.LookPath("pgrep")
	if err != nil {
		return nil, fmt.Errorf("could not find pgrep binary: %w", err)
	}
	return &Pgrep{path}, nil
}

func (pg *Pgrep) PidFile(path string) ([]PID, error) {
	var pids []PID
	pidString, err := os.ReadFile(path)
	if err != nil {
		return pids, fmt.Errorf("failed to read pidfile %q: %w",
			path, err)
	}
	pid, err := strconv.ParseInt(strings.TrimSpace(string(pidString)), 10, 32)
	if err != nil {
		return pids, err
	}
	pids = append(pids, PID(pid))
	return pids, nil
}

func (pg *Pgrep) Pattern(pattern string) ([]PID, error) {
	args := []string{pattern}
	return pg.find(args)
}

func (pg *Pgrep) UID(user string) ([]PID, error) {
	args := []string{"-u", user}
	return pg.find(args)
}

func (pg *Pgrep) FullPattern(pattern string) ([]PID, error) {
	args := []string{"-f", pattern}
	return pg.find(args)
}

func (pg *Pgrep) Children(pid PID) ([]PID, error) {
	args := []string{"-P", strconv.FormatInt(int64(pid), 10)}
	return pg.find(args)
}

func (pg *Pgrep) find(args []string) ([]PID, error) {
	// Execute pgrep with the given arguments
	buf, err := exec.Command(pg.path, args...).Output()
	if err != nil {
		// Exit code 1 means "no processes found" so we should not return
		// an error in this case.
		if status, _ := internal.ExitStatus(err); status == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("error running %q: %w", pg.path, err)
	}
	out := string(buf)

	// Parse the command output to extract the PIDs
	pids := []PID{}
	fields := strings.Fields(out)
	for _, field := range fields {
		pid, err := strconv.ParseInt(field, 10, 32)
		if err != nil {
			return nil, err
		}
		pids = append(pids, PID(pid))
	}
	return pids, nil
}
