package runc

import (
	"context"
	"fmt"
	"github.com/containerd/go-runc"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs"
	"strings"

	"time"
)

// Collector returns containers and stats from runc
type Collector interface {
	List(context.Context) ([]*runc.Container, error)
	Stats(context.Context, string) (*runc.Stats, error)
}

type Runc struct {
	collector Collector
	// Runc cgroup root
	Root string `toml:"root"`
	// Path to runc
	Command string `toml:"command"`
	// Report metrics for only these containers
	ContainerInclude []string `toml:"container_name_include"`
	// Do not report metrics for these containers
	ContainerExclude []string `toml:"container_name_exclude"`
	// filter.Filter for including containers by ID
	includeFilter filter.Filter
	// filter.Filter for excluding containers by ID
	excludeFilter filter.Filter
}

func (r *Runc) Description() string {
	return "Read metrics from runc containers"
}

func (r *Runc) SampleConfig() string {
	return `
 ## Path to the runc binary
 command = "/usr/bin/runc"
 ## Runc container root path
 root = "/run/runc"
 ## List of containers to include
 container_name_include = []
 ## List of containers ids to exclude
 container_name_exclude = []
`
}

func (r *Runc) Gather(acc telegraf.Accumulator) error {
	// Instantiate the collector if it has not
	// already been created yet.
	if r.collector == nil {
		// Add a new runc.Runc with user specified
		// configuration.
		r.collector = &runc.Runc{
			Root:    r.Root,
			Command: r.Command,
		}
	}
	// Call runc list
	containers, err := r.collector.List(context.Background())
	if err != nil {
		return err
	}
	// Filter the containers by ID
	containers, err = r.filter(containers)
	if err != nil {
		return err
	}
	// Range each container reported running by runc.
loop:
	for _, container := range containers {
		// Ignore any container that is not running.
		if running(container) {
			// Call runc events --stats
			stats, err := r.collector.Stats(context.Background(), container.ID)
			if err != nil {
				// Report the error and give up
				// for this collection period.
				acc.AddError(err)
				break loop
			}
			// Timestamp set right after the runc command returns.
			ts := time.Now()
			tags := map[string]string{
				"id": container.ID,
			}
			// Iterate each flattened field adding them to
			// the accumulator.
			for section, fields := range flatten(stats) {
				// Add the field to the accumulator.
				acc.AddFields(section, fields, tags, ts)
			}
		}
	}
	return nil
}

// Filtered returns all containers that match
// the user configured filters. If the filters
// do not yet exist they are created first.
func (r *Runc) filter(containers []*runc.Container) ([]*runc.Container, error) {
	// Check if any filters were added
	if len(r.ContainerInclude) == 0 && len(r.ContainerExclude) == 0 {
		// Nothing to filter
		return containers, nil
	}
	// Create the includeFilter if it hasn't already been created.
	if len(r.ContainerInclude) > 0 && r.includeFilter == nil {
		includeFilter, err := filter.Compile(r.ContainerInclude)
		if err != nil {
			return nil, err
		}
		r.includeFilter = includeFilter
	}
	// Create the excludeFilter if it hasn't already been created.
	if len(r.ContainerExclude) > 0 && r.excludeFilter == nil {
		excludeFilter, err := filter.Compile(r.ContainerExclude)
		if err != nil {
			return nil, err
		}
		r.excludeFilter = excludeFilter
	}
	// Filter each container based on include/exclude
	// match criteria.
	filtered := []*runc.Container{}
	for _, container := range containers {
		// If containers are explicitly included
		// we assume all non-matched containers
		// should be excluded.
		if len(r.ContainerInclude) > 0 {
			if r.includeFilter.Match(container.ID) {
				filtered = append(filtered, container)
			}
		}
		// If containers are explicitly excluded
		// we assume any container which doesn't
		// match should be included unless specific
		// containers to include were specified.
		if len(r.ContainerExclude) > 0 {
			if !r.excludeFilter.Match(container.ID) && len(r.ContainerInclude) == 0 {
				filtered = append(filtered, container)
			}
		}
	}
	return filtered, nil
}

// Check if a container is running.
func running(cont *runc.Container) bool { return cont.Status == "running" }

