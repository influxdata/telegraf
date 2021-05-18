// +build !windows

package intel_rdt

import "github.com/prometheus/procfs"

type ProcessesHandler interface {
	getAllProcesses() ([]Process, error)
}

type Process struct {
	Name string
	PID  int
}

type ProcessManager struct{}

func NewProcessor() ProcessesHandler {
	return &ProcessManager{}
}

func (p *ProcessManager) getAllProcesses() ([]Process, error) {
	var processes []Process
	allProcesses, err := procfs.AllProcs()
	if err != nil {
		return nil, err
	}
	for _, proc := range allProcesses {
		procComm, err := proc.Comm()
		if err != nil {
			continue
		}
		newProcess := Process{
			PID:  proc.PID,
			Name: procComm,
		}
		processes = append(processes, newProcess)
	}
	return processes, nil
}
