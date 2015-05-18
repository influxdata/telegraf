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

	ss := &SystemStats{ps: &mps}

	err := ss.Gather(&acc)
	require.NoError(t, err)

	assert.True(t, acc.CheckValue("system_load1", 0.3))
	assert.True(t, acc.CheckValue("system_load5", 1.5))
	assert.True(t, acc.CheckValue("system_load15", 0.8))

	cs := &CPUStats{ps: &mps}

	cputags := map[string]string{
		"cpu": "cpu0",
	}

	err = cs.Gather(&acc)
	require.NoError(t, err)

	assert.True(t, acc.CheckTaggedValue("cpu_user", 3.1, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu_system", 8.2, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu_idle", 80.1, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu_nice", 1.3, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu_iowait", 0.2, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu_irq", 0.1, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu_softirq", 0.11, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu_steal", 0.0001, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu_guest", 8.1, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu_guestNice", 0.324, cputags))
	assert.True(t, acc.CheckTaggedValue("cpu_stolen", 0.051, cputags))

	err = (&DiskStats{&mps}).Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"path": "/",
	}

	assert.True(t, acc.CheckTaggedValue("disk_total", uint64(128), tags))
	assert.True(t, acc.CheckTaggedValue("disk_used", uint64(105), tags))
	assert.True(t, acc.CheckTaggedValue("disk_free", uint64(23), tags))
	assert.True(t, acc.CheckTaggedValue("disk_inodes_total", uint64(1234), tags))
	assert.True(t, acc.CheckTaggedValue("disk_inodes_free", uint64(234), tags))
	assert.True(t, acc.CheckTaggedValue("disk_inodes_used", uint64(1000), tags))

	err = (&NetIOStats{&mps}).Gather(&acc)
	require.NoError(t, err)

	ntags := map[string]string{
		"interface": "eth0",
	}

	assert.True(t, acc.CheckTaggedValue("net_bytes_sent", uint64(1123), ntags))
	assert.True(t, acc.CheckTaggedValue("net_bytes_recv", uint64(8734422), ntags))
	assert.True(t, acc.CheckTaggedValue("net_packets_sent", uint64(781), ntags))
	assert.True(t, acc.CheckTaggedValue("net_packets_recv", uint64(23456), ntags))
	assert.True(t, acc.CheckTaggedValue("net_err_in", uint64(832), ntags))
	assert.True(t, acc.CheckTaggedValue("net_err_out", uint64(8), ntags))
	assert.True(t, acc.CheckTaggedValue("net_drop_in", uint64(7), ntags))
	assert.True(t, acc.CheckTaggedValue("net_drop_out", uint64(1), ntags))

	err = (&DiskIOStats{&mps}).Gather(&acc)
	require.NoError(t, err)

	dtags := map[string]string{
		"name":   "sda1",
		"serial": "ab-123-ad",
	}

	assert.True(t, acc.CheckTaggedValue("io_reads", uint64(888), dtags))
	assert.True(t, acc.CheckTaggedValue("io_writes", uint64(5341), dtags))
	assert.True(t, acc.CheckTaggedValue("io_read_bytes", uint64(100000), dtags))
	assert.True(t, acc.CheckTaggedValue("io_write_bytes", uint64(200000), dtags))
	assert.True(t, acc.CheckTaggedValue("io_read_time", uint64(7123), dtags))
	assert.True(t, acc.CheckTaggedValue("io_write_time", uint64(9087), dtags))
	assert.True(t, acc.CheckTaggedValue("io_io_time", uint64(123552), dtags))

	err = (&MemStats{&mps}).Gather(&acc)
	require.NoError(t, err)

	vmtags := map[string]string(nil)

	assert.True(t, acc.CheckTaggedValue("mem_total", uint64(12400), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem_available", uint64(7600), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem_used", uint64(5000), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem_used_prec", float64(47.1), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem_free", uint64(1235), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem_active", uint64(8134), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem_inactive", uint64(1124), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem_buffers", uint64(771), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem_cached", uint64(4312), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem_wired", uint64(134), vmtags))
	assert.True(t, acc.CheckTaggedValue("mem_shared", uint64(2142), vmtags))

	err = (&SwapStats{&mps}).Gather(&acc)
	require.NoError(t, err)

	swaptags := map[string]string(nil)

	assert.True(t, acc.CheckTaggedValue("swap_total", uint64(8123), swaptags))
	assert.True(t, acc.CheckTaggedValue("swap_used", uint64(1232), swaptags))
	assert.True(t, acc.CheckTaggedValue("swap_used_perc", float64(12.2), swaptags))
	assert.True(t, acc.CheckTaggedValue("swap_free", uint64(6412), swaptags))
	assert.True(t, acc.CheckTaggedValue("swap_in", uint64(7), swaptags))
	assert.True(t, acc.CheckTaggedValue("swap_out", uint64(830), swaptags))

	err = (&DockerStats{&mps}).Gather(&acc)
	require.NoError(t, err)

	dockertags := map[string]string{
		"id": "blah",
	}

	assert.True(t, acc.CheckTaggedValue("docker_user", 3.1, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_system", 8.2, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_idle", 80.1, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_nice", 1.3, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_iowait", 0.2, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_irq", 0.1, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_softirq", 0.11, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_steal", 0.0001, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_guest", 8.1, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_guestNice", 0.324, dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_stolen", 0.051, dockertags))

	assert.True(t, acc.CheckTaggedValue("docker_cache", uint64(1), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_rss", uint64(2), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_rss_huge", uint64(3), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_mapped_file", uint64(4), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_swap_in", uint64(5), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_swap_out", uint64(6), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_page_fault", uint64(7), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_page_major_fault", uint64(8), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_inactive_anon", uint64(9), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_active_anon", uint64(10), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_inactive_file", uint64(11), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_active_file", uint64(12), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_unevictable", uint64(13), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_memory_limit", uint64(14), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_total_cache", uint64(15), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_total_rss", uint64(16), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_total_rss_huge", uint64(17), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_total_mapped_file", uint64(18), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_total_swap_in", uint64(19), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_total_swap_out", uint64(20), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_total_page_fault", uint64(21), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_total_page_major_fault", uint64(22), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_total_inactive_anon", uint64(23), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_total_active_anon", uint64(24), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_total_inactive_file", uint64(25), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_total_active_file", uint64(26), dockertags))
	assert.True(t, acc.CheckTaggedValue("docker_total_unevictable", uint64(27), dockertags))
}
