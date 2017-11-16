package service

import (
	"strconv"

	"github.com/liquidm/telegraf"
	"github.com/liquidm/telegraf/plugins/inputs"
)

type MemoryStats struct {
	ps PS

	ProcessNames []string `toml:"process_names"`
}

func (_ *MemoryStats) Description() string {
	return "Read memory usage about a particular service"
}

var sampleConfig = `
  ## names of services
  ProcessNames = ["process"]
`

func (_ *MemoryStats) SampleConfig() string {
	return sampleConfig
}

func (s *MemoryStats) Gather(acc telegraf.Accumulator) error {

	for _, processName := range s.ProcessNames {
		memInfosForProcess, err := s.ps.MemInfo(processName)
		if err != nil {
			return err
		}

		for ii, memInfo := range memInfosForProcess {
			fields := map[string]interface{}{
				"rss":  memInfo.RSS,
				"vms":  memInfo.VMS,
				"swap": memInfo.Swap,
			}

			tags := map[string]string{
				"process_name":   processName,
				"process_number": strconv.Itoa(ii),
			}

			acc.AddGauge("service_mem", fields, tags)
		}
	}

	return nil
}

func init() {
	inputs.Add("service_mem", func() telegraf.Input {
		return &MemoryStats{ps: &servicePs{}}
	})
}
