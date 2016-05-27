package system

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/shirou/gopsutil/disk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiskStats(t *testing.T) {
	var mps MockPS
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator
	var err error

	duAll := []*disk.UsageStat{
		{
			Path:        "/",
			Fstype:      "ext4",
			Total:       128,
			Free:        23,
			Used:        100,
			InodesTotal: 1234,
			InodesFree:  234,
			InodesUsed:  1000,
		},
		{
			Path:        "/home",
			Fstype:      "ext4",
			Total:       256,
			Free:        46,
			Used:        200,
			InodesTotal: 2468,
			InodesFree:  468,
			InodesUsed:  2000,
		},
	}
	duFiltered := []*disk.UsageStat{
		{
			Path:        "/",
			Fstype:      "ext4",
			Total:       128,
			Free:        23,
			Used:        100,
			InodesTotal: 1234,
			InodesFree:  234,
			InodesUsed:  1000,
		},
	}

	mps.On("DiskUsage", []string(nil), []string(nil)).Return(duAll, nil)
	mps.On("DiskUsage", []string{"/", "/dev"}, []string(nil)).Return(duFiltered, nil)
	mps.On("DiskUsage", []string{"/", "/home"}, []string(nil)).Return(duAll, nil)

	err = (&DiskStats{ps: &mps}).Gather(&acc)
	require.NoError(t, err)

	numDiskMetrics := acc.NFields()
	expectedAllDiskMetrics := 14
	assert.Equal(t, expectedAllDiskMetrics, numDiskMetrics)

	tags1 := map[string]string{
		"path":   "/",
		"fstype": "ext4",
	}
	tags2 := map[string]string{
		"path":   "/home",
		"fstype": "ext4",
	}

	fields1 := map[string]interface{}{
		"total":        uint64(128),
		"used":         uint64(100),
		"free":         uint64(23),
		"inodes_total": uint64(1234),
		"inodes_free":  uint64(234),
		"inodes_used":  uint64(1000),
		"used_percent": float64(81.30081300813008),
	}
	fields2 := map[string]interface{}{
		"total":        uint64(256),
		"used":         uint64(200),
		"free":         uint64(46),
		"inodes_total": uint64(2468),
		"inodes_free":  uint64(468),
		"inodes_used":  uint64(2000),
		"used_percent": float64(81.30081300813008),
	}
	acc.AssertContainsTaggedFields(t, "disk", fields1, tags1)
	acc.AssertContainsTaggedFields(t, "disk", fields2, tags2)

	// We expect 6 more DiskMetrics to show up with an explicit match on "/"
	// and /home not matching the /dev in MountPoints
	err = (&DiskStats{ps: &mps, MountPoints: []string{"/", "/dev"}}).Gather(&acc)
	assert.Equal(t, expectedAllDiskMetrics+7, acc.NFields())

	// We should see all the diskpoints as MountPoints includes both
	// / and /home
	err = (&DiskStats{ps: &mps, MountPoints: []string{"/", "/home"}}).Gather(&acc)
	assert.Equal(t, 2*expectedAllDiskMetrics+7, acc.NFields())
}

// func TestDiskIOStats(t *testing.T) {
// 	var mps MockPS
// 	defer mps.AssertExpectations(t)
// 	var acc testutil.Accumulator
// 	var err error

// 	diskio1 := disk.IOCountersStat{
// 		ReadCount:    888,
// 		WriteCount:   5341,
// 		ReadBytes:    100000,
// 		WriteBytes:   200000,
// 		ReadTime:     7123,
// 		WriteTime:    9087,
// 		Name:         "sda1",
// 		IoTime:       123552,
// 		SerialNumber: "ab-123-ad",
// 	}
// 	diskio2 := disk.IOCountersStat{
// 		ReadCount:    444,
// 		WriteCount:   2341,
// 		ReadBytes:    200000,
// 		WriteBytes:   400000,
// 		ReadTime:     3123,
// 		WriteTime:    6087,
// 		Name:         "sdb1",
// 		IoTime:       246552,
// 		SerialNumber: "bb-123-ad",
// 	}

// 	mps.On("DiskIO").Return(
// 		map[string]disk.IOCountersStat{"sda1": diskio1, "sdb1": diskio2},
// 		nil)

// 	err = (&DiskIOStats{ps: &mps}).Gather(&acc)
// 	require.NoError(t, err)

// 	numDiskIOMetrics := acc.NFields()
// 	expectedAllDiskIOMetrics := 14
// 	assert.Equal(t, expectedAllDiskIOMetrics, numDiskIOMetrics)

// 	dtags1 := map[string]string{
// 		"name":   "sda1",
// 		"serial": "ab-123-ad",
// 	}
// 	dtags2 := map[string]string{
// 		"name":   "sdb1",
// 		"serial": "bb-123-ad",
// 	}

// 	assert.True(t, acc.CheckTaggedValue("reads", uint64(888), dtags1))
// 	assert.True(t, acc.CheckTaggedValue("writes", uint64(5341), dtags1))
// 	assert.True(t, acc.CheckTaggedValue("read_bytes", uint64(100000), dtags1))
// 	assert.True(t, acc.CheckTaggedValue("write_bytes", uint64(200000), dtags1))
// 	assert.True(t, acc.CheckTaggedValue("read_time", uint64(7123), dtags1))
// 	assert.True(t, acc.CheckTaggedValue("write_time", uint64(9087), dtags1))
// 	assert.True(t, acc.CheckTaggedValue("io_time", uint64(123552), dtags1))
// 	assert.True(t, acc.CheckTaggedValue("reads", uint64(444), dtags2))
// 	assert.True(t, acc.CheckTaggedValue("writes", uint64(2341), dtags2))
// 	assert.True(t, acc.CheckTaggedValue("read_bytes", uint64(200000), dtags2))
// 	assert.True(t, acc.CheckTaggedValue("write_bytes", uint64(400000), dtags2))
// 	assert.True(t, acc.CheckTaggedValue("read_time", uint64(3123), dtags2))
// 	assert.True(t, acc.CheckTaggedValue("write_time", uint64(6087), dtags2))
// 	assert.True(t, acc.CheckTaggedValue("io_time", uint64(246552), dtags2))

// 	// We expect 7 more DiskIOMetrics to show up with an explicit match on "sdb1"
// 	// and serial should be missing from the tags with SkipSerialNumber set
// 	err = (&DiskIOStats{ps: &mps, Devices: []string{"sdb1"}, SkipSerialNumber: true}).Gather(&acc)
// 	assert.Equal(t, expectedAllDiskIOMetrics+7, acc.NFields())

// 	dtags3 := map[string]string{
// 		"name": "sdb1",
// 	}

// 	assert.True(t, acc.CheckTaggedValue("reads", uint64(444), dtags3))
// 	assert.True(t, acc.CheckTaggedValue("writes", uint64(2341), dtags3))
// 	assert.True(t, acc.CheckTaggedValue("read_bytes", uint64(200000), dtags3))
// 	assert.True(t, acc.CheckTaggedValue("write_bytes", uint64(400000), dtags3))
// 	assert.True(t, acc.CheckTaggedValue("read_time", uint64(3123), dtags3))
// 	assert.True(t, acc.CheckTaggedValue("write_time", uint64(6087), dtags3))
// 	assert.True(t, acc.CheckTaggedValue("io_time", uint64(246552), dtags3))
// }
