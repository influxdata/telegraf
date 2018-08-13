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
		"commit_limit":      vm.CommitLimit,
		"committed_as":      vm.CommittedAS,
		"dirty":             vm.Dirty,
		"high_free":         vm.HighFree,
		"high_total":        vm.HighTotal,
		"huge_page_size":    vm.HugePageSize,
		"huge_pages_free":   vm.HugePagesFree,
		"huge_pages_total":  vm.HugePagesTotal,
		"low_free":          vm.LowFree,
		"low_total":         vm.LowTotal,
		"mapped":            vm.Mapped,
		"page_tables":       vm.PageTables,
		"shared":            vm.Shared,
		"swap_cached":       vm.SwapCached,
		"swap_free":         vm.SwapFree,
		"swap_total":        vm.SwapTotal,
		"vmalloc_chunk":     vm.VMallocChunk,
		"vmalloc_total":     vm.VMallocTotal,
		"vmalloc_used":      vm.VMallocUsed,
		"write_back":        vm.Writeback,
		"write_back_tmp":    vm.WritebackTmp,
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
