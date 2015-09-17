package docker

import "errors"

var ErrDockerNotAvailable = errors.New("docker not available")
var ErrCgroupNotAvailable = errors.New("cgroup not available")

type CgroupMemStat struct {
	ContainerID             string `json:"container_id"`
	Cache                   uint64 `json:"cache"`
	RSS                     uint64 `json:"rss"`
	RSSHuge                 uint64 `json:"rss_huge"`
	MappedFile              uint64 `json:"mapped_file"`
	Pgpgin                  uint64 `json:"pgpgin"`
	Pgpgout                 uint64 `json:"pgpgout"`
	Pgfault                 uint64 `json:"pgfault"`
	Pgmajfault              uint64 `json:"pgmajfault"`
	InactiveAnon            uint64 `json:"inactive_anon"`
	ActiveAnon              uint64 `json:"active_anon"`
	InactiveFile            uint64 `json:"inactive_file"`
	ActiveFile              uint64 `json:"active_file"`
	Unevictable             uint64 `json:"unevictable"`
	HierarchicalMemoryLimit uint64 `json:"hierarchical_memory_limit"`
	TotalCache              uint64 `json:"total_cache"`
	TotalRSS                uint64 `json:"total_rss"`
	TotalRSSHuge            uint64 `json:"total_rss_huge"`
	TotalMappedFile         uint64 `json:"total_mapped_file"`
	TotalPgpgIn             uint64 `json:"total_pgpgin"`
	TotalPgpgOut            uint64 `json:"total_pgpgout"`
	TotalPgFault            uint64 `json:"total_pgfault"`
	TotalPgMajFault         uint64 `json:"total_pgmajfault"`
	TotalInactiveAnon       uint64 `json:"total_inactive_anon"`
	TotalActiveAnon         uint64 `json:"total_active_anon"`
	TotalInactiveFile       uint64 `json:"total_inactive_file"`
	TotalActiveFile         uint64 `json:"total_active_file"`
	TotalUnevictable        uint64 `json:"total_unevictable"`
}
