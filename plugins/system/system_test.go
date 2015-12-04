package system

import (
	"fmt"
	"reflect"
	"syscall"
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemStats_GenerateStats(t *testing.T) {
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

	diskio1 := disk.DiskIOCountersStat{

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
	diskio2 := disk.DiskIOCountersStat{
		ReadCount:    444,
		WriteCount:   2341,
		ReadBytes:    200000,
		WriteBytes:   400000,
		ReadTime:     3123,
		WriteTime:    6087,
		Name:         "sdb1",
		IoTime:       246552,
		SerialNumber: "bb-123-ad",
	}

	mps.On("DiskIO").Return(map[string]disk.DiskIOCountersStat{"sda1": diskio1, "sdb1": diskio2}, nil)

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

	netprotos := []net.NetProtoCountersStat{
		net.NetProtoCountersStat{
			Protocol: "Udp",
			Stats: map[string]int64{
				"InDatagrams": 4655,
				"NoPorts":     892592,
			},
		},
	}
	mps.On("NetProto").Return(netprotos, nil)

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

	netstats := []net.NetConnectionStat{
		net.NetConnectionStat{
			Type: syscall.SOCK_DGRAM,
		},
		net.NetConnectionStat{
			Status: "ESTABLISHED",
		},
		net.NetConnectionStat{
			Status: "ESTABLISHED",
		},
		net.NetConnectionStat{
			Status: "CLOSE",
		},
	}

	mps.On("NetConnections").Return(netstats, nil)

	cs := NewCPUStats(&mps)

	cputags := map[string]string{
		"cpu": "cpu0",
	}

	preCPUPoints := len(acc.Points)
	err := cs.Gather(&acc)
	require.NoError(t, err)
	numCPUPoints := len(acc.Points) - preCPUPoints

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

	numCPUPoints = len(acc.Points) - (preCPUPoints + numCPUPoints)
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

	preDiskPoints := len(acc.Points)

	err = (&DiskStats{ps: &mps}).Gather(&acc)
	require.NoError(t, err)

	numDiskPoints := len(acc.Points) - preDiskPoints
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

	assert.True(t, acc.CheckTaggedValue("total", uint64(128), tags1))
	assert.True(t, acc.CheckTaggedValue("used", uint64(105), tags1))
	assert.True(t, acc.CheckTaggedValue("free", uint64(23), tags1))
	assert.True(t, acc.CheckTaggedValue("inodes_total", uint64(1234), tags1))
	assert.True(t, acc.CheckTaggedValue("inodes_free", uint64(234), tags1))
	assert.True(t, acc.CheckTaggedValue("inodes_used", uint64(1000), tags1))
	assert.True(t, acc.CheckTaggedValue("total", uint64(256), tags2))
	assert.True(t, acc.CheckTaggedValue("used", uint64(210), tags2))
	assert.True(t, acc.CheckTaggedValue("free", uint64(46), tags2))
	assert.True(t, acc.CheckTaggedValue("inodes_total", uint64(2468), tags2))
	assert.True(t, acc.CheckTaggedValue("inodes_free", uint64(468), tags2))
	assert.True(t, acc.CheckTaggedValue("inodes_used", uint64(2000), tags2))

	// We expect 6 more DiskPoints to show up with an explicit match on "/"
	// and /home not matching the /dev in Mountpoints
	err = (&DiskStats{ps: &mps, Mountpoints: []string{"/", "/dev"}}).Gather(&acc)
	assert.Equal(t, preDiskPoints+expectedAllDiskPoints+6, len(acc.Points))

	// We should see all the diskpoints as Mountpoints includes both
	// / and /home
	err = (&DiskStats{ps: &mps, Mountpoints: []string{"/", "/home"}}).Gather(&acc)
	assert.Equal(t, preDiskPoints+2*expectedAllDiskPoints+6, len(acc.Points))

	err = (&NetIOStats{ps: &mps, skipChecks: true}).Gather(&acc)
	require.NoError(t, err)

	ntags := map[string]string{
		"interface": "eth0",
	}

	assert.NoError(t, acc.ValidateTaggedValue("bytes_sent", uint64(1123), ntags))
	assert.NoError(t, acc.ValidateTaggedValue("bytes_recv", uint64(8734422), ntags))
	assert.NoError(t, acc.ValidateTaggedValue("packets_sent", uint64(781), ntags))
	assert.NoError(t, acc.ValidateTaggedValue("packets_recv", uint64(23456), ntags))
	assert.NoError(t, acc.ValidateTaggedValue("err_in", uint64(832), ntags))
	assert.NoError(t, acc.ValidateTaggedValue("err_out", uint64(8), ntags))
	assert.NoError(t, acc.ValidateTaggedValue("drop_in", uint64(7), ntags))
	assert.NoError(t, acc.ValidateTaggedValue("drop_out", uint64(1), ntags))
	assert.NoError(t, acc.ValidateValue("udp_noports", int64(892592)))
	assert.NoError(t, acc.ValidateValue("udp_indatagrams", int64(4655)))

	preDiskIOPoints := len(acc.Points)

	err = (&DiskIOStats{ps: &mps}).Gather(&acc)
	require.NoError(t, err)

	numDiskIOPoints := len(acc.Points) - preDiskIOPoints
	expectedAllDiskIOPoints := 14
	assert.Equal(t, expectedAllDiskIOPoints, numDiskIOPoints)

	dtags1 := map[string]string{
		"name":   "sda1",
		"serial": "ab-123-ad",
	}
	dtags2 := map[string]string{
		"name":   "sdb1",
		"serial": "bb-123-ad",
	}

	assert.True(t, acc.CheckTaggedValue("reads", uint64(888), dtags1))
	assert.True(t, acc.CheckTaggedValue("writes", uint64(5341), dtags1))
	assert.True(t, acc.CheckTaggedValue("read_bytes", uint64(100000), dtags1))
	assert.True(t, acc.CheckTaggedValue("write_bytes", uint64(200000), dtags1))
	assert.True(t, acc.CheckTaggedValue("read_time", uint64(7123), dtags1))
	assert.True(t, acc.CheckTaggedValue("write_time", uint64(9087), dtags1))
	assert.True(t, acc.CheckTaggedValue("io_time", uint64(123552), dtags1))
	assert.True(t, acc.CheckTaggedValue("reads", uint64(444), dtags2))
	assert.True(t, acc.CheckTaggedValue("writes", uint64(2341), dtags2))
	assert.True(t, acc.CheckTaggedValue("read_bytes", uint64(200000), dtags2))
	assert.True(t, acc.CheckTaggedValue("write_bytes", uint64(400000), dtags2))
	assert.True(t, acc.CheckTaggedValue("read_time", uint64(3123), dtags2))
	assert.True(t, acc.CheckTaggedValue("write_time", uint64(6087), dtags2))
	assert.True(t, acc.CheckTaggedValue("io_time", uint64(246552), dtags2))

	// We expect 7 more DiskIOPoints to show up with an explicit match on "sdb1"
	// and serial should be missing from the tags with SkipSerialNumber set
	err = (&DiskIOStats{ps: &mps, Devices: []string{"sdb1"}, SkipSerialNumber: true}).Gather(&acc)
	assert.Equal(t, preDiskIOPoints+expectedAllDiskIOPoints+7, len(acc.Points))

	dtags3 := map[string]string{
		"name": "sdb1",
	}

	assert.True(t, acc.CheckTaggedValue("reads", uint64(444), dtags3))
	assert.True(t, acc.CheckTaggedValue("writes", uint64(2341), dtags3))
	assert.True(t, acc.CheckTaggedValue("read_bytes", uint64(200000), dtags3))
	assert.True(t, acc.CheckTaggedValue("write_bytes", uint64(400000), dtags3))
	assert.True(t, acc.CheckTaggedValue("read_time", uint64(3123), dtags3))
	assert.True(t, acc.CheckTaggedValue("write_time", uint64(6087), dtags3))
	assert.True(t, acc.CheckTaggedValue("io_time", uint64(246552), dtags3))

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

	acc.Points = nil

	err = (&NetStats{&mps}).Gather(&acc)
	require.NoError(t, err)
	netstattags := map[string]string(nil)

	assert.NoError(t, acc.ValidateTaggedValue("tcp_established", 2, netstattags))
	assert.NoError(t, acc.ValidateTaggedValue("tcp_close", 1, netstattags))
	assert.NoError(t, acc.ValidateTaggedValue("udp_socket", 1, netstattags))

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
	expectedValue float64,
	delta float64,
	tags map[string]string,
) {
	var actualValue float64
	for _, pt := range acc.Points {
		if pt.Measurement == measurement {
			if (tags == nil) || reflect.DeepEqual(pt.Tags, tags) {
				if value, ok := pt.Fields["value"].(float64); ok {
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
	msg := fmt.Sprintf("Could not find measurement \"%s\" with requested tags within %f of %f, Actual: %f",
		measurement, delta, expectedValue, actualValue)
	assert.Fail(t, msg)
}
