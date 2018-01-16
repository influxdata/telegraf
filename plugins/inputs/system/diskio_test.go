package system

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
