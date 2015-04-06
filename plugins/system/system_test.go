package system

import (
	"testing"

	"github.com/influxdb/tivan/plugins/system/ps/cpu"
	"github.com/influxdb/tivan/plugins/system/ps/disk"
	"github.com/influxdb/tivan/plugins/system/ps/load"
	"github.com/influxdb/tivan/plugins/system/ps/net"
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

	du := &disk.DiskUsageStat{
		Path:        "/",
		Total:       128,
		Free:        23,
		InodesTotal: 1234,
		InodesFree:  234,
	}

	mps.On("DiskUsage").Return([]*disk.DiskUsageStat{du}, nil)

	diskio := disk.DiskIOCountersStat{
		ReadCount:    888,
		WriteCount:   5341,
		ReadBytes:    100000,
		WriteBytes:   200000,
		ReadTime:     7123,
		WriteTime:    9087,
		Name:         "sda1",
		IoTime:       123552,
		SerialNumber: "ab-123-ad",
	}

	mps.On("DiskIO").Return(map[string]disk.DiskIOCountersStat{"sda1": diskio}, nil)

	netio := net.NetIOCountersStat{
		Name:        "eth0",
		BytesSent:   1123,
		BytesRecv:   8734422,
		PacketsSent: 781,
		PacketsRecv: 23456,
		Errin:       832,
		Errout:      8,
		Dropin:      7,
		Dropout:     1,
	}

	mps.On("NetIO").Return([]net.NetIOCountersStat{netio}, nil)

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

	tags := map[string]string{
		"path": "/",
	}

	assert.True(t, acc.CheckTaggedValue("total", uint64(128), tags))
	assert.True(t, acc.CheckTaggedValue("used", uint64(105), tags))
	assert.True(t, acc.CheckTaggedValue("free", uint64(23), tags))
	assert.True(t, acc.CheckTaggedValue("inodes_total", uint64(1234), tags))
	assert.True(t, acc.CheckTaggedValue("inodes_free", uint64(234), tags))
	assert.True(t, acc.CheckTaggedValue("inodes_used", uint64(1000), tags))

	ntags := map[string]string{
		"interface": "eth0",
	}

	assert.True(t, acc.CheckTaggedValue("bytes_sent", uint64(1123), ntags))
	assert.True(t, acc.CheckTaggedValue("bytes_recv", uint64(8734422), ntags))
	assert.True(t, acc.CheckTaggedValue("packets_sent", uint64(781), ntags))
	assert.True(t, acc.CheckTaggedValue("packets_recv", uint64(23456), ntags))
	assert.True(t, acc.CheckTaggedValue("err_in", uint64(832), ntags))
	assert.True(t, acc.CheckTaggedValue("err_out", uint64(8), ntags))
	assert.True(t, acc.CheckTaggedValue("drop_in", uint64(7), ntags))
	assert.True(t, acc.CheckTaggedValue("drop_out", uint64(1), ntags))

	dtags := map[string]string{
		"name":   "sda1",
		"serial": "ab-123-ad",
	}

	assert.True(t, acc.CheckTaggedValue("reads", uint64(888), dtags))
	assert.True(t, acc.CheckTaggedValue("writes", uint64(5341), dtags))
	assert.True(t, acc.CheckTaggedValue("read_bytes", uint64(100000), dtags))
	assert.True(t, acc.CheckTaggedValue("write_bytes", uint64(200000), dtags))
	assert.True(t, acc.CheckTaggedValue("read_time", uint64(7123), dtags))
	assert.True(t, acc.CheckTaggedValue("write_time", uint64(9087), dtags))
	assert.True(t, acc.CheckTaggedValue("io_time", uint64(123552), dtags))
}
