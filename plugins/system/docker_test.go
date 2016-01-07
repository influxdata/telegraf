// +build linux

package system

import (
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/docker"

	"github.com/stretchr/testify/require"
)

func TestDockerStats_GenerateStats(t *testing.T) {
	var mps MockPS
	var acc testutil.Accumulator

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

	err := (&DockerStats{&mps}).Gather(&acc)
	require.NoError(t, err)

	dockertags := map[string]string{
		"name":    "blah",
		"id":      "",
		"command": "",
	}

	fields := map[string]interface{}{
		"user":       3.1,
		"system":     8.2,
		"idle":       80.1,
		"nice":       1.3,
		"iowait":     0.2,
		"irq":        0.1,
		"softirq":    0.11,
		"steal":      0.0001,
		"guest":      8.1,
		"guest_nice": 0.324,

		"cache":                  uint64(1),
		"rss":                    uint64(2),
		"rss_huge":               uint64(3),
		"mapped_file":            uint64(4),
		"swap_in":                uint64(5),
		"swap_out":               uint64(6),
		"page_fault":             uint64(7),
		"page_major_fault":       uint64(8),
		"inactive_anon":          uint64(9),
		"active_anon":            uint64(10),
		"inactive_file":          uint64(11),
		"active_file":            uint64(12),
		"unevictable":            uint64(13),
		"memory_limit":           uint64(14),
		"total_cache":            uint64(15),
		"total_rss":              uint64(16),
		"total_rss_huge":         uint64(17),
		"total_mapped_file":      uint64(18),
		"total_swap_in":          uint64(19),
		"total_swap_out":         uint64(20),
		"total_page_fault":       uint64(21),
		"total_page_major_fault": uint64(22),
		"total_inactive_anon":    uint64(23),
		"total_active_anon":      uint64(24),
		"total_inactive_file":    uint64(25),
		"total_active_file":      uint64(26),
		"total_unevictable":      uint64(27),
	}

	acc.AssertContainsTaggedFields(t, "docker", fields, dockertags)
}
