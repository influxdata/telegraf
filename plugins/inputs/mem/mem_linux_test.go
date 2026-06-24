//go:build linux

package mem

import (
	"maps"
	"path/filepath"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/mem"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/psutil"
	"github.com/influxdata/telegraf/testutil"
)

func TestMemStatsCollectExtended(t *testing.T) {
	baseFields := map[string]interface{}{
		"total":             uint64(12400),
		"available":         uint64(7600),
		"used":              uint64(5000),
		"used_percent":      100 * float64(5000) / float64(12400),
		"available_percent": 100 * float64(7600) / float64(12400),
		"free":              uint64(1235),
		"active":            uint64(0),
		"buffered":          uint64(0),
		"cached":            uint64(0),
		"commit_limit":      uint64(0),
		"committed_as":      uint64(0),
		"dirty":             uint64(0),
		"high_free":         uint64(0),
		"high_total":        uint64(0),
		"huge_pages_free":   uint64(0),
		"huge_page_size":    uint64(0),
		"huge_pages_total":  uint64(0),
		"inactive":          uint64(0),
		"low_free":          uint64(0),
		"low_total":         uint64(0),
		"mapped":            uint64(0),
		"page_tables":       uint64(0),
		"shared":            uint64(0),
		"slab":              uint64(0),
		"sreclaimable":      uint64(0),
		"sunreclaim":        uint64(0),
		"swap_cached":       uint64(0),
		"swap_free":         uint64(0),
		"swap_total":        uint64(0),
		"vmalloc_chunk":     uint64(0),
		"vmalloc_total":     uint64(0),
		"vmalloc_used":      uint64(0),
		"write_back_tmp":    uint64(0),
		"write_back":        uint64(0),
		"active_anon":       uint64(5765169152),
		"inactive_anon":     uint64(1082245120),
		"active_file":       uint64(3535425536),
		"inactive_file":     uint64(4421992448),
	}

	tests := []struct {
		name        string
		testdataDir string
		overrides   map[string]interface{}
	}{
		{
			name:        "normal",
			testdataDir: "normal",
			overrides: map[string]interface{}{
				"unevictable": uint64(143360),
				"percpu":      uint64(5767168),
			},
		},
		{
			name:        "missing fields",
			testdataDir: "missing_fields",
			overrides: map[string]interface{}{
				"unevictable": uint64(0),
				"percpu":      uint64(0),
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

			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))

			fields := maps.Clone(baseFields)
			maps.Copy(fields, tt.overrides)
			expected := []telegraf.Metric{
				metric.New("mem", map[string]string{}, fields, time.Unix(0, 0), telegraf.Gauge),
			}

			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}
