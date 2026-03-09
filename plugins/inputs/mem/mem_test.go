package mem

import (
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/mem"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/psutil"
	"github.com/influxdata/telegraf/testutil"
)

func TestMemStats(t *testing.T) {
	var mps psutil.MockPS
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
		Sreclaimable:   1923022848,
		Sunreclaim:     157728768,
		SwapCached:     0,
		SwapFree:       524280,
		SwapTotal:      524280,
		VmallocChunk:   3872908,
		VmallocTotal:   3874808,
		VmallocUsed:    1416,
		WriteBack:      0,
		WriteBackTmp:   0,
	}

	mps.On("VMStat").Return(vms, nil)
	plugin := &Mem{ps: &mps}

	err = plugin.Init()
	require.NoError(t, err)

	plugin.platform = "linux"

	require.NoError(t, err)
	err = plugin.Gather(&acc)
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"mem",
			map[string]string{},
			map[string]interface{}{
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
				// "wired":             uint64(134),
				"slab":             uint64(1234),
				"commit_limit":     uint64(1),
				"committed_as":     uint64(118680),
				"dirty":            uint64(4),
				"high_free":        uint64(0),
				"high_total":       uint64(0),
				"huge_page_size":   uint64(4096),
				"huge_pages_free":  uint64(0),
				"huge_pages_total": uint64(0),
				"low_free":         uint64(69936),
				"low_total":        uint64(255908),
				"mapped":           uint64(42236),
				"page_tables":      uint64(1236),
				"shared":           uint64(0),
				"sreclaimable":     uint64(1923022848),
				"sunreclaim":       uint64(157728768),
				"swap_cached":      uint64(0),
				"swap_free":        uint64(524280),
				"swap_total":       uint64(524280),
				"vmalloc_chunk":    uint64(3872908),
				"vmalloc_total":    uint64(3874808),
				"vmalloc_used":     uint64(1416),
				"write_back":       uint64(0),
				"write_back_tmp":   uint64(0),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestMemStatsCollectExtended(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific extended memory test")
	}

	tests := []struct {
		name           string
		testdataDir    string
		extendedFields map[string]interface{}
	}{
		{
			name:        "normal",
			testdataDir: "normal",
			extendedFields: map[string]interface{}{
				"active_anon":   uint64(5765169152),
				"inactive_anon": uint64(1082245120),
				"active_file":   uint64(3535425536),
				"inactive_file": uint64(4421992448),
				"unevictable":   uint64(143360),
				"percpu":        uint64(5767168),
			},
		},
		{
			name:        "missing fields",
			testdataDir: "missing_fields",
			extendedFields: map[string]interface{}{
				"active_anon":   uint64(5765169152),
				"inactive_anon": uint64(1082245120),
				"active_file":   uint64(3535425536),
				"inactive_file": uint64(4421992448),
				"unevictable":   uint64(0),
				"percpu":        uint64(0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostProc, err := filepath.Abs(filepath.Join("testdata", tt.testdataDir, "proc"))
			require.NoError(t, err)
			t.Setenv("HOST_PROC", hostProc)

			var mps psutil.MockPS
			defer mps.AssertExpectations(t)
			var acc testutil.Accumulator

			vms := &mem.VirtualMemoryStat{
				Total:     12400,
				Available: 7600,
				Used:      5000,
				Free:      1235,
			}

			mps.On("VMStat").Return(vms, nil)
			plugin := &Mem{
				ps:              &mps,
				CollectExtended: true,
			}

			require.NoError(t, plugin.Init())
			plugin.platform = "linux"

			require.NoError(t, plugin.Gather(&acc))

			fields := acc.GetTelegrafMetrics()[0].Fields()
			for k, v := range tt.extendedFields {
				require.Equal(t, v, fields[k], "field %q mismatch", k)
			}
		})
	}
}

func TestMemStatsCollectExtendedDisabled(t *testing.T) {
	var mps psutil.MockPS
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	vms := &mem.VirtualMemoryStat{
		Total:     12400,
		Available: 7600,
		Used:      5000,
	}

	mps.On("VMStat").Return(vms, nil)
	plugin := &Mem{
		ps:              &mps,
		CollectExtended: false,
	}

	err := plugin.Init()
	require.NoError(t, err)

	plugin.platform = "linux"

	err = plugin.Gather(&acc)
	require.NoError(t, err)

	fields := acc.GetTelegrafMetrics()[0].Fields()
	for _, key := range []string{"active_anon", "inactive_anon", "active_file", "inactive_file", "unevictable", "percpu"} {
		require.Nil(t, fields[key], "field %q should not be present when collect_extended is false", key)
	}
}
