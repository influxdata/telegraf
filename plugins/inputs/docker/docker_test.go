package docker

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/require"
)

func TestDockerGatherContainerStats(t *testing.T) {
	var acc testutil.Accumulator
	stats := testStats()

	tags := map[string]string{
		"container_name":  "redis",
		"container_image": "redis/image",
	}
	gatherContainerStats(stats, &acc, tags, "123456789", true, true)

	// test docker_container_net measurement
	netfields := map[string]interface{}{
		"rx_dropped":   uint64(1),
		"rx_bytes":     uint64(2),
		"rx_errors":    uint64(3),
		"tx_packets":   uint64(4),
		"tx_dropped":   uint64(1),
		"rx_packets":   uint64(2),
		"tx_errors":    uint64(3),
		"tx_bytes":     uint64(4),
		"container_id": "123456789",
	}
	nettags := copyTags(tags)
	nettags["network"] = "eth0"
	acc.AssertContainsTaggedFields(t, "docker_container_net", netfields, nettags)

	netfields = map[string]interface{}{
		"rx_dropped":   uint64(6),
		"rx_bytes":     uint64(8),
		"rx_errors":    uint64(10),
		"tx_packets":   uint64(12),
		"tx_dropped":   uint64(6),
		"rx_packets":   uint64(8),
		"tx_errors":    uint64(10),
		"tx_bytes":     uint64(12),
		"container_id": "123456789",
	}
	nettags = copyTags(tags)
	nettags["network"] = "total"
	acc.AssertContainsTaggedFields(t, "docker_container_net", netfields, nettags)

	// test docker_blkio measurement
	blkiotags := copyTags(tags)
	blkiotags["device"] = "6:0"
	blkiofields := map[string]interface{}{
		"io_service_bytes_recursive_read": uint64(100),
		"io_serviced_recursive_write":     uint64(101),
		"container_id":                    "123456789",
	}
	acc.AssertContainsTaggedFields(t, "docker_container_blkio", blkiofields, blkiotags)

	blkiotags = copyTags(tags)
	blkiotags["device"] = "total"
	blkiofields = map[string]interface{}{
		"io_service_bytes_recursive_read": uint64(100),
		"io_serviced_recursive_write":     uint64(302),
		"container_id":                    "123456789",
	}
	acc.AssertContainsTaggedFields(t, "docker_container_blkio", blkiofields, blkiotags)

	// test docker_container_mem measurement
	memfields := map[string]interface{}{
		"max_usage":                 uint64(1001),
		"usage":                     uint64(1111),
		"fail_count":                uint64(1),
		"limit":                     uint64(2000),
		"total_pgmafault":           uint64(0),
		"cache":                     uint64(0),
		"mapped_file":               uint64(0),
		"total_inactive_file":       uint64(0),
		"pgpgout":                   uint64(0),
		"rss":                       uint64(0),
		"total_mapped_file":         uint64(0),
		"writeback":                 uint64(0),
		"unevictable":               uint64(0),
		"pgpgin":                    uint64(0),
		"total_unevictable":         uint64(0),
		"pgmajfault":                uint64(0),
		"total_rss":                 uint64(44),
		"total_rss_huge":            uint64(444),
		"total_writeback":           uint64(55),
		"total_inactive_anon":       uint64(0),
		"rss_huge":                  uint64(0),
		"hierarchical_memory_limit": uint64(0),
		"total_pgfault":             uint64(0),
		"total_active_file":         uint64(0),
		"active_anon":               uint64(0),
		"total_active_anon":         uint64(0),
		"total_pgpgout":             uint64(0),
		"total_cache":               uint64(0),
		"inactive_anon":             uint64(0),
		"active_file":               uint64(1),
		"pgfault":                   uint64(2),
		"inactive_file":             uint64(3),
		"total_pgpgin":              uint64(4),
		"usage_percent":             float64(55.55),
		"container_id":              "123456789",
	}

	acc.AssertContainsTaggedFields(t, "docker_container_mem", memfields, tags)

	// test docker_container_cpu measurement
	cputags := copyTags(tags)
	cputags["cpu"] = "cpu-total"
	cpufields := map[string]interface{}{
		"usage_total":                  uint64(500),
		"usage_in_usermode":            uint64(100),
		"usage_in_kernelmode":          uint64(200),
		"usage_system":                 uint64(100),
		"throttling_periods":           uint64(1),
		"throttling_throttled_periods": uint64(0),
		"throttling_throttled_time":    uint64(0),
		"usage_percent":                float64(400.0),
		"container_id":                 "123456789",
	}
	acc.AssertContainsTaggedFields(t, "docker_container_cpu", cpufields, cputags)

	cputags["cpu"] = "cpu0"
	cpu0fields := map[string]interface{}{
		"usage_total":  uint64(1),
		"container_id": "123456789",
	}
	acc.AssertContainsTaggedFields(t, "docker_container_cpu", cpu0fields, cputags)

	cputags["cpu"] = "cpu1"
	cpu1fields := map[string]interface{}{
		"usage_total":  uint64(1002),
		"container_id": "123456789",
	}
	acc.AssertContainsTaggedFields(t, "docker_container_cpu", cpu1fields, cputags)
}

