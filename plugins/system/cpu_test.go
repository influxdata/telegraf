package system

import (
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/shirou/gopsutil/cpu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCPUStats(t *testing.T) {
	var mps MockPS
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	cts := cpu.CPUTimesStat{
		CPU:       "cpu0",
		User:      3.1,
		System:    8.2,
		Idle:      80.1,
		Nice:      1.3,
		Iowait:    0.2,
		Irq:       0.1,
		Softirq:   0.11,
		Steal:     0.0511,
		Guest:     8.1,
		GuestNice: 0.324,
	}

	cts2 := cpu.CPUTimesStat{
		CPU:       "cpu0",
		User:      11.4,     // increased by 8.3
		System:    10.9,     // increased by 2.7
		Idle:      158.8699, // increased by 78.7699 (for total increase of 100)
		Nice:      2.5,      // increased by 1.2
		Iowait:    0.7,      // increased by 0.5
		Irq:       1.2,      // increased by 1.1
		Softirq:   0.31,     // increased by 0.2
		Steal:     0.2812,   // increased by 0.0001
		Guest:     12.9,     // increased by 4.8
		GuestNice: 2.524,    // increased by 2.2
	}

	mps.On("CPUTimes").Return([]cpu.CPUTimesStat{cts}, nil)

	cs := NewCPUStats(&mps)

	cputags := map[string]string{
		"cpu": "cpu0",
	}

	err := cs.Gather(&acc)
	require.NoError(t, err)
	numCPUPoints := len(acc.Points)

	expectedCPUPoints := 10
	assert.Equal(t, expectedCPUPoints, numCPUPoints)

	// Computed values are checked with delta > 0 becasue of floating point arithmatic
	// imprecision
	assertContainsTaggedFloat(t, &acc, "time_user", 3.1, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "time_system", 8.2, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "time_idle", 80.1, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "time_nice", 1.3, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "time_iowait", 0.2, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "time_irq", 0.1, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "time_softirq", 0.11, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "time_steal", 0.0511, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "time_guest", 8.1, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "time_guest_nice", 0.324, 0, cputags)

	mps2 := MockPS{}
	mps2.On("CPUTimes").Return([]cpu.CPUTimesStat{cts2}, nil)
	cs.ps = &mps2

	// Should have added cpu percentages too
	err = cs.Gather(&acc)
	require.NoError(t, err)

	numCPUPoints = len(acc.Points) - numCPUPoints
	expectedCPUPoints = 20
	assert.Equal(t, expectedCPUPoints, numCPUPoints)

	assertContainsTaggedFloat(t, &acc, "time_user", 11.4, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "time_system", 10.9, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "time_idle", 158.8699, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "time_nice", 2.5, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "time_iowait", 0.7, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "time_irq", 1.2, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "time_softirq", 0.31, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "time_steal", 0.2812, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "time_guest", 12.9, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "time_guest_nice", 2.524, 0, cputags)

	assertContainsTaggedFloat(t, &acc, "usage_user", 8.3, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "usage_system", 2.7, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "usage_idle", 78.7699, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "usage_nice", 1.2, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "usage_iowait", 0.5, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "usage_irq", 1.1, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "usage_softirq", 0.2, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "usage_steal", 0.2301, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "usage_guest", 4.8, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "usage_guest_nice", 2.2, 0.0005, cputags)
}
