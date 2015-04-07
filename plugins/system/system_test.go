package system

import (
	"testing"

	"github.com/influxdb/tivan/plugins/system/ps/cpu"
	"github.com/influxdb/tivan/plugins/system/ps/disk"
	"github.com/influxdb/tivan/plugins/system/ps/docker"
	"github.com/influxdb/tivan/plugins/system/ps/load"
	"github.com/influxdb/tivan/plugins/system/ps/mem"
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
		CPU:       "cpu0",
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

	vms := &mem.VirtualMemoryStat{
		Total:       12400,
		Available:   7600,
		Used:        5000,
		UsedPercent: 47.1,
		Free:        1235,
		Active:      8134,
		Inactive:    1124,
		Buffers:     771,
		Cached:      4312,
		Wired:       134,
		Shared:      2142,
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

	ds := &DockerContainerStat{
		Name: "blah",
		CPU: &cpu.CPUTimesStat{
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
		},
		Mem: &docker.CgroupMemStat{
			ContainerID:             "blah",
			Cache:                   1,
			RSS:                     2,
			RSSHuge:                 3,
			MappedFile:              4,
			Pgpgin:                  5,
			Pgpgout:                 6,
			Pgfault:                 7,
			Pgmajfault:              8,
			InactiveAnon:            9,
			ActiveAnon:              10,
			InactiveFile:            11,
			ActiveFile:              12,
			Unevictable:             13,
			HierarchicalMemoryLimit: 14,
			TotalCache:              15,
			TotalRSS:                16,
			TotalRSSHuge:            17,
			TotalMappedFile:         18,
			TotalPgpgIn:             19,
			TotalPgpgOut:            20,
			TotalPgFault:            21,
			TotalPgMajFault:         22,
			TotalInactiveAnon:       23,
			TotalActiveAnon:         24,
			TotalInactiveFile:       25,
			TotalActiveFile:         26,
			TotalUnevictable:        27,
		},
	}

	mps.On("DockerStat").Return([]*DockerContainerStat{ds}, nil)

	err := ss.Gather(&acc)
	require.NoError(t, err)

	assert.True(t, acc.CheckValue("system.load1", 0.3))
	assert.True(t, acc.CheckValue("system.load5", 1.5))
	assert.True(t, acc.CheckValue("system.load15", 0.8))

	cputags := map[string]string{
		"cpu": "cpu0",
	}

	assert.True(t, acc.CheckTaggedValue("cpu.user", 3.1, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu.system", 8.2, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu.idle", 80.1, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu.nice", 1.3, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu.iowait", 0.2, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu.irq", 0.1, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu.softirq", 0.11, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu.steal", 0.0001, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu.guest", 8.1, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu.guestNice", 0.324, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu.stolen", 0.051, cputags))

	tags := map[string]string{
		"path": "/",
	}

	assert.True(t, acc.CheckTaggedValue("disk.total", uint64(128), tags))
	assert.True(t, acc.CheckTaggedValue("disk.used", uint64(105), tags))
	assert.True(t, acc.CheckTaggedValue("disk.free", uint64(23), tags))
	assert.True(t, acc.CheckTaggedValue("disk.inodes_total", uint64(1234), tags))
	assert.True(t, acc.CheckTaggedValue("disk.inodes_free", uint64(234), tags))
	assert.True(t, acc.CheckTaggedValue("disk.inodes_used", uint64(1000), tags))

	ntags := map[string]string{
		"interface": "eth0",
	}

	assert.True(t, acc.CheckTaggedValue("net.bytes_sent", uint64(1123), ntags))
	assert.True(t, acc.CheckTaggedValue("net.bytes_recv", uint64(8734422), ntags))
	assert.True(t, acc.CheckTaggedValue("net.packets_sent", uint64(781), ntags))
	assert.True(t, acc.CheckTaggedValue("net.packets_recv", uint64(23456), ntags))
	assert.True(t, acc.CheckTaggedValue("net.err_in", uint64(832), ntags))
	assert.True(t, acc.CheckTaggedValue("net.err_out", uint64(8), ntags))
	assert.True(t, acc.CheckTaggedValue("net.drop_in", uint64(7), ntags))
	assert.True(t, acc.CheckTaggedValue("net.drop_out", uint64(1), ntags))

	dtags := map[string]string{
		"name":   "sda1",
		"serial": "ab-123-ad",
	}

	assert.True(t, acc.CheckTaggedValue("io.reads", uint64(888), dtags))
	assert.True(t, acc.CheckTaggedValue("io.writes", uint64(5341), dtags))
	assert.True(t, acc.CheckTaggedValue("io.read_bytes", uint64(100000), dtags))
	assert.True(t, acc.CheckTaggedValue("io.write_bytes", uint64(200000), dtags))
	assert.True(t, acc.CheckTaggedValue("io.read_time", uint64(7123), dtags))
	assert.True(t, acc.CheckTaggedValue("io.write_time", uint64(9087), dtags))
	assert.True(t, acc.CheckTaggedValue("io.io_time", uint64(123552), dtags))

	vmtags := map[string]string(nil)

	assert.True(t, acc.CheckTaggedValue("mem.total", uint64(12400), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem.available", uint64(7600), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem.used", uint64(5000), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem.used_prec", float64(47.1), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem.free", uint64(1235), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem.active", uint64(8134), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem.inactive", uint64(1124), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem.buffers", uint64(771), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem.cached", uint64(4312), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem.wired", uint64(134), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem.shared", uint64(2142), vmtags))

	swaptags := map[string]string(nil)

	assert.True(t, acc.CheckTaggedValue("swap.total", uint64(8123), swaptags))
	assert.True(t, acc.CheckTaggedValue("swap.used", uint64(1232), swaptags))
	assert.True(t, acc.CheckTaggedValue("swap.used_perc", float64(12.2), swaptags))
	assert.True(t, acc.CheckTaggedValue("swap.free", uint64(6412), swaptags))
	assert.True(t, acc.CheckTaggedValue("swap.swap_in", uint64(7), swaptags))
	assert.True(t, acc.CheckTaggedValue("swap.swap_out", uint64(830), swaptags))

	dockertags := map[string]string{
		"id": "blah",
	}

	assert.True(t, acc.CheckTaggedValue("docker.user", 3.1, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.system", 8.2, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.idle", 80.1, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.nice", 1.3, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.iowait", 0.2, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.irq", 0.1, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.softirq", 0.11, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.steal", 0.0001, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.guest", 8.1, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.guestNice", 0.324, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.stolen", 0.051, dockertags))

	assert.True(t, acc.CheckTaggedValue("docker.cache", uint64(1), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.rss", uint64(2), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.rss_huge", uint64(3), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.mapped_file", uint64(4), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.swap_in", uint64(5), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.swap_out", uint64(6), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.page_fault", uint64(7), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.page_major_fault", uint64(8), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.inactive_anon", uint64(9), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.active_anon", uint64(10), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.inactive_file", uint64(11), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.active_file", uint64(12), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.unevictable", uint64(13), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.memory_limit", uint64(14), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.total_cache", uint64(15), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.total_rss", uint64(16), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.total_rss_huge", uint64(17), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.total_mapped_file", uint64(18), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.total_swap_in", uint64(19), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.total_swap_out", uint64(20), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.total_page_fault", uint64(21), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.total_page_major_fault", uint64(22), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.total_inactive_anon", uint64(23), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.total_active_anon", uint64(24), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.total_inactive_file", uint64(25), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.total_active_file", uint64(26), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker.total_unevictable", uint64(27), dockertags))
}
