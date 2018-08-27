package cpu

import (
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/system"
	"github.com/influxdata/telegraf/testutil"
	"github.com/shirou/gopsutil/cpu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCPUStats(t *testing.T) {
	var mps system.MockPS
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	cts := cpu.TimesStat{
		CPU:       "cpu0",
		User:      8.8,
		System:    8.2,
		Idle:      80.1,
		Nice:      1.3,
		Iowait:    0.8389,
		Irq:       0.6,
		Softirq:   0.11,
		Steal:     0.0511,
		Guest:     3.1,
		GuestNice: 0.324,
	}

	cts2 := cpu.TimesStat{
		CPU:       "cpu0",
		User:      24.9,     // increased by 16.1
		System:    10.9,     // increased by 2.7
		Idle:      157.9798, // increased by 77.8798 (for total increase of 100)
		Nice:      3.5,      // increased by 2.2
		Iowait:    0.929,    // increased by 0.0901
		Irq:       1.2,      // increased by 0.6
		Softirq:   0.31,     // increased by 0.2
		Steal:     0.2812,   // increased by 0.2301
		Guest:     11.4,     // increased by 8.3
		GuestNice: 2.524,    // increased by 2.2
	}

	mps.On("CPUTimes").Return([]cpu.TimesStat{cts}, nil)

	cs := NewCPUStats(&mps)

	cputags := map[string]string{
		"cpu": "cpu0",
	}

	err := cs.Gather(&acc)
	require.NoError(t, err)

	// Computed values are checked with delta > 0 because of floating point arithmatic
	// imprecision
	assertContainsTaggedFloat(t, &acc, "cpu", "time_user", 8.8, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_system", 8.2, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_idle", 80.1, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_active", 19.9, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_nice", 1.3, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_iowait", 0.8389, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_irq", 0.6, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_softirq", 0.11, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_steal", 0.0511, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_guest", 3.1, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_guest_nice", 0.324, 0, cputags)

	mps2 := system.MockPS{}
	mps2.On("CPUTimes").Return([]cpu.TimesStat{cts2}, nil)
	cs.ps = &mps2

	// Should have added cpu percentages too
	err = cs.Gather(&acc)
	require.NoError(t, err)

	assertContainsTaggedFloat(t, &acc, "cpu", "time_user", 24.9, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_system", 10.9, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_idle", 157.9798, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_active", 42.0202, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_nice", 3.5, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_iowait", 0.929, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_irq", 1.2, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_softirq", 0.31, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_steal", 0.2812, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_guest", 11.4, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_guest_nice", 2.524, 0, cputags)

	assertContainsTaggedFloat(t, &acc, "cpu", "usage_user", 7.8, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_system", 2.7, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_idle", 77.8798, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_active", 22.1202, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_nice", 0, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_iowait", 0.0901, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_irq", 0.6, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_softirq", 0.2, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_steal", 0.2301, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_guest", 8.3, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_guest_nice", 2.2, 0.0005, cputags)
}

// Asserts that a given accumulator contains a measurment of type float64 with
// specific tags within a certain distance of a given expected value. Asserts a failure
// if the measurement is of the wrong type, or if no matching measurements are found
//
// Parameters:
//     t *testing.T            : Testing object to use
//     acc testutil.Accumulator: Accumulator to examine
//     measurement string      : Name of the measurement to examine
//     expectedValue float64   : Value to search for within the measurement
//     delta float64           : Maximum acceptable distance of an accumulated value
//                               from the expectedValue parameter. Useful when
//                               floating-point arithmatic imprecision makes looking
//                               for an exact match impractical
//     tags map[string]string  : Tag set the found measurement must have. Set to nil to
//                               ignore the tag set.
func assertContainsTaggedFloat(
	t *testing.T,
	acc *testutil.Accumulator,
	measurement string,
	field string,
	expectedValue float64,
	delta float64,
	tags map[string]string,
) {
	var actualValue float64
	for _, pt := range acc.Metrics {
		if pt.Measurement == measurement {
			for fieldname, value := range pt.Fields {
				if fieldname == field {
					if value, ok := value.(float64); ok {
						actualValue = value
						if (value >= expectedValue-delta) && (value <= expectedValue+delta) {
							// Found the point, return without failing
							return
						}
					} else {
						assert.Fail(t, fmt.Sprintf("Measurement \"%s\" does not have type float64",
							measurement))
					}
				}
			}
		}
	}
	msg := fmt.Sprintf(
		"Could not find measurement \"%s\" with requested tags within %f of %f, Actual: %f",
		measurement, delta, expectedValue, actualValue)
	assert.Fail(t, msg)
}

// TestCPUCountChange tests that no errors are encountered if the number of
// CPUs increases as reported with LXC.
func TestCPUCountIncrease(t *testing.T) {
	var mps system.MockPS
	var mps2 system.MockPS
	var acc testutil.Accumulator
	var err error

	cs := NewCPUStats(&mps)

	mps.On("CPUTimes").Return(
		[]cpu.TimesStat{
			cpu.TimesStat{
				CPU: "cpu0",
			},
		}, nil)

	err = cs.Gather(&acc)
	require.NoError(t, err)

	mps2.On("CPUTimes").Return(
		[]cpu.TimesStat{
			cpu.TimesStat{
				CPU: "cpu0",
			},
			cpu.TimesStat{
				CPU: "cpu1",
			},
		}, nil)
	cs.ps = &mps2

	err = cs.Gather(&acc)
	require.NoError(t, err)
}

// TestCPUTimesDecrease tests that telegraf continue to works after
// CPU times decrease, which seems to occur when Linux system is suspended.
func TestCPUTimesDecrease(t *testing.T) {
	var mps system.MockPS
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	cts := cpu.TimesStat{
		CPU:    "cpu0",
		User:   18,
		Idle:   80,
		Iowait: 2,
	}

	cts2 := cpu.TimesStat{
		CPU:    "cpu0",
		User:   38, // increased by 20
		Idle:   40, // decreased by 40
		Iowait: 1,  // decreased by 1
	}

	cts3 := cpu.TimesStat{
		CPU:    "cpu0",
		User:   56,  // increased by 18
		Idle:   120, // increased by 80
		Iowait: 3,   // increased by 2
	}

	mps.On("CPUTimes").Return([]cpu.TimesStat{cts}, nil)

	cs := NewCPUStats(&mps)

	cputags := map[string]string{
		"cpu": "cpu0",
	}

	err := cs.Gather(&acc)
	require.NoError(t, err)

	// Computed values are checked with delta > 0 because of floating point arithmatic
	// imprecision
	assertContainsTaggedFloat(t, &acc, "cpu", "time_user", 18, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_idle", 80, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_iowait", 2, 0, cputags)

	mps2 := system.MockPS{}
	mps2.On("CPUTimes").Return([]cpu.TimesStat{cts2}, nil)
	cs.ps = &mps2

	// CPU times decreased. An error should be raised
	err = cs.Gather(&acc)
	require.Error(t, err)

	mps3 := system.MockPS{}
	mps3.On("CPUTimes").Return([]cpu.TimesStat{cts3}, nil)
	cs.ps = &mps3

	err = cs.Gather(&acc)
	require.NoError(t, err)

	assertContainsTaggedFloat(t, &acc, "cpu", "time_user", 56, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_idle", 120, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_iowait", 3, 0, cputags)

	assertContainsTaggedFloat(t, &acc, "cpu", "usage_user", 18, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_idle", 80, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_iowait", 2, 0.0005, cputags)
}
