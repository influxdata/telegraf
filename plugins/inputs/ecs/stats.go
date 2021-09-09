package ecs

import (
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs/docker"
)

func parseContainerStats(c Container, acc telegraf.Accumulator, tags map[string]string) {
	id := c.ID
	stats := c.Stats
	tm := stats.Read

	if tm.Before(time.Unix(0, 0)) {
		tm = time.Now()
	}

	metastats(id, c, acc, tags, tm)
	memstats(id, stats, acc, tags, tm)
	cpustats(id, stats, acc, tags, tm)
	netstats(id, stats, acc, tags, tm)
	blkstats(id, stats, acc, tags, tm)
}

func metastats(id string, c Container, acc telegraf.Accumulator, tags map[string]string, tm time.Time) {
	metafields := map[string]interface{}{
		"container_id":   id,
		"docker_name":    c.DockerName,
		"image":          c.Image,
		"image_id":       c.ImageID,
		"desired_status": c.DesiredStatus,
		"known_status":   c.KnownStatus,
		"limit_cpu":      c.Limits["CPU"],
		"limit_mem":      c.Limits["Memory"],
		"created_at":     c.CreatedAt,
		"started_at":     c.StartedAt,
		"type":           c.Type,
	}

	acc.AddFields("ecs_container_meta", metafields, tags, tm)
}

func memstats(id string, stats types.StatsJSON, acc telegraf.Accumulator, tags map[string]string, tm time.Time) {
	memfields := map[string]interface{}{
		"container_id": id,
	}

	memstats := []string{
		"active_anon",
		"active_file",
		"cache",
		"hierarchical_memory_limit",
		"inactive_anon",
		"inactive_file",
		"mapped_file",
		"pgfault",
		"pgmajfault",
		"pgpgin",
		"pgpgout",
		"rss",
		"rss_huge",
		"total_active_anon",
		"total_active_file",
		"total_cache",
		"total_inactive_anon",
		"total_inactive_file",
		"total_mapped_file",
		"total_pgfault",
		"total_pgmajfault",
		"total_pgpgin",
		"total_pgpgout",
		"total_rss",
		"total_rss_huge",
		"total_unevictable",
		"total_writeback",
		"unevictable",
		"writeback",
	}

	for _, field := range memstats {
		if value, ok := stats.MemoryStats.Stats[field]; ok {
			memfields[field] = value
		}
	}
	if stats.MemoryStats.Failcnt != 0 {
		memfields["fail_count"] = stats.MemoryStats.Failcnt
	}

	memfields["limit"] = stats.MemoryStats.Limit
	memfields["max_usage"] = stats.MemoryStats.MaxUsage

	mem := docker.CalculateMemUsageUnixNoCache(stats.MemoryStats)
	memLimit := float64(stats.MemoryStats.Limit)
	memfields["usage"] = uint64(mem)
	memfields["usage_percent"] = docker.CalculateMemPercentUnixNoCache(memLimit, mem)

	acc.AddFields("ecs_container_mem", memfields, tags, tm)
}

func cpustats(id string, stats types.StatsJSON, acc telegraf.Accumulator, tags map[string]string, tm time.Time) {
	cpufields := map[string]interface{}{
		"usage_total":                  stats.CPUStats.CPUUsage.TotalUsage,
		"usage_in_usermode":            stats.CPUStats.CPUUsage.UsageInUsermode,
		"usage_in_kernelmode":          stats.CPUStats.CPUUsage.UsageInKernelmode,
		"usage_system":                 stats.CPUStats.SystemUsage,
		"throttling_periods":           stats.CPUStats.ThrottlingData.Periods,
		"throttling_throttled_periods": stats.CPUStats.ThrottlingData.ThrottledPeriods,
		"throttling_throttled_time":    stats.CPUStats.ThrottlingData.ThrottledTime,
		"container_id":                 id,
	}

	previousCPU := stats.PreCPUStats.CPUUsage.TotalUsage
	previousSystem := stats.PreCPUStats.SystemUsage
	cpuPercent := docker.CalculateCPUPercentUnix(previousCPU, previousSystem, &stats)
	cpufields["usage_percent"] = cpuPercent

	cputags := copyTags(tags)
	cputags["cpu"] = "cpu-total"
	acc.AddFields("ecs_container_cpu", cpufields, cputags, tm)

	// If we have OnlineCPUs field, then use it to restrict stats gathering to only Online CPUs
	// (https://github.com/moby/moby/commit/115f91d7575d6de6c7781a96a082f144fd17e400)
	var percpuusage []uint64
	if stats.CPUStats.OnlineCPUs > 0 {
		percpuusage = stats.CPUStats.CPUUsage.PercpuUsage[:stats.CPUStats.OnlineCPUs]
	} else {
		percpuusage = stats.CPUStats.CPUUsage.PercpuUsage
	}

	for i, percpu := range percpuusage {
		percputags := copyTags(tags)
		percputags["cpu"] = fmt.Sprintf("cpu%d", i)
		fields := map[string]interface{}{
			"usage_total":  percpu,
			"container_id": id,
		}
		acc.AddFields("ecs_container_cpu", fields, percputags, tm)
	}
}