func testStats() *types.StatsJSON {
	stats := &types.StatsJSON{}
	stats.Read = time.Now()
	stats.Networks = make(map[string]types.NetworkStats)

	stats.CPUStats.CPUUsage.PercpuUsage = []uint64{1, 1002}
	stats.CPUStats.CPUUsage.UsageInUsermode = 100
	stats.CPUStats.CPUUsage.TotalUsage = 500
	stats.CPUStats.CPUUsage.UsageInKernelmode = 200
	stats.CPUStats.SystemUsage = 100
	stats.CPUStats.ThrottlingData.Periods = 1

	stats.PreCPUStats.CPUUsage.TotalUsage = 400
	stats.PreCPUStats.SystemUsage = 50

	stats.MemoryStats.Stats = make(map[string]uint64)
	stats.MemoryStats.Stats["total_pgmajfault"] = 0
	stats.MemoryStats.Stats["cache"] = 0
	stats.MemoryStats.Stats["mapped_file"] = 0
	stats.MemoryStats.Stats["total_inactive_file"] = 0
	stats.MemoryStats.Stats["pagpgout"] = 0
	stats.MemoryStats.Stats["rss"] = 0
	stats.MemoryStats.Stats["total_mapped_file"] = 0
	stats.MemoryStats.Stats["writeback"] = 0
	stats.MemoryStats.Stats["unevictable"] = 0
	stats.MemoryStats.Stats["pgpgin"] = 0
	stats.MemoryStats.Stats["total_unevictable"] = 0
	stats.MemoryStats.Stats["pgmajfault"] = 0
	stats.MemoryStats.Stats["total_rss"] = 44
	stats.MemoryStats.Stats["total_rss_huge"] = 444
	stats.MemoryStats.Stats["total_write_back"] = 55
	stats.MemoryStats.Stats["total_inactive_anon"] = 0
	stats.MemoryStats.Stats["rss_huge"] = 0
	stats.MemoryStats.Stats["hierarchical_memory_limit"] = 0
	stats.MemoryStats.Stats["total_pgfault"] = 0
	stats.MemoryStats.Stats["total_active_file"] = 0
	stats.MemoryStats.Stats["active_anon"] = 0
	stats.MemoryStats.Stats["total_active_anon"] = 0
	stats.MemoryStats.Stats["total_pgpgout"] = 0
	stats.MemoryStats.Stats["total_cache"] = 0
	stats.MemoryStats.Stats["inactive_anon"] = 0
	stats.MemoryStats.Stats["active_file"] = 1
	stats.MemoryStats.Stats["pgfault"] = 2
	stats.MemoryStats.Stats["inactive_file"] = 3
	stats.MemoryStats.Stats["total_pgpgin"] = 4

	stats.MemoryStats.MaxUsage = 1001
	stats.MemoryStats.Usage = 1111
	stats.MemoryStats.Failcnt = 1
	stats.MemoryStats.Limit = 2000

	stats.Networks["eth0"] = types.NetworkStats{
		RxDropped: 1,
		RxBytes:   2,
		RxErrors:  3,
		TxPackets: 4,
		TxDropped: 1,
		RxPackets: 2,
		TxErrors:  3,
		TxBytes:   4,
	}

	stats.Networks["eth1"] = types.NetworkStats{
		RxDropped: 5,
		RxBytes:   6,
		RxErrors:  7,
		TxPackets: 8,
		TxDropped: 5,
		RxPackets: 6,
		TxErrors:  7,
		TxBytes:   8,
	}

	sbr := types.BlkioStatEntry{
		Major: 6,
		Minor: 0,
		Op:    "read",
		Value: 100,
	}
	sr := types.BlkioStatEntry{
		Major: 6,
		Minor: 0,
		Op:    "write",
		Value: 101,
	}
	sr2 := types.BlkioStatEntry{
		Major: 6,
		Minor: 1,
		Op:    "write",
		Value: 201,
	}

	stats.BlkioStats.IoServiceBytesRecursive = append(
		stats.BlkioStats.IoServiceBytesRecursive, sbr)
	stats.BlkioStats.IoServicedRecursive = append(
		stats.BlkioStats.IoServicedRecursive, sr)
	stats.BlkioStats.IoServicedRecursive = append(
		stats.BlkioStats.IoServicedRecursive, sr2)

	return stats
}

