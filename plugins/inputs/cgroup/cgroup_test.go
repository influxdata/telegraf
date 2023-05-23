//go:build linux

package cgroup

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestCgroupStatistics_1(t *testing.T) {
	var acc testutil.Accumulator

	var cg = &CGroup{
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

	require.NoError(t, acc.GatherError(cg.Gather))
	expected := []telegraf.Metric{
		metric.New(
			"cgroup",
			map[string]string{
				"path": "testdata/memory",
			},
			map[string]interface{}{
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
			},

			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestCgroupStatistics_2(t *testing.T) {
	var acc testutil.Accumulator

	var cg = &CGroup{
		Paths: []string{"testdata/cpu"},
		Files: []string{
			"cpuacct.usage_percpu",
			"cpu.stat",
		},
	}

	require.NoError(t, acc.GatherError(cg.Gather))
	expected := []telegraf.Metric{
		metric.New(
			"cgroup",
			map[string]string{
				"path": "testdata/cpu",
			},
			map[string]interface{}{
				"cpu.stat.core_sched.force_idle_usec": int64(0),
				"cpu.stat.system_usec":                int64(103537582650),
				"cpu.stat.usage_usec":                 int64(614953149468),
				"cpu.stat.user_usec":                  int64(511415566817),
				"cpuacct.usage_percpu.0":              int64(-1452543795404),
				"cpuacct.usage_percpu.1":              int64(1376681271659),
				"cpuacct.usage_percpu.2":              int64(1450950799997),
				"cpuacct.usage_percpu.3":              int64(-1473113374257),
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestCgroupStatistics_3(t *testing.T) {
	var acc testutil.Accumulator

	var cg = &CGroup{
		Paths: []string{"testdata/memory/*"},
		Files: []string{"memory.limit_in_bytes"},
	}

	fields := map[string]interface{}{
		"memory.limit_in_bytes": int64(223372036854771712),
	}

	require.NoError(t, acc.GatherError(cg.Gather))
	expected := []telegraf.Metric{
		metric.New(
			"cgroup",
			map[string]string{
				"path": "testdata/memory/group_1",
			},
			fields,
			time.Unix(0, 0),
		),
		metric.New(
			"cgroup",
			map[string]string{
				"path": "testdata/memory/group_2",
			},
			fields,
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestCgroupStatistics_4(t *testing.T) {
	var acc testutil.Accumulator

	var cg = &CGroup{
		Paths: []string{"testdata/memory/*/*", "testdata/memory/group_2"},
		Files: []string{"memory.limit_in_bytes"},
	}

	fields := map[string]interface{}{
		"memory.limit_in_bytes": int64(223372036854771712),
	}

	require.NoError(t, acc.GatherError(cg.Gather))
	expected := []telegraf.Metric{
		metric.New(
			"cgroup",
			map[string]string{
				"path": "testdata/memory/group_1/group_1_1",
			},
			fields,
			time.Unix(0, 0),
		),
		metric.New(
			"cgroup",
			map[string]string{
				"path": "testdata/memory/group_1/group_1_2",
			},
			fields,
			time.Unix(0, 0),
		),
		metric.New(
			"cgroup",
			map[string]string{
				"path": "testdata/memory/group_2/group_1_1",
			},
			fields,
			time.Unix(0, 0),
		),
		metric.New(
			"cgroup",
			map[string]string{
				"path": "testdata/memory/group_2",
			},
			map[string]interface{}{
				"memory.limit_in_bytes": int64(223372036854771712),
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestCgroupStatistics_5(t *testing.T) {
	var acc testutil.Accumulator

	var cg = &CGroup{
		Paths: []string{"testdata/memory/*/group_1_1"},
		Files: []string{"memory.limit_in_bytes"},
	}

	require.NoError(t, acc.GatherError(cg.Gather))
	expected := []telegraf.Metric{
		metric.New(
			"cgroup",
			map[string]string{
				"path": "testdata/memory/group_1/group_1_1",
			},
			map[string]interface{}{
				"memory.limit_in_bytes": int64(223372036854771712),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"cgroup",
			map[string]string{
				"path": "testdata/memory/group_2/group_1_1",
			},
			map[string]interface{}{
				"memory.limit_in_bytes": int64(223372036854771712),
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestCgroupStatistics_6(t *testing.T) {
	var acc testutil.Accumulator

	var cg = &CGroup{
		Paths: []string{"testdata/memory"},
		Files: []string{"memory.us*", "*/memory.kmem.*"},
	}

	require.NoError(t, acc.GatherError(cg.Gather))
	expected := []telegraf.Metric{
		metric.New(
			"cgroup",
			map[string]string{
				"path": "testdata/memory",
			},
			map[string]interface{}{
				"memory.usage_in_bytes":      int64(3513667584),
				"memory.use_hierarchy":       "12-781",
				"memory.kmem.limit_in_bytes": int64(9223372036854771712),
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestCgroupStatistics_7(t *testing.T) {
	var acc testutil.Accumulator

	var cg = &CGroup{
		Paths: []string{"testdata/blkio"},
		Files: []string{"blkio.throttle.io_serviced"},
	}

	require.NoError(t, acc.GatherError(cg.Gather))
	expected := []telegraf.Metric{
		metric.New(
			"cgroup",
			map[string]string{
				"path": "testdata/blkio",
			},
			map[string]interface{}{
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
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestCgroupStatistics_8(t *testing.T) {
	var acc testutil.Accumulator

	var cg = &CGroup{
		Paths:  []string{"testdata/broken"},
		Files:  []string{"malformed.file", "memory.limit_in_bytes"},
		logged: make(map[string]bool),
	}

	require.Error(t, acc.GatherError(cg.Gather))
	require.Len(t, cg.logged, 1)

	expected := []telegraf.Metric{
		metric.New(
			"cgroup",
			map[string]string{"path": "testdata/broken"},
			map[string]interface{}{"memory.limit_in_bytes": int64(1)},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())

	// clear errors so we can check for new errors in next round
	acc.Errors = nil

	require.NoError(t, acc.GatherError(cg.Gather))
	require.Len(t, cg.logged, 1)
}