func netstats(id string, stats types.StatsJSON, acc telegraf.Accumulator, tags map[string]string, tm time.Time) {
	totalNetworkStatMap := make(map[string]interface{})
	for network, netstats := range stats.Networks {
		netfields := map[string]interface{}{
			"rx_dropped":   netstats.RxDropped,
			"rx_bytes":     netstats.RxBytes,
			"rx_errors":    netstats.RxErrors,
			"tx_packets":   netstats.TxPackets,
			"tx_dropped":   netstats.TxDropped,
			"rx_packets":   netstats.RxPackets,
			"tx_errors":    netstats.TxErrors,
			"tx_bytes":     netstats.TxBytes,
			"container_id": id,
		}

		nettags := copyTags(tags)
		nettags["network"] = network
		acc.AddFields("ecs_container_net", netfields, nettags, tm)

		for field, value := range netfields {
			if field == "container_id" {
				continue
			}

			var uintV uint64
			switch v := value.(type) {
			case uint64:
				uintV = v
			case int64:
				uintV = uint64(v)
			default:
				continue
			}

			_, ok := totalNetworkStatMap[field]
			if ok {
				totalNetworkStatMap[field] = totalNetworkStatMap[field].(uint64) + uintV
			} else {
				totalNetworkStatMap[field] = uintV
			}
		}
	}

	// totalNetworkStatMap could be empty if container is running with --net=host.
	if len(totalNetworkStatMap) != 0 {
		nettags := copyTags(tags)
		nettags["network"] = "total"
		totalNetworkStatMap["container_id"] = id
		acc.AddFields("ecs_container_net", totalNetworkStatMap, nettags, tm)
	}
}

func blkstats(id string, stats types.StatsJSON, acc telegraf.Accumulator, tags map[string]string, tm time.Time) {
	blkioStats := stats.BlkioStats
	// Make a map of devices to their block io stats
	deviceStatMap := make(map[string]map[string]interface{})

	for _, metric := range blkioStats.IoServiceBytesRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		_, ok := deviceStatMap[device]
		if !ok {
			deviceStatMap[device] = make(map[string]interface{})
		}

		field := fmt.Sprintf("io_service_bytes_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IoServicedRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		_, ok := deviceStatMap[device]
		if !ok {
			deviceStatMap[device] = make(map[string]interface{})
		}

		field := fmt.Sprintf("io_serviced_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IoQueuedRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := fmt.Sprintf("io_queue_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IoServiceTimeRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := fmt.Sprintf("io_service_time_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IoWaitTimeRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := fmt.Sprintf("io_wait_time_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IoMergedRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		field := fmt.Sprintf("io_merged_recursive_%s", strings.ToLower(metric.Op))
		deviceStatMap[device][field] = metric.Value
	}

	for _, metric := range blkioStats.IoTimeRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		deviceStatMap[device]["io_time_recursive"] = metric.Value
	}

	for _, metric := range blkioStats.SectorsRecursive {
		device := fmt.Sprintf("%d:%d", metric.Major, metric.Minor)
		deviceStatMap[device]["sectors_recursive"] = metric.Value
	}

	totalStatMap := make(map[string]interface{})
	for device, fields := range deviceStatMap {
		fields["container_id"] = id

		iotags := copyTags(tags)
		iotags["device"] = device
		acc.AddFields("ecs_container_blkio", fields, iotags, tm)

		for field, value := range fields {
			if field == "container_id" {
				continue
			}

			var uintV uint64
			switch v := value.(type) {
			case uint64:
				uintV = v
			case int64:
				uintV = uint64(v)
			default:
				continue
			}

			_, ok := totalStatMap[field]
			if ok {
				totalStatMap[field] = totalStatMap[field].(uint64) + uintV
			} else {
				totalStatMap[field] = uintV
			}
		}
	}

	totalStatMap["container_id"] = id
	iotags := copyTags(tags)
	iotags["device"] = "total"
	acc.AddFields("ecs_container_blkio", totalStatMap, iotags, tm)
}
