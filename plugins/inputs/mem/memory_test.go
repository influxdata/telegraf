package mem

import (
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/system"
	"github.com/influxdata/telegraf/testutil"
	"github.com/shirou/gopsutil/mem"
	"github.com/stretchr/testify/require"
)

func TestMemStats(t *testing.T) {
	var mps system.MockPS
	var err error
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	vms := &mem.VirtualMemoryStat{
		Total:     12400,
		Available: 7600,
		Used:      5000,
		Free:      1235,
		Active:    8134,
		Inactive:  1124,
		Slab:      1234,
		Wired:     134,
		// Buffers:     771,
		// Cached:      4312,
		// Shared:      2142,
		CommitLimit:    1,
		CommittedAS:    118680,
		Dirty:          4,
		HighFree:       0,
		HighTotal:      0,
		HugePageSize:   4096,
		HugePagesFree:  0,
		HugePagesTotal: 0,
		LowFree:        69936,
		LowTotal:       255908,
		Mapped:         42236,
		PageTables:     1236,
		Shared:         0,
		SwapCached:     0,
		SwapFree:       524280,
		SwapTotal:      524280,
		VMallocChunk:   3872908,
		VMallocTotal:   3874808,
		VMallocUsed:    1416,
		Writeback:      0,
		WritebackTmp:   0,
	}

	mps.On("VMStat").Return(vms, nil)

	err = (&MemStats{&mps}).Gather(&acc)
	require.NoError(t, err)

	memfields := map[string]interface{}{
		"total":             uint64(12400),
		"available":         uint64(7600),
		"used":              uint64(5000),
		"available_percent": float64(7600) / float64(12400) * 100,
		"used_percent":      float64(5000) / float64(12400) * 100,
		"free":              uint64(1235),
		"cached":            uint64(0),
		"buffered":          uint64(0),
		"active":            uint64(8134),
		"inactive":          uint64(1124),
		"wired":             uint64(134),
		"slab":              uint64(1234),
		"commit_limit":      uint64(1),
		"committed_as":      uint64(118680),
		"dirty":             uint64(4),
		"high_free":         uint64(0),
		"high_total":        uint64(0),
		"huge_page_size":    uint64(4096),
		"huge_pages_free":   uint64(0),
		"huge_pages_total":  uint64(0),
		"low_free":          uint64(69936),
		"low_total":         uint64(255908),
		"mapped":            uint64(42236),
		"page_tables":       uint64(1236),
		"shared":            uint64(0),
		"swap_cached":       uint64(0),
		"swap_free":         uint64(524280),
		"swap_total":        uint64(524280),
		"vmalloc_chunk":     uint64(3872908),
		"vmalloc_total":     uint64(3874808),
		"vmalloc_used":      uint64(1416),
		"write_back":        uint64(0),
		"write_back_tmp":    uint64(0),
	}
	acc.AssertContainsTaggedFields(t, "mem", memfields, make(map[string]string))

	acc.Metrics = nil
}
