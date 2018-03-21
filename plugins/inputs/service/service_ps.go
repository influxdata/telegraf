package service

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"time"
	"strings"

	"github.com/shirou/gopsutil/process"
)

type PS interface {
	MemInfo(processName string) ([]*process.MemoryInfoStat, error)
}

type servicePs struct{}

func (s *servicePs) MemInfo(processName string) ([]*process.MemoryInfoStat, error) {
	pids, err := getPidof(processName)
	if err != nil {
		return nil, fmt.Errorf("error getting process id for %s: %s", processName, err)
	}

	if len(pids) == 0 {
		return nil, fmt.Errorf("could not get pid for %s", processName)
	}

	var memoryInfo []*process.MemoryInfoStat

	for _, pid := range pids {
		pidInt, err := strconv.ParseInt(pid, 10, 32)
		if err != nil {
			return nil, err
		}

		p, err := process.NewProcess(int32(pidInt))
		if err != nil {
			return nil, err
		}

		info, err :=  p.MemoryInfo()
		if err != nil {
			return nil, err
		}

		memoryInfo = append(memoryInfo, info)
	}

	return memoryInfo, nil
}

func getPidof(processName string) ([]string, error) {
	c := exec.Command("pidof", processName)
	timeout := time.Duration(100000000)

	result, err := combinedOutputTimeout(c, timeout)
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.TrimSpace(result), " "), nil
}

// combinedOutputTimeout runs the given command with the given timeout and
// returns the combined output of stdout and stderr.
// If the command times out, it attempts to kill the process.
// copied from https://github.com/influxdata/telegraf
func combinedOutputTimeout(c *exec.Cmd, timeout time.Duration) (string, error) {
	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = &b
	if err := c.Start(); err != nil {
		return "", err
	}
	err := waitTimeout(c, timeout)
	return b.String(), err
}

// waitTimeout waits for the given command to finish with a timeout.
// It assumes the command has already been started.
// If the command times out, it attempts to kill the process.
// copied from https://github.com/influxdata/telegraf
func waitTimeout(c *exec.Cmd, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	done := make(chan error)
	go func() { done <- c.Wait() }()
	select {
	case err := <-done:
		timer.Stop()
		return err
	case <-timer.C:
		if err := c.Process.Kill(); err != nil {
			log.Printf("FATAL error killing process: %s", err)
			return err
		}
		// wait for the command to return after killing it
		<-done
		return fmt.Errorf("Command timed out after %s", timeout.String())
	}
}
