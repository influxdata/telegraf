// Package docker contains few helper functions copied from
// https://github.com/docker/cli/blob/master/cli/command/container/stats_helpers.go
package docker

import "github.com/docker/docker/api/types"

func CalculateCPUPercentUnix(previousCPU, previousSystem uint64, v *types.StatsJSON) float64 {
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(v.CPUStats.CPUUsage.TotalUsage) - float64(previousCPU)
		// calculate the change for the entire system between readings
		systemDelta = float64(v.CPUStats.SystemUsage) - float64(previousSystem)
		onlineCPUs  = float64(v.CPUStats.OnlineCPUs)
	)

	if onlineCPUs == 0.0 {
		onlineCPUs = float64(len(v.CPUStats.CPUUsage.PercpuUsage))
	}
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0
	}
	return cpuPercent
}

func calculateCPUPercentWindows(v *types.StatsJSON) float64 {
	// Max number of 100ns intervals between the previous time read and now
	possIntervals := uint64(v.Read.Sub(v.PreRead).Nanoseconds()) // Start with number of ns intervals
	possIntervals /= 100                                         // Convert to number of 100ns intervals
	possIntervals *= uint64(v.NumProcs)                          // Multiple by the number of processors

	// Intervals used
	intervalsUsed := v.CPUStats.CPUUsage.TotalUsage - v.PreCPUStats.CPUUsage.TotalUsage

	// Percentage avoiding divide-by-zero
	if possIntervals > 0 {
		return float64(intervalsUsed) / float64(possIntervals) * 100.0
	}
	return 0.00
}

// CalculateMemUsageUnixNoCache calculate memory usage of the container.
// Cache is intentionally excluded to avoid misinterpretation of the output.
//
// On Docker 19.03 and older, the result is `mem.Usage - mem.Stats["cache"]`.
// On new docker with cgroup v1 host, the result is `mem.Usage - mem.Stats["total_inactive_file"]`.
// On new docker with cgroup v2 host, the result is `mem.Usage - mem.Stats["inactive_file"]`.
//
// This definition is designed to be consistent with past values and the latest docker CLI
// * https://github.com/docker/cli/blob/6e2838e18645e06f3e4b6c5143898ccc44063e3b/cli/command/container/stats_helpers.go#L239
func CalculateMemUsageUnixNoCache(mem types.MemoryStats) float64 {
	// Docker 19.03 and older
	if v, isOldDocker := mem.Stats["cache"]; isOldDocker && v < mem.Usage {
		return float64(mem.Usage - v)
	}
	// cgroup v1
	if v, isCgroup1 := mem.Stats["total_inactive_file"]; isCgroup1 && v < mem.Usage {
		return float64(mem.Usage - v)
	}
	// cgroup v2
	if v := mem.Stats["inactive_file"]; v < mem.Usage {
		return float64(mem.Usage - v)
	}
	return float64(mem.Usage)
}

func CalculateMemPercentUnixNoCache(limit float64, usedNoCache float64) float64 {
	// MemoryStats.Limit will never be 0 unless the container is not running and we haven't
	// got any data from cgroup
	if limit != 0 {
		return usedNoCache / limit * 100.0
	}
	return 0
}
