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

	du := []*disk.DiskUsageStat{
		{
			Path:        "/",
			Fstype:      "ext4",
			Total:       128,
			Free:        23,
			InodesTotal: 1234,
			InodesFree:  234,
		},
		{
			Path:        "/home",
			Fstype:      "ext4",
			Total:       256,
			Free:        46,
			InodesTotal: 2468,
			InodesFree:  468,
		},
	}

	mps.On("DiskUsage").Return(du, nil)

	err = (&DiskStats{ps: &mps}).Gather(&acc)
	require.NoError(t, err)

	numDiskPoints := acc.NFields()
	expectedAllDiskPoints := 12
	assert.Equal(t, expectedAllDiskPoints, numDiskPoints)

	tags1 := map[string]string{
		"path":   "/",
		"fstype": "ext4",
	}
	tags2 := map[string]string{
		"path":   "/home",
		"fstype": "ext4",
	}

	fields1 := map[string]interface{}{
		"total":        uint64(128),  //tags1)
		"used":         uint64(105),  //tags1)
		"free":         uint64(23),   //tags1)
		"inodes_total": uint64(1234), //tags1)
		"inodes_free":  uint64(234),  //tags1)
		"inodes_used":  uint64(1000), //tags1)
	}
	fields2 := map[string]interface{}{
		"total":        uint64(256),  //tags2)
		"used":         uint64(210),  //tags2)
		"free":         uint64(46),   //tags2)
		"inodes_total": uint64(2468), //tags2)
		"inodes_free":  uint64(468),  //tags2)
		"inodes_used":  uint64(2000), //tags2)
	}
	acc.AssertContainsTaggedFields(t, "disk", fields1, tags1)
	acc.AssertContainsTaggedFields(t, "disk", fields2, tags2)

	// We expect 6 more DiskPoints to show up with an explicit match on "/"
	// and /home not matching the /dev in Mountpoints
	err = (&DiskStats{ps: &mps, Mountpoints: []string{"/", "/dev"}}).Gather(&acc)
	assert.Equal(t, expectedAllDiskPoints+6, acc.NFields())

	// We should see all the diskpoints as Mountpoints includes both
	// / and /home
	err = (&DiskStats{ps: &mps, Mountpoints: []string{"/", "/home"}}).Gather(&acc)
	assert.Equal(t, 2*expectedAllDiskPoints+6, acc.NFields())
}

// func TestDiskIOStats(t *testing.T) {
// 	var mps MockPS
// 	defer mps.AssertExpectations(t)
// 	var acc testutil.Accumulator
// 	var err error

// 	diskio1 := disk.DiskIOCountersStat{
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
// 	diskio2 := disk.DiskIOCountersStat{
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
// 		map[string]disk.DiskIOCountersStat{"sda1": diskio1, "sdb1": diskio2},
// 		nil)

// 	err = (&DiskIOStats{ps: &mps}).Gather(&acc)
// 	require.NoError(t, err)

// 	numDiskIOPoints := acc.NFields()
// 	expectedAllDiskIOPoints := 14
// 	assert.Equal(t, expectedAllDiskIOPoints, numDiskIOPoints)

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

// 	// We expect 7 more DiskIOPoints to show up with an explicit match on "sdb1"
// 	// and serial should be missing from the tags with SkipSerialNumber set
// 	err = (&DiskIOStats{ps: &mps, Devices: []string{"sdb1"}, SkipSerialNumber: true}).Gather(&acc)
// 	assert.Equal(t, expectedAllDiskIOPoints+7, acc.NFields())

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
