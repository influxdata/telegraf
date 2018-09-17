// +build linux

package cgroup

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var cg1 = &CGroup{
	Paths: []string{"testdata/memory"},
	Files: []string{
		"memory.empty",
		"memory.max_usage_in_bytes",
		"memory.limit_in_bytes",
		"memory.stat",
		"memory.use_hierarchy",
		"notify_on_release",
	},
}

func TestCgroupStatistics_1(t *testing.T) {
	var acc testutil.Accumulator

	err := acc.GatherError(cg1.Gather)
	require.NoError(t, err)

	tags := map[string]string{
		"path": "testdata/memory",
	}
	fields := map[string]interface{}{
		"memory.stat.cache":           int64(1739362304123123123),
		"memory.stat.rss":             int64(1775325184),
		"memory.stat.rss_huge":        int64(778043392),
		"memory.stat.mapped_file":     int64(421036032),
		"memory.stat.dirty":           int64(-307200),
		"memory.max_usage_in_bytes.0": int64(0),
		"memory.max_usage_in_bytes.1": int64(-1),
		"memory.max_usage_in_bytes.2": int64(2),
		"memory.limit_in_bytes":       int64(223372036854771712),
		"memory.use_hierarchy":        "12-781",
		"notify_on_release":           int64(0),
	}
	acc.AssertContainsTaggedFields(t, "cgroup", fields, tags)
}

// ======================================================================

var cg2 = &CGroup{
	Paths: []string{"testdata/cpu"},
	Files: []string{"cpuacct.usage_percpu"},
}

func TestCgroupStatistics_2(t *testing.T) {
	var acc testutil.Accumulator

	err := acc.GatherError(cg2.Gather)
	require.NoError(t, err)

	tags := map[string]string{
		"path": "testdata/cpu",
	}
	fields := map[string]interface{}{
		"cpuacct.usage_percpu.0": int64(-1452543795404),
		"cpuacct.usage_percpu.1": int64(1376681271659),
		"cpuacct.usage_percpu.2": int64(1450950799997),
		"cpuacct.usage_percpu.3": int64(-1473113374257),
	}
	acc.AssertContainsTaggedFields(t, "cgroup", fields, tags)
}

// ======================================================================

var cg3 = &CGroup{
	Paths: []string{"testdata/memory/*"},
	Files: []string{"memory.limit_in_bytes"},
}

func TestCgroupStatistics_3(t *testing.T) {
	var acc testutil.Accumulator

	err := acc.GatherError(cg3.Gather)
	require.NoError(t, err)

	tags := map[string]string{
		"path": "testdata/memory/group_1",
	}
	fields := map[string]interface{}{
		"memory.limit_in_bytes": int64(223372036854771712),
	}
	acc.AssertContainsTaggedFields(t, "cgroup", fields, tags)

	tags = map[string]string{
		"path": "testdata/memory/group_2",
	}
	acc.AssertContainsTaggedFields(t, "cgroup", fields, tags)
}

// ======================================================================

var cg4 = &CGroup{
	Paths: []string{"testdata/memory/*/*", "testdata/memory/group_2"},
	Files: []string{"memory.limit_in_bytes"},
}

func TestCgroupStatistics_4(t *testing.T) {
	var acc testutil.Accumulator

	err := acc.GatherError(cg4.Gather)
	require.NoError(t, err)

	tags := map[string]string{
		"path": "testdata/memory/group_1/group_1_1",
	}
	fields := map[string]interface{}{
		"memory.limit_in_bytes": int64(223372036854771712),
	}
	acc.AssertContainsTaggedFields(t, "cgroup", fields, tags)

	tags = map[string]string{
		"path": "testdata/memory/group_1/group_1_2",
	}
	acc.AssertContainsTaggedFields(t, "cgroup", fields, tags)

	tags = map[string]string{
		"path": "testdata/memory/group_2",
	}
	acc.AssertContainsTaggedFields(t, "cgroup", fields, tags)
}

// ======================================================================

var cg5 = &CGroup{
	Paths: []string{"testdata/memory/*/group_1_1"},
	Files: []string{"memory.limit_in_bytes"},
}

func TestCgroupStatistics_5(t *testing.T) {
	var acc testutil.Accumulator

	err := acc.GatherError(cg5.Gather)
	require.NoError(t, err)

	tags := map[string]string{
		"path": "testdata/memory/group_1/group_1_1",
	}
	fields := map[string]interface{}{
		"memory.limit_in_bytes": int64(223372036854771712),
	}
	acc.AssertContainsTaggedFields(t, "cgroup", fields, tags)

	tags = map[string]string{
		"path": "testdata/memory/group_2/group_1_1",
	}
	acc.AssertContainsTaggedFields(t, "cgroup", fields, tags)
}

// ======================================================================

var cg6 = &CGroup{
	Paths: []string{"testdata/memory"},
	Files: []string{"memory.us*", "*/memory.kmem.*"},
}

func TestCgroupStatistics_6(t *testing.T) {
	var acc testutil.Accumulator

	err := acc.GatherError(cg6.Gather)
	require.NoError(t, err)

	tags := map[string]string{
		"path": "testdata/memory",
	}
	fields := map[string]interface{}{
		"memory.usage_in_bytes":      int64(3513667584),
		"memory.use_hierarchy":       "12-781",
		"memory.kmem.limit_in_bytes": int64(9223372036854771712),
	}
	acc.AssertContainsTaggedFields(t, "cgroup", fields, tags)
}
