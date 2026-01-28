//go:build linux

package mem

import (
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/mem"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/psutil"
	"github.com/influxdata/telegraf/testutil"
)

// mockExtendedMemoryStats implements extendedMemoryStats for testing.
type mockExtendedMemoryStats struct {
	fields map[string]interface{}
	err    error
}

func (m *mockExtendedMemoryStats) getFields() (map[string]interface{}, error) {
	return m.fields, m.err
}

func TestMemStatsLinux(t *testing.T) {
	var mps psutil.MockPS
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	vms := &mem.VirtualMemoryStat{
		Total:          16589934592,
		Available:      8294967296,
		Used:           8294967296,
		Free:           2147483648,
		Active:         4294967296,
		Inactive:       2147483648,
		Buffers:        536870912,
		Cached:         2147483648,
		Slab:           268435456,
		Shared:         134217728,
		Sreclaimable:   134217728,
		Sunreclaim:     67108864,
		CommitLimit:    8589934592,
		CommittedAS:    4294967296,
		Dirty:          1048576,
		HighFree:       0,
		HighTotal:      0,
		HugePageSize:   2097152,
		HugePagesFree:  0,
		HugePagesTotal: 0,
		LowFree:        2147483648,
		LowTotal:       16589934592,
		Mapped:         536870912,
		PageTables:     33554432,
		SwapCached:     0,
		SwapFree:       8589934592,
		SwapTotal:      8589934592,
		VmallocChunk:   0,
		VmallocTotal:   35184372088832,
		VmallocUsed:    67108864,
		WriteBack:      0,
		WriteBackTmp:   0,
	}

	mps.On("VMStat").Return(vms, nil)

	// Inject mock for extended memory stats
	mockExStats := &mockExtendedMemoryStats{
		fields: map[string]interface{}{
			"active_file":   uint64(1073741824),
			"inactive_file": uint64(2147483648),
			"active_anon":   uint64(536870912),
			"inactive_anon": uint64(268435456),
			"unevictable":   uint64(134217728),
			"percpu":        uint64(67108864),
		},
	}

	plugin := &Mem{
		ps:      &mps,
		exStats: mockExStats,
	}

	err := plugin.Init()
	require.NoError(t, err)

	plugin.platform = "linux"

	err = plugin.Gather(&acc)
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"mem",
			map[string]string{},
			map[string]interface{}{
				// Common fields
				"total":             uint64(16589934592),
				"available":         uint64(8294967296),
				"used":              uint64(8294967296),
				"used_percent":      float64(8294967296) / float64(16589934592) * 100,
				"available_percent": float64(8294967296) / float64(16589934592) * 100,
				// Linux-specific fields from VirtualMemoryStat
				"free":             uint64(2147483648),
				"buffered":         uint64(536870912),
				"cached":           uint64(2147483648),
				"active":           uint64(4294967296),
				"inactive":         uint64(2147483648),
				"slab":             uint64(268435456),
				"shared":           uint64(134217728),
				"sreclaimable":     uint64(134217728),
				"sunreclaim":       uint64(67108864),
				"commit_limit":     uint64(8589934592),
				"committed_as":     uint64(4294967296),
				"dirty":            uint64(1048576),
				"high_free":        uint64(0),
				"high_total":       uint64(0),
				"huge_page_size":   uint64(2097152),
				"huge_pages_free":  uint64(0),
				"huge_pages_total": uint64(0),
				"low_free":         uint64(2147483648),
				"low_total":        uint64(16589934592),
				"mapped":           uint64(536870912),
				"page_tables":      uint64(33554432),
				"swap_cached":      uint64(0),
				"swap_free":        uint64(8589934592),
				"swap_total":       uint64(8589934592),
				"vmalloc_chunk":    uint64(0),
				"vmalloc_total":    uint64(35184372088832),
				"vmalloc_used":     uint64(67108864),
				"write_back":       uint64(0),
				"write_back_tmp":   uint64(0),
				// Extended fields from ExVirtualMemory (mocked)
				"active_file":   uint64(1073741824),
				"inactive_file": uint64(2147483648),
				"active_anon":   uint64(536870912),
				"inactive_anon": uint64(268435456),
				"unevictable":   uint64(134217728),
				"percpu":        uint64(67108864),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}
