package system

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"

	"github.com/fsouza/go-dockerclient"
)

func TestDockerGatherContainerStats(t *testing.T) {
	var acc testutil.Accumulator
	stats := testStats()

	tags := map[string]string{
		"cont_id":    "foobarbaz",
		"cont_name":  "redis",
		"cont_image": "redis/image",
	}
	gatherContainerStats(stats, &acc, tags)

	// test docker_net measurement
	netfields := map[string]interface{}{
		"rx_dropped": uint64(1),
		"rx_bytes":   uint64(2),
		"rx_errors":  uint64(3),
		"tx_packets": uint64(4),
		"tx_dropped": uint64(1),
		"rx_packets": uint64(2),
		"tx_errors":  uint64(3),
		"tx_bytes":   uint64(4),
	}
	nettags := copyTags(tags)
	nettags["network"] = "eth0"
	acc.AssertContainsTaggedFields(t, "docker_net", netfields, nettags)

	// test docker_blkio measurement
	blkiotags := copyTags(tags)
	blkiotags["device"] = "6:0"
	blkiofields := map[string]interface{}{
		"io_service_bytes_recursive_read": uint64(100),
		"io_serviced_recursive_write":     uint64(101),
	}
	acc.AssertContainsTaggedFields(t, "docker_blkio", blkiofields, blkiotags)

	// test docker_mem measurement
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
	}

	acc.AssertContainsTaggedFields(t, "docker_mem", memfields, tags)

	// test docker_cpu measurement
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
	}
	acc.AssertContainsTaggedFields(t, "docker_cpu", cpufields, cputags)

	cputags["cpu"] = "cpu0"
	cpu0fields := map[string]interface{}{
		"usage_total": uint64(1),
	}
	acc.AssertContainsTaggedFields(t, "docker_cpu", cpu0fields, cputags)

	cputags["cpu"] = "cpu1"
	cpu1fields := map[string]interface{}{
		"usage_total": uint64(1002),
	}
	acc.AssertContainsTaggedFields(t, "docker_cpu", cpu1fields, cputags)
}

func testStats() *docker.Stats {
	stats := &docker.Stats{
		Read:     time.Now(),
		Networks: make(map[string]docker.NetworkStats),
	}

	stats.CPUStats.CPUUsage.PercpuUsage = []uint64{1, 1002}
	stats.CPUStats.CPUUsage.UsageInUsermode = 100
	stats.CPUStats.CPUUsage.TotalUsage = 500
	stats.CPUStats.CPUUsage.UsageInKernelmode = 200
	stats.CPUStats.SystemCPUUsage = 100
	stats.CPUStats.ThrottlingData.Periods = 1

	stats.PreCPUStats.CPUUsage.TotalUsage = 400
	stats.PreCPUStats.SystemCPUUsage = 50

	stats.MemoryStats.Stats.TotalPgmafault = 0
	stats.MemoryStats.Stats.Cache = 0
	stats.MemoryStats.Stats.MappedFile = 0
	stats.MemoryStats.Stats.TotalInactiveFile = 0
	stats.MemoryStats.Stats.Pgpgout = 0
	stats.MemoryStats.Stats.Rss = 0
	stats.MemoryStats.Stats.TotalMappedFile = 0
	stats.MemoryStats.Stats.Writeback = 0
	stats.MemoryStats.Stats.Unevictable = 0
	stats.MemoryStats.Stats.Pgpgin = 0
	stats.MemoryStats.Stats.TotalUnevictable = 0
	stats.MemoryStats.Stats.Pgmajfault = 0
	stats.MemoryStats.Stats.TotalRss = 44
	stats.MemoryStats.Stats.TotalRssHuge = 444
	stats.MemoryStats.Stats.TotalWriteback = 55
	stats.MemoryStats.Stats.TotalInactiveAnon = 0
	stats.MemoryStats.Stats.RssHuge = 0
	stats.MemoryStats.Stats.HierarchicalMemoryLimit = 0
	stats.MemoryStats.Stats.TotalPgfault = 0
	stats.MemoryStats.Stats.TotalActiveFile = 0
	stats.MemoryStats.Stats.ActiveAnon = 0
	stats.MemoryStats.Stats.TotalActiveAnon = 0
	stats.MemoryStats.Stats.TotalPgpgout = 0
	stats.MemoryStats.Stats.TotalCache = 0
	stats.MemoryStats.Stats.InactiveAnon = 0
	stats.MemoryStats.Stats.ActiveFile = 1
	stats.MemoryStats.Stats.Pgfault = 2
	stats.MemoryStats.Stats.InactiveFile = 3
	stats.MemoryStats.Stats.TotalPgpgin = 4

	stats.MemoryStats.MaxUsage = 1001
	stats.MemoryStats.Usage = 1111
	stats.MemoryStats.Failcnt = 1
	stats.MemoryStats.Limit = 2000

	stats.Networks["eth0"] = docker.NetworkStats{
		RxDropped: 1,
		RxBytes:   2,
		RxErrors:  3,
		TxPackets: 4,
		TxDropped: 1,
		RxPackets: 2,
		TxErrors:  3,
		TxBytes:   4,
	}

	sbr := docker.BlkioStatsEntry{
		Major: 6,
		Minor: 0,
		Op:    "read",
		Value: 100,
	}
	sr := docker.BlkioStatsEntry{
		Major: 6,
		Minor: 0,
		Op:    "write",
		Value: 101,
	}

	stats.BlkioStats.IOServiceBytesRecursive = append(
		stats.BlkioStats.IOServiceBytesRecursive, sbr)
	stats.BlkioStats.IOServicedRecursive = append(
		stats.BlkioStats.IOServicedRecursive, sr)

	return stats
}
