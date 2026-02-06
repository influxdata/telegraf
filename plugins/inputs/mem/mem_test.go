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

func TestMemStatsBasic(t *testing.T) {
	var mps psutil.MockPS
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	vms := &mem.VirtualMemoryStat{
		Total:     8589934592,
		Available: 4294967296,
		Used:      4294967296,
	}

	mps.On("VMStat").Return(vms, nil)
	plugin := &Mem{ps: &mps}

	err := plugin.Init()
	require.NoError(t, err)

	// Use unknown platform to get only basic fields
	plugin.platform = "unknown"

	err = plugin.Gather(&acc)
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"mem",
			map[string]string{},
			map[string]interface{}{
				"total":             uint64(8589934592),
				"available":         uint64(4294967296),
				"used":              uint64(4294967296),
				"used_percent":      float64(4294967296) / float64(8589934592) * 100,
				"available_percent": float64(4294967296) / float64(8589934592) * 100,
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestMemStatsDarwin(t *testing.T) {
	var mps psutil.MockPS
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	vms := &mem.VirtualMemoryStat{
		Total:     17179869184,
		Available: 6653214720,
		Used:      10526654464,
		Free:      3221225472,
		Active:    4294967296,
		Inactive:  2147483648,
		Wired:     3758096384,
	}

	mps.On("VMStat").Return(vms, nil)
	plugin := &Mem{ps: &mps}

	err := plugin.Init()
	require.NoError(t, err)

	plugin.platform = "darwin"

	err = plugin.Gather(&acc)
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"mem",
			map[string]string{},
			map[string]interface{}{
				"total":             uint64(17179869184),
				"available":         uint64(6653214720),
				"used":              uint64(10526654464),
				"used_percent":      float64(10526654464) / float64(17179869184) * 100,
				"available_percent": float64(6653214720) / float64(17179869184) * 100,
				"free":              uint64(3221225472),
				"active":            uint64(4294967296),
				"inactive":          uint64(2147483648),
				"wired":             uint64(3758096384),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestMemStatsFreeBSD(t *testing.T) {
	var mps psutil.MockPS
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	vms := &mem.VirtualMemoryStat{
		Total:     8589934592,
		Available: 4294967296,
		Used:      4294967296,
		Free:      2147483648,
		Active:    1073741824,
		Inactive:  536870912,
		Buffers:   268435456,
		Cached:    1073741824,
		Wired:     536870912,
		Laundry:   134217728,
	}

	mps.On("VMStat").Return(vms, nil)
	plugin := &Mem{ps: &mps}

	err := plugin.Init()
	require.NoError(t, err)

	plugin.platform = "freebsd"

	err = plugin.Gather(&acc)
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"mem",
			map[string]string{},
			map[string]interface{}{
				"total":             uint64(8589934592),
				"available":         uint64(4294967296),
				"used":              uint64(4294967296),
				"used_percent":      float64(4294967296) / float64(8589934592) * 100,
				"available_percent": float64(4294967296) / float64(8589934592) * 100,
				"free":              uint64(2147483648),
				"active":            uint64(1073741824),
				"inactive":          uint64(536870912),
				"buffered":          uint64(268435456),
				"cached":            uint64(1073741824),
				"wired":             uint64(536870912),
				"laundry":           uint64(134217728),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestMemStatsOpenBSD(t *testing.T) {
	var mps psutil.MockPS
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	vms := &mem.VirtualMemoryStat{
		Total:     4294967296,
		Available: 2147483648,
		Used:      2147483648,
		Free:      1073741824,
		Active:    536870912,
		Inactive:  268435456,
		Cached:    536870912,
		Wired:     268435456,
	}

	mps.On("VMStat").Return(vms, nil)
	plugin := &Mem{ps: &mps}

	err := plugin.Init()
	require.NoError(t, err)

	plugin.platform = "openbsd"

	err = plugin.Gather(&acc)
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"mem",
			map[string]string{},
			map[string]interface{}{
				"total":             uint64(4294967296),
				"available":         uint64(2147483648),
				"used":              uint64(2147483648),
				"used_percent":      float64(2147483648) / float64(4294967296) * 100,
				"available_percent": float64(2147483648) / float64(4294967296) * 100,
				"free":              uint64(1073741824),
				"active":            uint64(536870912),
				"inactive":          uint64(268435456),
				"cached":            uint64(536870912),
				"wired":             uint64(268435456),
			},
			time.Unix(0, 0),
			telegraf.Gauge,
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}