// flatten all runc metrics into a telegraf compatible format.
func flatten(stats *runc.Stats) map[string]map[string]interface{} {
	statsMap := map[string]map[string]interface{}{}
	// CPU Metrics
	statsMap["cpu"] = map[string]interface{}{}
	statsMap["cpu"]["usage_user"] = stats.Cpu.Usage.User
	statsMap["cpu"]["usage_total"] = stats.Cpu.Usage.Total
	statsMap["cpu"]["usage_kernel"] = stats.Cpu.Usage.Kernel
	for i, core := range stats.Cpu.Usage.Percpu {
		statsMap["cpu"][fmt.Sprintf("usage_core_%d", i)] = core
	}
	statsMap["cpu"]["throttling_periods"] = stats.Cpu.Throttling.Periods
	statsMap["cpu"]["throttling_throttled_time"] = stats.Cpu.Throttling.ThrottledTime
	statsMap["cpu"]["throttling_throttled_periods"] = stats.Cpu.Throttling.ThrottledPeriods
	// Pids Metrics
	statsMap["pids"] = map[string]interface{}{}
	statsMap["pids"]["limit"] = stats.Pids.Limit
	statsMap["pids"]["current"] = stats.Pids.Current
	// Memory Metrics
	statsMap["memory"] = map[string]interface{}{}
	statsMap["memory"]["cache"] = stats.Memory.Cache
	statsMap["memory"]["swap_max"] = stats.Memory.Swap.Max
	statsMap["memory"]["swap_limit"] = stats.Memory.Swap.Limit
	statsMap["memory"]["swap_usage"] = stats.Memory.Swap.Usage
	statsMap["memory"]["swap_failcnt"] = stats.Memory.Swap.Failcnt
	statsMap["memory"]["usage_max"] = stats.Memory.Usage.Max
	statsMap["memory"]["usage_limit"] = stats.Memory.Usage.Limit
	statsMap["memory"]["usage_usage"] = stats.Memory.Usage.Usage
	statsMap["memory"]["usage_failcnt"] = stats.Memory.Usage.Failcnt
	statsMap["memory"]["kernel_max"] = stats.Memory.Kernel.Max
	statsMap["memory"]["kernel_limit"] = stats.Memory.Kernel.Limit
	statsMap["memory"]["kernel_usage"] = stats.Memory.Kernel.Usage
	statsMap["memory"]["kernel_failcnt"] = stats.Memory.Kernel.Failcnt
	statsMap["memory"]["kernel_tcp_max"] = stats.Memory.KernelTCP.Max
	statsMap["memory"]["kernel_tcp_limit"] = stats.Memory.KernelTCP.Limit
	statsMap["memory"]["kernel_tcp_usage"] = stats.Memory.KernelTCP.Usage
	statsMap["memory"]["kernel_tcp_failcnt"] = stats.Memory.KernelTCP.Failcnt
	for key, value := range stats.Memory.Raw {
		statsMap["memory"][fmt.Sprintf("raw_%s", key)] = value
	}
	// Hugetlb Metrics
	statsMap["hugetlb"] = map[string]interface{}{}
	for key, value := range stats.Hugetlb {
		statsMap["hugetlb"][fmt.Sprintf("%s_max", key)] = value.Max
		statsMap["hugetlb"][fmt.Sprintf("%s_usage", key)] = value.Usage
		statsMap["hugetlb"][fmt.Sprintf("%s_failcnt", key)] = value.Failcnt
	}
	// Blkio Metrics
	statsMap["blkio"] = map[string]interface{}{}
	AddBlkioEntries(stats.Blkio.IoTimeRecursive, "time_recursive", statsMap["blkio"])
	AddBlkioEntries(stats.Blkio.SectorsRecursive, "sectors_recursive", statsMap["blkio"])
	AddBlkioEntries(stats.Blkio.IoMergedRecursive, "merged_recursive", statsMap["blkio"])
	AddBlkioEntries(stats.Blkio.IoQueuedRecursive, "queued_recursive", statsMap["blkio"])
	AddBlkioEntries(stats.Blkio.IoServicedRecursive, "serviced_recursive", statsMap["blkio"])
	AddBlkioEntries(stats.Blkio.IoWaitTimeRecursive, "wait_time_recursive", statsMap["blkio"])
	AddBlkioEntries(stats.Blkio.IoServiceBytesRecursive, "service_bytes_recursive", statsMap["blkio"])
	return statsMap
}

// AddBlkioEntries flattens a BlkioEntry into a map of key-values and adds them to statsMap
func AddBlkioEntries(entries []runc.BlkioEntry, prefix string, statsMap map[string]interface{}) {
	for _, entry := range entries {
		statsMap[fmt.Sprintf("%s_%s_%d_%d", prefix, strings.ToLower(entry.Op), entry.Major, entry.Minor)] = entry.Value
	}
}

func init() {
	inputs.Add("runc", func() telegraf.Input {
		return &Runc{
			Command: "/usr/bin/runc",
			Root:    "/run/runc",
		}
	})
}
