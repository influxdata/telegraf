//go:build linux
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

// ======================================================================

var cg7 = &CGroup{
	Paths: []string{"testdata/blkio"},
	Files: []string{"blkio.throttle.io_serviced"},
}

func TestCgroupStatistics_7(t *testing.T) {
	var acc testutil.Accumulator

	err := acc.GatherError(cg7.Gather)
	require.NoError(t, err)

	tags := map[string]string{
		"path": "testdata/blkio",
	}
	fields := map[string]interface{}{
		"blkio.throttle.io_serviced.11:0.Read":  int64(0),
		"blkio.throttle.io_serviced.11:0.Write": int64(0),
		"blkio.throttle.io_serviced.11:0.Sync":  int64(0),
		"blkio.throttle.io_serviced.11:0.Async": int64(0),
		"blkio.throttle.io_serviced.11:0.Total": int64(0),
		"blkio.throttle.io_serviced.8:0.Read":   int64(49134),
		"blkio.throttle.io_serviced.8:0.Write":  int64(216703),
		"blkio.throttle.io_serviced.8:0.Sync":   int64(177906),
		"blkio.throttle.io_serviced.8:0.Async":  int64(87931),
		"blkio.throttle.io_serviced.8:0.Total":  int64(265837),
		"blkio.throttle.io_serviced.7:7.Read":   int64(0),
		"blkio.throttle.io_serviced.7:7.Write":  int64(0),
		"blkio.throttle.io_serviced.7:7.Sync":   int64(0),
		"blkio.throttle.io_serviced.7:7.Async":  int64(0),
		"blkio.throttle.io_serviced.7:7.Total":  int64(0),
		"blkio.throttle.io_serviced.7:6.Read":   int64(0),
		"blkio.throttle.io_serviced.7:6.Write":  int64(0),
		"blkio.throttle.io_serviced.7:6.Sync":   int64(0),
		"blkio.throttle.io_serviced.7:6.Async":  int64(0),
		"blkio.throttle.io_serviced.7:6.Total":  int64(0),
		"blkio.throttle.io_serviced.7:5.Read":   int64(0),
		"blkio.throttle.io_serviced.7:5.Write":  int64(0),
		"blkio.throttle.io_serviced.7:5.Sync":   int64(0),
		"blkio.throttle.io_serviced.7:5.Async":  int64(0),
		"blkio.throttle.io_serviced.7:5.Total":  int64(0),
		"blkio.throttle.io_serviced.7:4.Read":   int64(0),
		"blkio.throttle.io_serviced.7:4.Write":  int64(0),
		"blkio.throttle.io_serviced.7:4.Sync":   int64(0),
		"blkio.throttle.io_serviced.7:4.Async":  int64(0),
		"blkio.throttle.io_serviced.7:4.Total":  int64(0),
		"blkio.throttle.io_serviced.7:3.Read":   int64(0),
		"blkio.throttle.io_serviced.7:3.Write":  int64(0),
		"blkio.throttle.io_serviced.7:3.Sync":   int64(0),
		"blkio.throttle.io_serviced.7:3.Async":  int64(0),
		"blkio.throttle.io_serviced.7:3.Total":  int64(0),
		"blkio.throttle.io_serviced.7:2.Read":   int64(0),
		"blkio.throttle.io_serviced.7:2.Write":  int64(0),
		"blkio.throttle.io_serviced.7:2.Sync":   int64(0),
		"blkio.throttle.io_serviced.7:2.Async":  int64(0),
		"blkio.throttle.io_serviced.7:2.Total":  int64(0),
		"blkio.throttle.io_serviced.7:1.Read":   int64(0),
		"blkio.throttle.io_serviced.7:1.Write":  int64(0),
		"blkio.throttle.io_serviced.7:1.Sync":   int64(0),
		"blkio.throttle.io_serviced.7:1.Async":  int64(0),
		"blkio.throttle.io_serviced.7:1.Total":  int64(0),
		"blkio.throttle.io_serviced.7:0.Read":   int64(0),
		"blkio.throttle.io_serviced.7:0.Write":  int64(0),
		"blkio.throttle.io_serviced.7:0.Sync":   int64(0),
		"blkio.throttle.io_serviced.7:0.Async":  int64(0),
		"blkio.throttle.io_serviced.7:0.Total":  int64(0),
		"blkio.throttle.io_serviced.1:15.Read":  int64(3),
		"blkio.throttle.io_serviced.1:15.Write": int64(0),
		"blkio.throttle.io_serviced.1:15.Sync":  int64(0),
		"blkio.throttle.io_serviced.1:15.Async": int64(3),
		"blkio.throttle.io_serviced.1:15.Total": int64(3),
		"blkio.throttle.io_serviced.1:14.Read":  int64(3),
		"blkio.throttle.io_serviced.1:14.Write": int64(0),
		"blkio.throttle.io_serviced.1:14.Sync":  int64(0),
		"blkio.throttle.io_serviced.1:14.Async": int64(3),
		"blkio.throttle.io_serviced.1:14.Total": int64(3),
		"blkio.throttle.io_serviced.1:13.Read":  int64(3),
		"blkio.throttle.io_serviced.1:13.Write": int64(0),
		"blkio.throttle.io_serviced.1:13.Sync":  int64(0),
		"blkio.throttle.io_serviced.1:13.Async": int64(3),
		"blkio.throttle.io_serviced.1:13.Total": int64(3),
		"blkio.throttle.io_serviced.1:12.Read":  int64(3),
		"blkio.throttle.io_serviced.1:12.Write": int64(0),
		"blkio.throttle.io_serviced.1:12.Sync":  int64(0),
		"blkio.throttle.io_serviced.1:12.Async": int64(3),
		"blkio.throttle.io_serviced.1:12.Total": int64(3),
		"blkio.throttle.io_serviced.1:11.Read":  int64(3),
		"blkio.throttle.io_serviced.1:11.Write": int64(0),
		"blkio.throttle.io_serviced.1:11.Sync":  int64(0),
		"blkio.throttle.io_serviced.1:11.Async": int64(3),
		"blkio.throttle.io_serviced.1:11.Total": int64(3),
		"blkio.throttle.io_serviced.1:10.Read":  int64(3),
		"blkio.throttle.io_serviced.1:10.Write": int64(0),
		"blkio.throttle.io_serviced.1:10.Sync":  int64(0),
		"blkio.throttle.io_serviced.1:10.Async": int64(3),
		"blkio.throttle.io_serviced.1:10.Total": int64(3),
		"blkio.throttle.io_serviced.1:9.Read":   int64(3),
		"blkio.throttle.io_serviced.1:9.Write":  int64(0),
		"blkio.throttle.io_serviced.1:9.Sync":   int64(0),
		"blkio.throttle.io_serviced.1:9.Async":  int64(3),
		"blkio.throttle.io_serviced.1:9.Total":  int64(3),
		"blkio.throttle.io_serviced.1:8.Read":   int64(3),
		"blkio.throttle.io_serviced.1:8.Write":  int64(0),
		"blkio.throttle.io_serviced.1:8.Sync":   int64(0),
		"blkio.throttle.io_serviced.1:8.Async":  int64(3),
		"blkio.throttle.io_serviced.1:8.Total":  int64(3),
		"blkio.throttle.io_serviced.1:7.Read":   int64(3),
		"blkio.throttle.io_serviced.1:7.Write":  int64(0),
		"blkio.throttle.io_serviced.1:7.Sync":   int64(0),
		"blkio.throttle.io_serviced.1:7.Async":  int64(3),
		"blkio.throttle.io_serviced.1:7.Total":  int64(3),
		"blkio.throttle.io_serviced.1:6.Read":   int64(3),
		"blkio.throttle.io_serviced.1:6.Write":  int64(0),
		"blkio.throttle.io_serviced.1:6.Sync":   int64(0),
		"blkio.throttle.io_serviced.1:6.Async":  int64(3),
		"blkio.throttle.io_serviced.1:6.Total":  int64(3),
		"blkio.throttle.io_serviced.1:5.Read":   int64(3),
		"blkio.throttle.io_serviced.1:5.Write":  int64(0),
		"blkio.throttle.io_serviced.1:5.Sync":   int64(0),
		"blkio.throttle.io_serviced.1:5.Async":  int64(3),
		"blkio.throttle.io_serviced.1:5.Total":  int64(3),
		"blkio.throttle.io_serviced.1:4.Read":   int64(3),
		"blkio.throttle.io_serviced.1:4.Write":  int64(0),
		"blkio.throttle.io_serviced.1:4.Sync":   int64(0),
		"blkio.throttle.io_serviced.1:4.Async":  int64(3),
		"blkio.throttle.io_serviced.1:4.Total":  int64(3),
		"blkio.throttle.io_serviced.1:3.Read":   int64(3),
		"blkio.throttle.io_serviced.1:3.Write":  int64(0),
		"blkio.throttle.io_serviced.1:3.Sync":   int64(0),
		"blkio.throttle.io_serviced.1:3.Async":  int64(3),
		"blkio.throttle.io_serviced.1:3.Total":  int64(3),
		"blkio.throttle.io_serviced.1:2.Read":   int64(3),
		"blkio.throttle.io_serviced.1:2.Write":  int64(0),
		"blkio.throttle.io_serviced.1:2.Sync":   int64(0),
		"blkio.throttle.io_serviced.1:2.Async":  int64(3),
		"blkio.throttle.io_serviced.1:2.Total":  int64(3),
		"blkio.throttle.io_serviced.1:1.Read":   int64(3),
		"blkio.throttle.io_serviced.1:1.Write":  int64(0),
		"blkio.throttle.io_serviced.1:1.Sync":   int64(0),
		"blkio.throttle.io_serviced.1:1.Async":  int64(3),
		"blkio.throttle.io_serviced.1:1.Total":  int64(3),
		"blkio.throttle.io_serviced.1:0.Read":   int64(3),
		"blkio.throttle.io_serviced.1:0.Write":  int64(0),
		"blkio.throttle.io_serviced.1:0.Sync":   int64(0),
		"blkio.throttle.io_serviced.1:0.Async":  int64(3),
		"blkio.throttle.io_serviced.1:0.Total":  int64(3),
		"blkio.throttle.io_serviced.Total":      int64(265885),
	}
	acc.AssertContainsTaggedFields(t, "cgroup", fields, tags)
}
