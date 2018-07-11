package mem

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

type MemStats struct {
	ps system.PS
}

func (_ *MemStats) Description() string {
	return "Read metrics about memory usage"
}

func (_ *MemStats) SampleConfig() string { return "" }

func (s *MemStats) Gather(acc telegraf.Accumulator) error {
	vm, err := s.ps.VMStat()
	if err != nil {
		return fmt.Errorf("error getting virtual memory info: %s", err)
	}

	fields := map[string]interface{}{
		"total":             vm.Total,
		"available":         vm.Available,
		"used":              vm.Used,
		"free":              vm.Free,
		"cached":            vm.Cached,
		"buffered":          vm.Buffers,
		"active":            vm.Active,
		"inactive":          vm.Inactive,
		"wired":             vm.Wired,
		"slab":              vm.Slab,
		"used_percent":      100 * float64(vm.Used) / float64(vm.Total),
		"available_percent": 100 * float64(vm.Available) / float64(vm.Total),
	}
	acc.AddGauge("mem", fields, nil)

	return nil
}

func init() {
	ps := system.NewSystemPS()
	inputs.Add("mem", func() telegraf.Input {
		return &MemStats{ps: ps}
	})
}
