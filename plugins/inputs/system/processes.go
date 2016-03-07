package system

import (
	"fmt"
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/shirou/gopsutil/process"
)

type Processes struct {
}

func (_ *Processes) Description() string {
	return "Get the number of processes and group them by status (Linux only)"
}

func (_ *Processes) SampleConfig() string { return "" }

func (s *Processes) Gather(acc telegraf.Accumulator) error {
	pids, err := process.Pids()
	if err != nil {
		return fmt.Errorf("error getting pids list: %s", err)
	}
	// TODO handle other OS (Windows/BSD/Solaris/OSX)
	fields := map[string]interface{}{
		"paging":   uint64(0),
		"blocked":  uint64(0),
		"zombie":   uint64(0),
		"stopped":  uint64(0),
		"running":  uint64(0),
		"sleeping": uint64(0),
	}
	for _, pid := range pids {
		process, err := process.NewProcess(pid)
		if err != nil {
			log.Printf("Can not get process %d status: %s", pid, err)
			continue
		}
		status, err := process.Status()
		if err != nil {
			log.Printf("Can not get process %d status: %s\n", pid, err)
			continue
		}
		_, exists := fields[status]
		if !exists {
			log.Printf("Status '%s' for process with pid: %d\n", status, pid)
			continue
		}
		fields[status] = fields[status].(uint64) + uint64(1)
	}

	acc.AddFields("processes", fields, nil)
	return nil
}
func init() {
	inputs.Add("processes", func() telegraf.Input {
		return &Processes{}
	})
}
