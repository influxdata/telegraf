package service

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/shirou/gopsutil/process"
	"github.com/stretchr/testify/require"
)

func TestGather(t *testing.T) {
	var mps MockPs
	var err error
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	memInfo := []*process.MemoryInfoStat{
		{
			RSS:  123,
			VMS:  456,
			Swap: 789,
		},
	}

	mps.On("MemInfo").Return(memInfo, nil)

	err = (&MemoryStats{ps: &mps, ProcessNames: []string{"foobar"}}).Gather(&acc)
	require.NoError(t, err)

	memFields := map[string]interface{}{
		"rss":  uint64(123),
		"vms":  uint64(456),
		"swap": uint64(789),
	}

	tags := map[string]string {
		"process_name": "foobar",
		"process_number": "0",
	}

	acc.AssertContainsTaggedFields(t, "service_mem", memFields, tags)
}
