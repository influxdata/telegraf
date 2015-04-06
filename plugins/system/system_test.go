package system

import (
	"testing"

	"github.com/influxdb/tivan/plugins/system/ps/cpu"
	"github.com/influxdb/tivan/plugins/system/ps/load"
	"github.com/influxdb/tivan/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemStats_GenerateStats(t *testing.T) {
	var mps MockPS

	defer mps.AssertExpectations(t)

	var acc testutil.Accumulator

	ss := &SystemStats{ps: &mps}

	lv := &load.LoadAvgStat{
		Load1:  0.3,
		Load5:  1.5,
		Load15: 0.8,
	}

	mps.On("LoadAvg").Return(lv, nil)

	cts := cpu.CPUTimesStat{
		CPU:       "all",
		User:      3.1,
		System:    8.2,
		Idle:      80.1,
		Nice:      1.3,
		Iowait:    0.2,
		Irq:       0.1,
		Softirq:   0.11,
		Steal:     0.0001,
		Guest:     8.1,
		GuestNice: 0.324,
		Stolen:    0.051,
	}

	mps.On("CPUTimes").Return([]cpu.CPUTimesStat{cts}, nil)

	err := ss.Gather(&acc)
	require.NoError(t, err)

	assert.True(t, acc.CheckValue("load1", 0.3))
	assert.True(t, acc.CheckValue("load5", 1.5))
	assert.True(t, acc.CheckValue("load15", 0.8))

	assert.True(t, acc.CheckValue("all.user", 3.1))
	assert.True(t, acc.CheckValue("all.system", 8.2))
	assert.True(t, acc.CheckValue("all.idle", 80.1))
	assert.True(t, acc.CheckValue("all.nice", 1.3))
	assert.True(t, acc.CheckValue("all.iowait", 0.2))
	assert.True(t, acc.CheckValue("all.irq", 0.1))
	assert.True(t, acc.CheckValue("all.softirq", 0.11))
	assert.True(t, acc.CheckValue("all.steal", 0.0001))
	assert.True(t, acc.CheckValue("all.guest", 8.1))
	assert.True(t, acc.CheckValue("all.guestNice", 0.324))
	assert.True(t, acc.CheckValue("all.stolen", 0.051))
}
