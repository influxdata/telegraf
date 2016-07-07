package system

import (
	"fmt"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/shirou/gopsutil/cpu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCPUStats(t *testing.T) {
	var mps MockPS
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	cts := cpu.TimesStat{
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

	cts2 := cpu.TimesStat{
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

	mps.On("CPUTimes").Return([]cpu.TimesStat{cts}, nil)

	cs := NewCPUStats(&mps)

	cputags := map[string]string{
		"cpu": "cpu0",
	}

	err := cs.Gather(&acc)
	require.NoError(t, err)

	// Computed values are checked with delta > 0 becasue of floating point arithmatic
	// imprecision
	assertContainsTaggedFloat(t, &acc, "cpu", "time_user", 3.1, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_system", 8.2, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_idle", 80.1, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_nice", 1.3, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_iowait", 0.2, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_irq", 0.1, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_softirq", 0.11, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_steal", 0.0511, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_guest", 8.1, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_guest_nice", 0.324, 0, cputags)

	mps2 := MockPS{}
	mps2.On("CPUTimes").Return([]cpu.TimesStat{cts2}, nil)
	cs.ps = &mps2

	// Should have added cpu percentages too
	err = cs.Gather(&acc)
	require.NoError(t, err)

	assertContainsTaggedFloat(t, &acc, "cpu", "time_user", 11.4, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_system", 10.9, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_idle", 158.8699, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_nice", 2.5, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_iowait", 0.7, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_irq", 1.2, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_softirq", 0.31, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_steal", 0.2812, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_guest", 12.9, 0, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "time_guest_nice", 2.524, 0, cputags)

	assertContainsTaggedFloat(t, &acc, "cpu", "usage_user", 8.3, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_system", 2.7, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_idle", 78.7699, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_nice", 1.2, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_iowait", 0.5, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_irq", 1.1, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_softirq", 0.2, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_steal", 0.2301, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_guest", 4.8, 0.0005, cputags)
	assertContainsTaggedFloat(t, &acc, "cpu", "usage_guest_nice", 2.2, 0.0005, cputags)
}

// Asserts that a given accumulator contains a measurment of type float64 with
// specific tags within a certain distance of a given expected value. Asserts a failure
// if the measurement is of the wrong type, or if no matching measurements are found
//
// Paramaters:
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