var gatherLabelsTests = []struct {
	include     []string
	exclude     []string
	expected    []string
	notexpected []string
}{
	{[]string{}, []string{}, []string{"label1", "label2"}, []string{}},
	{[]string{"*"}, []string{}, []string{"label1", "label2"}, []string{}},
	{[]string{"lab*"}, []string{}, []string{"label1", "label2"}, []string{}},
	{[]string{"label1"}, []string{}, []string{"label1"}, []string{"label2"}},
	{[]string{"label1*"}, []string{}, []string{"label1"}, []string{"label2"}},
	{[]string{}, []string{"*"}, []string{}, []string{"label1", "label2"}},
	{[]string{}, []string{"lab*"}, []string{}, []string{"label1", "label2"}},
	{[]string{}, []string{"label1"}, []string{"label2"}, []string{"label1"}},
	{[]string{"*"}, []string{"*"}, []string{}, []string{"label1", "label2"}},
}

func TestDockerGatherLabels(t *testing.T) {
	for _, tt := range gatherLabelsTests {
		var acc testutil.Accumulator
		d := Docker{
			client:  nil,
			testing: true,
		}

		for _, label := range tt.include {
			d.LabelInclude = append(d.LabelInclude, label)
		}
		for _, label := range tt.exclude {
			d.LabelExclude = append(d.LabelExclude, label)
		}

		err := d.Gather(&acc)
		require.NoError(t, err)

		for _, label := range tt.expected {
			if !acc.HasTag("docker_container_cpu", label) {
				t.Errorf("Didn't get expected label of %s.  Test was:  Include: %s  Exclude %s",
					label, tt.include, tt.exclude)
			}
		}

		for _, label := range tt.notexpected {
			if acc.HasTag("docker_container_cpu", label) {
				t.Errorf("Got unexpected label of %s.  Test was:  Include: %s  Exclude %s",
					label, tt.include, tt.exclude)
			}
		}
	}
}

func TestDockerGatherInfo(t *testing.T) {
	var acc testutil.Accumulator
	d := Docker{
		client:  nil,
		testing: true,
	}

	err := acc.GatherError(d.Gather)
	require.NoError(t, err)

	acc.AssertContainsTaggedFields(t,
		"docker",
		map[string]interface{}{
			"n_listener_events":       int(0),
			"n_cpus":                  int(4),
			"n_used_file_descriptors": int(19),
			"n_containers":            int(108),
			"n_containers_running":    int(98),
			"n_containers_stopped":    int(6),
			"n_containers_paused":     int(3),
			"n_images":                int(199),
			"n_goroutines":            int(39),
		},
		map[string]string{"engine_host": "absol"},
	)

	acc.AssertContainsTaggedFields(t,
		"docker_data",
		map[string]interface{}{
			"used":      int64(17300000000),
			"total":     int64(107400000000),
			"available": int64(36530000000),
		},
		map[string]string{
			"unit":        "bytes",
			"engine_host": "absol",
		},
	)
	acc.AssertContainsTaggedFields(t,
		"docker_container_cpu",
		map[string]interface{}{
			"usage_total":  uint64(1231652),
			"container_id": "b7dfbb9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296e2173",
		},
		map[string]string{
			"container_name":    "etcd2",
			"container_image":   "quay.io:4443/coreos/etcd",
			"cpu":               "cpu3",
			"container_version": "v2.2.2",
			"engine_host":       "absol",
			"label1":            "test_value_1",
			"label2":            "test_value_2",
		},
	)
	acc.AssertContainsTaggedFields(t,
		"docker_container_mem",
		map[string]interface{}{
			"total_pgpgout":             uint64(0),
			"usage_percent":             float64(0),
			"rss":                       uint64(0),
			"total_writeback":           uint64(0),
			"active_anon":               uint64(0),
			"total_pgmafault":           uint64(0),
			"total_rss":                 uint64(0),
			"total_unevictable":         uint64(0),
			"active_file":               uint64(0),
			"total_mapped_file":         uint64(0),
			"pgpgin":                    uint64(0),
			"total_active_file":         uint64(0),
			"total_active_anon":         uint64(0),
			"total_cache":               uint64(0),
			"inactive_anon":             uint64(0),
			"pgmajfault":                uint64(0),
			"total_inactive_anon":       uint64(0),
			"total_rss_huge":            uint64(0),
			"rss_huge":                  uint64(0),
			"hierarchical_memory_limit": uint64(0),
			"pgpgout":                   uint64(0),
			"unevictable":               uint64(0),
			"total_inactive_file":       uint64(0),
			"writeback":                 uint64(0),
			"total_pgfault":             uint64(0),
			"total_pgpgin":              uint64(0),
			"cache":                     uint64(0),
			"mapped_file":               uint64(0),
			"inactive_file":             uint64(0),
			"max_usage":                 uint64(0),
			"fail_count":                uint64(0),
			"pgfault":                   uint64(0),
			"usage":                     uint64(0),
			"limit":                     uint64(18935443456),
			"container_id":              "b7dfbb9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296e2173",
		},
		map[string]string{
			"engine_host":       "absol",
			"container_name":    "etcd2",
			"container_image":   "quay.io:4443/coreos/etcd",
			"container_version": "v2.2.2",
			"label1":            "test_value_1",
			"label2":            "test_value_2",
		},
	)

	//fmt.Print(info)
}
