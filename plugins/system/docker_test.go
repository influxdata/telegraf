// +build linux

package system

import (
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/docker"

	"github.com/stretchr/testify/assert"
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

	assert.True(t, acc.CheckTaggedValue("user", 3.1, dockertags))
	assert.True(t, acc.CheckTaggedValue("system", 8.2, dockertags))
	assert.True(t, acc.CheckTaggedValue("idle", 80.1, dockertags))
	assert.True(t, acc.CheckTaggedValue("nice", 1.3, dockertags))
	assert.True(t, acc.CheckTaggedValue("iowait", 0.2, dockertags))
	assert.True(t, acc.CheckTaggedValue("irq", 0.1, dockertags))
	assert.True(t, acc.CheckTaggedValue("softirq", 0.11, dockertags))
	assert.True(t, acc.CheckTaggedValue("steal", 0.0001, dockertags))
	assert.True(t, acc.CheckTaggedValue("guest", 8.1, dockertags))
	assert.True(t, acc.CheckTaggedValue("guest_nice", 0.324, dockertags))

	assert.True(t, acc.CheckTaggedValue("cache", uint64(1), dockertags))
	assert.True(t, acc.CheckTaggedValue("rss", uint64(2), dockertags))
	assert.True(t, acc.CheckTaggedValue("rss_huge", uint64(3), dockertags))
	assert.True(t, acc.CheckTaggedValue("mapped_file", uint64(4), dockertags))
	assert.True(t, acc.CheckTaggedValue("swap_in", uint64(5), dockertags))
	assert.True(t, acc.CheckTaggedValue("swap_out", uint64(6), dockertags))
	assert.True(t, acc.CheckTaggedValue("page_fault", uint64(7), dockertags))
	assert.True(t, acc.CheckTaggedValue("page_major_fault", uint64(8), dockertags))
	assert.True(t, acc.CheckTaggedValue("inactive_anon", uint64(9), dockertags))
	assert.True(t, acc.CheckTaggedValue("active_anon", uint64(10), dockertags))
	assert.True(t, acc.CheckTaggedValue("inactive_file", uint64(11), dockertags))
	assert.True(t, acc.CheckTaggedValue("active_file", uint64(12), dockertags))
	assert.True(t, acc.CheckTaggedValue("unevictable", uint64(13), dockertags))
	assert.True(t, acc.CheckTaggedValue("memory_limit", uint64(14), dockertags))
	assert.True(t, acc.CheckTaggedValue("total_cache", uint64(15), dockertags))
	assert.True(t, acc.CheckTaggedValue("total_rss", uint64(16), dockertags))
	assert.True(t, acc.CheckTaggedValue("total_rss_huge", uint64(17), dockertags))
	assert.True(t, acc.CheckTaggedValue("total_mapped_file", uint64(18), dockertags))
	assert.True(t, acc.CheckTaggedValue("total_swap_in", uint64(19), dockertags))
	assert.True(t, acc.CheckTaggedValue("total_swap_out", uint64(20), dockertags))
	assert.True(t, acc.CheckTaggedValue("total_page_fault", uint64(21), dockertags))
	assert.True(t, acc.CheckTaggedValue("total_page_major_fault", uint64(22), dockertags))
	assert.True(t, acc.CheckTaggedValue("total_inactive_anon", uint64(23), dockertags))
	assert.True(t, acc.CheckTaggedValue("total_active_anon", uint64(24), dockertags))
	assert.True(t, acc.CheckTaggedValue("total_inactive_file", uint64(25), dockertags))
	assert.True(t, acc.CheckTaggedValue("total_active_file", uint64(26), dockertags))
	assert.True(t, acc.CheckTaggedValue("total_unevictable", uint64(27), dockertags))
}
