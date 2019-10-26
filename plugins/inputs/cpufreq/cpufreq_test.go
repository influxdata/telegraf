package cpufreq

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestCPUFreq_NoThrottles(t *testing.T) {
	var acc testutil.Accumulator
	var cpufreq = &CPUFreq{
		PathSysfs: "testdata",
	}

	err := acc.GatherError(cpufreq.Gather)
	require.NoError(t, err)

	// CPU 0 Core 0
	acc.AssertContainsTaggedFields(t,
		"cpufreq",
		map[string]interface{}{
			"cur_freq": uint64(2597101000),
			"max_freq": uint64(3400000000),
			"min_freq": uint64(1200000000),
		},
		map[string]string{
			"cpu": "0",
		},
	)
	// CPU 1 Core 0
	acc.AssertContainsTaggedFields(t,
		"cpufreq",
		map[string]interface{}{
			"cur_freq": uint64(2597027000),
			"max_freq": uint64(3400000000),
			"min_freq": uint64(1200000000),
		},
		map[string]string{
			"cpu": "1",
		},
	)
	// CPU 0 Core 1
	acc.AssertContainsTaggedFields(t,
		"cpufreq",
		map[string]interface{}{
			"cur_freq": uint64(2597328000),
			"max_freq": uint64(3400000000),
			"min_freq": uint64(1200000000),
		},
		map[string]string{
			"cpu": "2",
		},
	)
	// CPU 1 Core 1
	acc.AssertContainsTaggedFields(t,
		"cpufreq",
		map[string]interface{}{
			"cur_freq": uint64(2597176000),
			"max_freq": uint64(3400000000),
			"min_freq": uint64(1200000000),
		},
		map[string]string{
			"cpu": "3",
		},
	)
}

func TestCPUFreq_WithThrottles(t *testing.T) {
	var acc testutil.Accumulator
	var cpufreq = &CPUFreq{
		PathSysfs:       "testdata",
		GatherThrottles: true,
	}

	err := acc.GatherError(cpufreq.Gather)
	require.NoError(t, err)

	// CPU 0 Core 0
	acc.AssertContainsTaggedFields(t,
		"cpufreq",
		map[string]interface{}{
			"cur_freq":       uint64(2597101000),
			"max_freq":       uint64(3400000000),
			"min_freq":       uint64(1200000000),
			"throttle_count": uint64(0),
		},
		map[string]string{
			"cpu": "0",
		},
	)
	// CPU 1 Core 0
	acc.AssertContainsTaggedFields(t,
		"cpufreq",
		map[string]interface{}{
			"cur_freq":       uint64(2597027000),
			"max_freq":       uint64(3400000000),
			"min_freq":       uint64(1200000000),
			"throttle_count": uint64(0),
		},
		map[string]string{
			"cpu": "1",
		},
	)
	// CPU 0 Core 1
	acc.AssertContainsTaggedFields(t,
		"cpufreq",
		map[string]interface{}{
			"cur_freq":       uint64(2597328000),
			"max_freq":       uint64(3400000000),
			"min_freq":       uint64(1200000000),
			"throttle_count": uint64(0),
		},
		map[string]string{
			"cpu": "2",
		},
	)
	// CPU 1 Core 1
	acc.AssertContainsTaggedFields(t,
		"cpufreq",
		map[string]interface{}{
			"cur_freq":       uint64(2597176000),
			"max_freq":       uint64(3400000000),
			"min_freq":       uint64(1200000000),
			"throttle_count": uint64(100),
		},
		map[string]string{
			"cpu": "3",
		},
	)
}
