package system

import (
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/shirou/gopsutil/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStats(t *testing.T) {
	var mps MockPS
	var err error
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	vms := &mem.VirtualMemoryStat{
		Total:     12400,
		Available: 7600,
		Used:      5000,
		Free:      1235,
		// Active:      8134,
		// Inactive:    1124,
		// Buffers:     771,
		// Cached:      4312,
		// Wired:       134,
		// Shared:      2142,
	}

	mps.On("VMStat").Return(vms, nil)

	sms := &mem.SwapMemoryStat{
		Total:       8123,
		Used:        1232,
		Free:        6412,
		UsedPercent: 12.2,
		Sin:         7,
		Sout:        830,
	}

	mps.On("SwapStat").Return(sms, nil)

	err = (&MemStats{&mps}).Gather(&acc)
	require.NoError(t, err)

	vmtags := map[string]string(nil)

	assert.True(t, acc.CheckTaggedValue("total", uint64(12400), vmtags))
	assert.True(t, acc.CheckTaggedValue("available", uint64(7600), vmtags))
	assert.True(t, acc.CheckTaggedValue("used", uint64(5000), vmtags))
	assert.True(t, acc.CheckTaggedValue("available_percent",
		float64(7600)/float64(12400)*100,
		vmtags))
	assert.True(t, acc.CheckTaggedValue("used_percent",
		float64(5000)/float64(12400)*100,
		vmtags))
	assert.True(t, acc.CheckTaggedValue("free", uint64(1235), vmtags))

	acc.Points = nil

	err = (&SwapStats{&mps}).Gather(&acc)
	require.NoError(t, err)

	swaptags := map[string]string(nil)

	assert.NoError(t, acc.ValidateTaggedValue("total", uint64(8123), swaptags))
	assert.NoError(t, acc.ValidateTaggedValue("used", uint64(1232), swaptags))
	assert.NoError(t, acc.ValidateTaggedValue("used_percent", float64(12.2), swaptags))
	assert.NoError(t, acc.ValidateTaggedValue("free", uint64(6412), swaptags))
	assert.NoError(t, acc.ValidateTaggedValue("in", uint64(7), swaptags))
	assert.NoError(t, acc.ValidateTaggedValue("out", uint64(830), swaptags))
}
