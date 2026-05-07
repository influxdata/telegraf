package docker

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/system"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/testutil"
)

func TestInit(t *testing.T) {
	plugin := &Docker{
		Log:              testutil.Logger{},
		PerDeviceInclude: []string{"cpu", "network", "blkio"},
		TotalInclude:     []string{"cpu", "network", "blkio"},
	}
	require.NoError(t, plugin.Init())
}

func TestInitFail(t *testing.T) {
	tests := []struct {
		name      string
		perDevice []string
		total     []string
		expected  string
	}{
		{
			name:      "unsupported perdevice_include",
			perDevice: []string{"nonExistentClass"},
			total:     []string{"cpu"},
			expected:  "unknown choice nonExistentClass",
		},
		{
			name:      "unsupported total_include",
			perDevice: []string{"cpu"},
			total:     []string{"nonExistentClass"},
			expected:  "unknown choice nonExistentClass",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Docker{
				PerDeviceInclude: tt.perDevice,
				TotalInclude:     tt.total,
				Log:              testutil.Logger{},
			}
			require.ErrorContains(t, plugin.Init(), tt.expected)
		})
	}
}

func TestContainerStats(t *testing.T) {
	// Load input data
	data, err := readContainerData("testdata")
	require.NoError(t, err)
	stats := data.stats["123456789"]

	// Setup plugin
	plugin := &Docker{
		PerDeviceInclude: containerMetricClasses,
		TotalInclude:     containerMetricClasses,
		Log:              testutil.Logger{},
	}

	// Collect the data
	tags := map[string]string{
		"container_name":  "redis",
		"container_image": "redis/image",
	}
	var acc testutil.Accumulator
	plugin.parseContainerStats(&stats, &acc, tags, "123456789", "linux")

	// Check the result
	expected := []telegraf.Metric{
		metric.New(
			"docker_container_net",
			map[string]string{
				"container_name":  "redis",
				"container_image": "redis/image",
				"network":         "eth0",
			},
			map[string]interface{}{
				"rx_dropped":   uint64(1),
				"rx_bytes":     uint64(2),
				"rx_errors":    uint64(3),
				"rx_packets":   uint64(2),
				"tx_packets":   uint64(4),
				"tx_dropped":   uint64(1),
				"tx_errors":    uint64(3),
				"tx_bytes":     uint64(4),
				"container_id": "123456789",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_container_net",
			map[string]string{
				"container_name":  "redis",
				"container_image": "redis/image",
				"network":         "eth1",
			},
			map[string]interface{}{
				"rx_dropped":   uint64(5),
				"rx_bytes":     uint64(6),
				"rx_errors":    uint64(7),
				"rx_packets":   uint64(6),
				"tx_dropped":   uint64(5),
				"tx_errors":    uint64(7),
				"tx_bytes":     uint64(8),
				"tx_packets":   uint64(8),
				"container_id": "123456789",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_container_net",
			map[string]string{
				"container_name":  "redis",
				"container_image": "redis/image",
				"network":         "total",
			},
			map[string]interface{}{
				"rx_dropped":   uint64(6),
				"rx_bytes":     uint64(8),
				"rx_errors":    uint64(10),
				"rx_packets":   uint64(8),
				"tx_packets":   uint64(12),
				"tx_dropped":   uint64(6),
				"tx_errors":    uint64(10),
				"tx_bytes":     uint64(12),
				"container_id": "123456789",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_container_blkio",
			map[string]string{
				"container_name":  "redis",
				"container_image": "redis/image",
				"device":          "6:0",
			},
			map[string]interface{}{
				"io_service_bytes_recursive_read": uint64(100),
				"io_serviced_recursive_write":     uint64(101),
				"container_id":                    "123456789",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_container_blkio",
			map[string]string{
				"container_name":  "redis",
				"container_image": "redis/image",
				"device":          "6:1",
			},
			map[string]interface{}{
				"io_serviced_recursive_write": uint64(201),
				"container_id":                "123456789",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_container_blkio",
			map[string]string{
				"container_name":  "redis",
				"container_image": "redis/image",
				"device":          "total",
			},
			map[string]interface{}{
				"io_service_bytes_recursive_read": uint64(100),
				"io_serviced_recursive_write":     uint64(302),
				"container_id":                    "123456789",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_container_mem",
			map[string]string{
				"container_name":  "redis",
				"container_image": "redis/image",
			},
			map[string]interface{}{
				"active_anon":               uint64(0),
				"active_file":               uint64(1),
				"cache":                     uint64(0),
				"container_id":              "123456789",
				"fail_count":                uint64(1),
				"hierarchical_memory_limit": uint64(0),
				"inactive_anon":             uint64(0),
				"inactive_file":             uint64(3),
				"limit":                     uint64(2000),
				"mapped_file":               uint64(0),
				"max_usage":                 uint64(1001),
				"pgfault":                   uint64(2),
				"pgmajfault":                uint64(0),
				"pgpgin":                    uint64(0),
				"pgpgout":                   uint64(0),
				"rss_huge":                  uint64(0),
				"rss":                       uint64(0),
				"total_active_anon":         uint64(0),
				"total_active_file":         uint64(0),
				"total_cache":               uint64(0),
				"total_inactive_anon":       uint64(0),
				"total_inactive_file":       uint64(0),
				"total_mapped_file":         uint64(0),
				"total_pgfault":             uint64(0),
				"total_pgmajfault":          uint64(0),
				"total_pgpgin":              uint64(4),
				"total_pgpgout":             uint64(0),
				"total_rss_huge":            uint64(444),
				"total_rss":                 uint64(44),
				"total_unevictable":         uint64(0),
				"total_writeback":           uint64(55),
				"unevictable":               uint64(0),
				"usage_percent":             float64(55.55),
				"usage":                     uint64(1111),
				"writeback":                 uint64(0),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_container_cpu",
			map[string]string{
				"container_name":  "redis",
				"container_image": "redis/image",
				"cpu":             "cpu0",
			},
			map[string]interface{}{
				"usage_total":  uint64(1),
				"container_id": "123456789",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_container_cpu",
			map[string]string{
				"container_name":  "redis",
				"container_image": "redis/image",
				"cpu":             "cpu1",
			},
			map[string]interface{}{
				"usage_total":  uint64(1002),
				"container_id": "123456789",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_container_cpu",
			map[string]string{
				"container_name":  "redis",
				"container_image": "redis/image",
				"cpu":             "cpu-total",
			},
			map[string]interface{}{
				"usage_total":                  uint64(500),
				"usage_in_usermode":            uint64(100),
				"usage_in_kernelmode":          uint64(200),
				"usage_system":                 uint64(100),
				"throttling_periods":           uint64(1),
				"throttling_throttled_periods": uint64(0),
				"throttling_throttled_time":    uint64(0),
				"usage_percent":                float64(400.0),
				"container_id":                 "123456789",
			},
			time.Unix(0, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestMemoryExcludesCache(t *testing.T) {
	tests := []struct {
		name     string
		override map[string]uint64
		expected []telegraf.Metric
	}{
		{
			name: "pre_19_03",
			override: map[string]uint64{
				"cache":               16,
				"total_inactive_file": 7,
				"inactive_file":       9,
			},
			expected: []telegraf.Metric{
				metric.New(
					"docker_container_mem",
					map[string]string{
						"container_name":  "redis",
						"container_image": "redis/image",
					},
					map[string]interface{}{
						"active_anon":               uint64(0),
						"active_file":               uint64(1),
						"cache":                     uint64(16),
						"container_id":              "123456789",
						"fail_count":                uint64(1),
						"hierarchical_memory_limit": uint64(0),
						"inactive_anon":             uint64(0),
						"inactive_file":             uint64(9),
						"limit":                     uint64(2000),
						"mapped_file":               uint64(0),
						"max_usage":                 uint64(1001),
						"pgfault":                   uint64(2),
						"pgmajfault":                uint64(0),
						"pgpgin":                    uint64(0),
						"pgpgout":                   uint64(0),
						"rss_huge":                  uint64(0),
						"rss":                       uint64(0),
						"total_active_anon":         uint64(0),
						"total_active_file":         uint64(0),
						"total_cache":               uint64(0),
						"total_inactive_anon":       uint64(0),
						"total_inactive_file":       uint64(7),
						"total_mapped_file":         uint64(0),
						"total_pgfault":             uint64(0),
						"total_pgmajfault":          uint64(0),
						"total_pgpgin":              uint64(4),
						"total_pgpgout":             uint64(0),
						"total_rss_huge":            uint64(444),
						"total_rss":                 uint64(44),
						"total_unevictable":         uint64(0),
						"total_writeback":           uint64(55),
						"unevictable":               uint64(0),
						"usage_percent":             float64(54.75), // 1095 / 2000
						"usage":                     uint64(1095),
						"writeback":                 uint64(0),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "cgroup_v1",
			override: map[string]uint64{
				"total_inactive_file": 7,
				"inactive_file":       9,
			},
			expected: []telegraf.Metric{
				metric.New(
					"docker_container_mem",
					map[string]string{
						"container_name":  "redis",
						"container_image": "redis/image",
					},
					map[string]interface{}{
						"active_anon":               uint64(0),
						"active_file":               uint64(1),
						"container_id":              "123456789",
						"fail_count":                uint64(1),
						"hierarchical_memory_limit": uint64(0),
						"inactive_anon":             uint64(0),
						"inactive_file":             uint64(9),
						"limit":                     uint64(2000),
						"mapped_file":               uint64(0),
						"max_usage":                 uint64(1001),
						"pgfault":                   uint64(2),
						"pgmajfault":                uint64(0),
						"pgpgin":                    uint64(0),
						"pgpgout":                   uint64(0),
						"rss_huge":                  uint64(0),
						"rss":                       uint64(0),
						"total_active_anon":         uint64(0),
						"total_active_file":         uint64(0),
						"total_cache":               uint64(0),
						"total_inactive_anon":       uint64(0),
						"total_inactive_file":       uint64(7),
						"total_mapped_file":         uint64(0),
						"total_pgfault":             uint64(0),
						"total_pgmajfault":          uint64(0),
						"total_pgpgin":              uint64(4),
						"total_pgpgout":             uint64(0),
						"total_rss_huge":            uint64(444),
						"total_rss":                 uint64(44),
						"total_unevictable":         uint64(0),
						"total_writeback":           uint64(55),
						"unevictable":               uint64(0),
						"usage_percent":             float64(55.2), // 1104 / 2000
						"usage":                     uint64(1104),
						"writeback":                 uint64(0),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "cgroup_v2",
			override: map[string]uint64{
				"inactive_file": 9,
			},
			expected: []telegraf.Metric{
				metric.New(
					"docker_container_mem",
					map[string]string{
						"container_name":  "redis",
						"container_image": "redis/image",
					},
					map[string]interface{}{
						"active_anon":               uint64(0),
						"active_file":               uint64(1),
						"container_id":              "123456789",
						"fail_count":                uint64(1),
						"hierarchical_memory_limit": uint64(0),
						"inactive_anon":             uint64(0),
						"inactive_file":             uint64(9),
						"limit":                     uint64(2000),
						"mapped_file":               uint64(0),
						"max_usage":                 uint64(1001),
						"pgfault":                   uint64(2),
						"pgmajfault":                uint64(0),
						"pgpgin":                    uint64(0),
						"pgpgout":                   uint64(0),
						"rss_huge":                  uint64(0),
						"rss":                       uint64(0),
						"total_active_anon":         uint64(0),
						"total_active_file":         uint64(0),
						"total_cache":               uint64(0),
						"total_inactive_anon":       uint64(0),
						"total_mapped_file":         uint64(0),
						"total_pgfault":             uint64(0),
						"total_pgmajfault":          uint64(0),
						"total_pgpgin":              uint64(4),
						"total_pgpgout":             uint64(0),
						"total_rss_huge":            uint64(444),
						"total_rss":                 uint64(44),
						"total_unevictable":         uint64(0),
						"total_writeback":           uint64(55),
						"unevictable":               uint64(0),
						"usage_percent":             float64(55.1), // 1102 / 2000
						"usage":                     uint64(1102),
						"writeback":                 uint64(0),
					},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Load input data
			data, err := readContainerData("testdata")
			require.NoError(t, err)

			// Patch the statistic data
			stats := data.stats["123456789"]
			delete(stats.MemoryStats.Stats, "cache")
			delete(stats.MemoryStats.Stats, "inactive_file")
			delete(stats.MemoryStats.Stats, "total_inactive_file")
			for k, v := range tt.override {
				stats.MemoryStats.Stats[k] = v
			}

			// Setup plugin
			plugin := &Docker{
				Log: testutil.Logger{},
			}

			// Collect the data and check the result
			tags := map[string]string{
				"container_name":  "redis",
				"container_image": "redis/image",
			}
			var acc testutil.Accumulator
			plugin.parseContainerStats(&stats, &acc, tags, "123456789", "linux")
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
		})
	}
}

func TestWindowsMemoryContainerStats(t *testing.T) {
	// Setup client factory from data
	factory := newFactoryFromFiles("testdata", true)

	// Setup the expected result
	expected := []telegraf.Metric{
		metric.New(
			"docker",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
			},
			map[string]interface{}{
				"n_containers":            int64(108),
				"n_containers_paused":     int64(3),
				"n_containers_running":    int64(98),
				"n_containers_stopped":    int64(6),
				"n_cpus":                  int64(4),
				"n_goroutines":            int64(39),
				"n_images":                int64(199),
				"n_listener_events":       int64(0),
				"n_used_file_descriptors": int64(19),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
			},
			map[string]interface{}{
				"memory_total": int64(3840757760),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
				"unit":           "bytes",
			},
			map[string]interface{}{
				"pool_blocksize": int64(65540),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_data",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
				"unit":           "bytes",
			},
			map[string]interface{}{
				"available": int64(36530000000),
				"total":     int64(107400000000),
				"used":      int64(17300000000),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_metadata",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
				"unit":           "bytes",
			},
			map[string]interface{}{
				"available": int64(2126999999),
				"total":     int64(2146999999),
				"used":      int64(20970000),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_devicemapper",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
				"pool_name":      "docker-8:1-1182287-pool",
			},
			map[string]interface{}{
				"base_device_size_bytes":             int64(10740000000),
				"data_space_available_bytes":         int64(36530000000),
				"data_space_total_bytes":             int64(107400000000),
				"data_space_used_bytes":              int64(17300000000),
				"metadata_space_available_bytes":     int64(2126999999),
				"metadata_space_total_bytes":         int64(2146999999),
				"metadata_space_used_bytes":          int64(20970000),
				"pool_blocksize_bytes":               int64(65540),
				"thin_pool_minimum_free_space_bytes": int64(10740000000),
			},
			time.Unix(0, 0),
		),
	}

	// Setup the plugin
	plugin := &Docker{
		Timeout:   config.Duration(5 * time.Second),
		Log:       testutil.Logger{},
		newClient: factory,
	}
	require.NoError(t, plugin.Init())

	// Start the plugin
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Collect data and test the result
	require.NoError(t, plugin.Gather(&acc))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestContainerLabels(t *testing.T) {
	var tests = []struct {
		name     string
		labels   map[string]string
		include  []string
		exclude  []string
		expected map[string]string
	}{
		{
			name: "nil filters matches all",
			labels: map[string]string{
				"a": "x",
			},
			expected: map[string]string{
				"a": "x",
			},
		},
		{
			name: "empty filters matches all",
			labels: map[string]string{
				"a": "x",
			},
			expected: map[string]string{
				"a": "x",
			},
		},
		{
			name: "must match include",
			labels: map[string]string{
				"a": "x",
				"b": "y",
			},
			include: []string{"a"},
			expected: map[string]string{
				"a": "x",
			},
		},
		{
			name: "must not match exclude",
			labels: map[string]string{
				"a": "x",
				"b": "y",
			},
			exclude: []string{"b"},
			expected: map[string]string{
				"a": "x",
			},
		},
		{
			name: "include glob",
			labels: map[string]string{
				"aa": "x",
				"ab": "y",
				"bb": "z",
			},
			include: []string{"a*"},
			expected: map[string]string{
				"aa": "x",
				"ab": "y",
			},
		},
		{
			name: "exclude glob",
			labels: map[string]string{
				"aa": "x",
				"ab": "y",
				"bb": "z",
			},
			exclude: []string{"a*"},
			expected: map[string]string{
				"bb": "z",
			},
		},
		{
			name: "excluded and includes",
			labels: map[string]string{
				"aa": "x",
				"ab": "y",
				"bb": "z",
			},
			include: []string{"a*"},
			exclude: []string{"*b"},
			expected: map[string]string{
				"aa": "x",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup client factory and override container list data
			data, err := readContainerData("testdata")
			require.NoError(t, err)
			c := data.summaries[0]
			c.Labels = tt.labels
			c.State = "running"
			data.summaries = []container.Summary{c}
			factory := func(string, *tls.Config) (dockerClient, error) {
				return newClientFromData(data, false), nil
			}

			// Setup plugin
			plugin := &Docker{
				LabelInclude: tt.include,
				LabelExclude: tt.exclude,
				TotalInclude: []string{"cpu"},
				Log:          testutil.Logger{},
				newClient:    factory,
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Collect data and check result
			require.NoError(t, acc.GatherError(plugin.Gather))
			var actual map[string]string
			for _, mt := range acc.Metrics {
				if mt.Measurement == "docker_container_cpu" {
					actual = mt.Tags
					break
				}
			}
			require.Subset(t, actual, tt.expected)
		})
	}
}

func TestContainerNames(t *testing.T) {
	var tests = []struct {
		name       string
		containers [][]string
		include    []string
		exclude    []string
		expected   []string
	}{
		{
			name:     "nil filters matches all",
			expected: []string{"etcd", "etcd2", "acme", "acme-test", "foo"},
		},
		{
			name:     "empty filters matches all",
			expected: []string{"etcd", "etcd2", "acme", "acme-test", "foo"},
		},
		{
			name:     "match all containers",
			include:  []string{"*"},
			expected: []string{"etcd", "etcd2", "acme", "acme-test", "foo"},
		},
		{
			name:     "include prefix match",
			include:  []string{"etc*"},
			expected: []string{"etcd", "etcd2"},
		},
		{
			name:     "exact match",
			include:  []string{"etcd"},
			expected: []string{"etcd"},
		},
		{
			name:     "star matches zero length",
			include:  []string{"etcd2*"},
			expected: []string{"etcd2"},
		},
		{
			name:     "exclude matches all",
			exclude:  []string{"etc*"},
			expected: []string{"acme", "acme-test", "foo"},
		},
		{
			name:     "exclude single",
			exclude:  []string{"etcd"},
			expected: []string{"etcd2", "acme", "acme-test", "foo"},
		},
		{
			name:    "exclude all",
			include: []string{"*"},
			exclude: []string{"*"},
		},
		{
			name:     "exclude item matching include",
			include:  []string{"acme*"},
			exclude:  []string{"*test*"},
			expected: []string{"acme"},
		},
		{
			name:     "exclude item no wildcards",
			include:  []string{"acme*"},
			exclude:  []string{"test"},
			expected: []string{"acme", "acme-test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup client factory
			factory := newFactoryFromFiles("testdata", false)

			// Setup plugin
			plugin := &Docker{
				ContainerInclude: tt.include,
				ContainerExclude: tt.exclude,
				Log:              testutil.Logger{},
				newClient:        factory,
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Collect data and check the results
			require.NoError(t, acc.GatherError(plugin.Gather))
			actual := make([]string, 0)
			for _, mt := range acc.Metrics {
				if name, ok := mt.Tags["container_name"]; ok {
					actual = append(actual, name)
				}
			}
			require.Subset(t, tt.expected, actual)
		})
	}
}

func TestContainerStatus(t *testing.T) {
	var tests = []struct {
		name     string
		now      time.Time
		started  *string
		finished *string
		expected []telegraf.Metric
	}{
		{
			name: "finished_at is zero value",
			now:  time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			expected: []telegraf.Metric{
				metric.New(
					"docker_container_status",
					map[string]string{
						"container_name":    "etcd",
						"container_image":   "quay.io/coreos/etcd",
						"container_version": "v3.3.25",
						"engine_host":       "absol",
						"label1":            "test_value_1",
						"label2":            "test_value_2",
						"server_version":    "17.09.0-ce",
						"container_status":  "running",
						"source":            "e2173b9478a6",
					},
					map[string]interface{}{
						"oomkilled":     false,
						"pid":           1234,
						"restart_count": 0,
						"exitcode":      0,
						"container_id":  "e2173b9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296b7dfb",
						"started_at":    time.Date(2018, 6, 14, 5, 48, 53, 266176036, time.UTC).UnixNano(),
						"uptime_ns":     int64(3 * time.Minute),
					},
					time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
				),
			},
		},
		{
			name:     "finished_at is non-zero value",
			now:      time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			finished: new("2018-06-14T05:53:53.266176036Z"),
			expected: []telegraf.Metric{
				metric.New(
					"docker_container_status",
					map[string]string{
						"container_name":    "etcd",
						"container_image":   "quay.io/coreos/etcd",
						"container_version": "v3.3.25",
						"engine_host":       "absol",
						"label1":            "test_value_1",
						"label2":            "test_value_2",
						"server_version":    "17.09.0-ce",
						"container_status":  "running",
						"source":            "e2173b9478a6",
					},
					map[string]interface{}{
						"oomkilled":     false,
						"pid":           1234,
						"exitcode":      0,
						"restart_count": 0,
						"container_id":  "e2173b9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296b7dfb",
						"started_at":    time.Date(2018, 6, 14, 5, 48, 53, 266176036, time.UTC).UnixNano(),
						"finished_at":   time.Date(2018, 6, 14, 5, 53, 53, 266176036, time.UTC).UnixNano(),
						"uptime_ns":     int64(5 * time.Minute),
					},
					time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
				),
			},
		},
		{
			name:     "started_at is zero value",
			now:      time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
			started:  new(""),
			finished: new("2018-06-14T05:53:53.266176036Z"),
			expected: []telegraf.Metric{
				metric.New(
					"docker_container_status",
					map[string]string{
						"container_name":    "etcd",
						"container_image":   "quay.io/coreos/etcd",
						"container_version": "v3.3.25",
						"engine_host":       "absol",
						"label1":            "test_value_1",
						"label2":            "test_value_2",
						"server_version":    "17.09.0-ce",
						"container_status":  "running",
						"source":            "e2173b9478a6",
					},
					map[string]interface{}{
						"oomkilled":     false,
						"pid":           1234,
						"exitcode":      0,
						"restart_count": 0,
						"container_id":  "e2173b9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296b7dfb",
						"finished_at":   time.Date(2018, 6, 14, 5, 53, 53, 266176036, time.UTC).UnixNano(),
					},
					time.Date(2018, 6, 14, 5, 51, 53, 266176036, time.UTC),
				),
			},
		},
		{
			name:     "container has been restarted",
			now:      time.Date(2019, 1, 1, 0, 0, 3, 0, time.UTC),
			started:  new("2019-01-01T00:00:02Z"),
			finished: new("2019-01-01T00:00:01Z"),
			expected: []telegraf.Metric{
				metric.New(
					"docker_container_status",
					map[string]string{
						"container_name":    "etcd",
						"container_image":   "quay.io/coreos/etcd",
						"container_version": "v3.3.25",
						"engine_host":       "absol",
						"label1":            "test_value_1",
						"label2":            "test_value_2",
						"server_version":    "17.09.0-ce",
						"container_status":  "running",
						"source":            "e2173b9478a6",
					},
					map[string]interface{}{
						"oomkilled":     false,
						"pid":           1234,
						"exitcode":      0,
						"restart_count": 0,
						"container_id":  "e2173b9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296b7dfb",
						"started_at":    time.Date(2019, 1, 1, 0, 0, 2, 0, time.UTC).UnixNano(),
						"finished_at":   time.Date(2019, 1, 1, 0, 0, 1, 0, time.UTC).UnixNano(),
						"uptime_ns":     int64(1 * time.Second),
					},
					time.Date(2019, 1, 1, 0, 0, 3, 0, time.UTC),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the time
			now = func() time.Time { return tt.now }
			defer func() { now = time.Now }()

			// Setup client factory and override values with test-data
			data, err := readContainerData("testdata")
			require.NoError(t, err)
			data.summaries = data.summaries[:1]
			if tt.started != nil {
				data.inspection.ContainerJSONBase.State.StartedAt = *tt.started
			}
			if tt.finished != nil {
				data.inspection.ContainerJSONBase.State.FinishedAt = *tt.finished
			}
			factory := func(string, *tls.Config) (dockerClient, error) {
				return newClientFromData(data, false), nil
			}

			// Setup plugin
			plugin := &Docker{
				IncludeSourceTag: true,
				Log:              testutil.Logger{},
				newClient:        factory,
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Collect data and check the result
			require.NoError(t, acc.GatherError(plugin.Gather))
			testutil.RequireMetricsSubset(t, tt.expected, acc.GetTelegrafMetrics())
		})
	}
}

func TestGatherInfo(t *testing.T) {
	// Setup client factory and override values with test-data
	factory := newFactoryFromFiles("testdata", false)

	// Setup plugin
	plugin := &Docker{
		TagEnvironment: []string{"ENVVAR1", "ENVVAR2", "ENVVAR3", "ENVVAR5",
			"ENVVAR6", "ENVVAR7", "ENVVAR8", "ENVVAR9"},
		PerDeviceInclude: []string{"cpu", "network", "blkio"},
		TotalInclude:     []string{"cpu", "blkio", "network"},
		Log:              testutil.Logger{},
		newClient:        factory,
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Define expected result
	expected := []telegraf.Metric{
		metric.New(
			"docker",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
			},
			map[string]interface{}{
				"n_listener_events":       int(0),
				"n_cpus":                  int(4),
				"n_used_file_descriptors": int(19),
				"n_containers":            int(108),
				"n_containers_running":    int(98),
				"n_containers_stopped":    int(6),
				"n_containers_paused":     int(3),
				"n_images":                int(199),
				"n_goroutines":            int(39),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
			},
			map[string]interface{}{
				"memory_total": int64(3840757760),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
				"unit":           "bytes",
			},
			map[string]interface{}{
				"pool_blocksize": int64(65540),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_data",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
				"unit":           "bytes",
			},
			map[string]interface{}{
				"used":      int64(17300000000),
				"total":     int64(107400000000),
				"available": int64(36530000000),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_metadata",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
				"unit":           "bytes",
			},
			map[string]interface{}{
				"used":      int64(20970000),
				"total":     int64(2146999999),
				"available": int64(2126999999),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_devicemapper",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
				"pool_name":      "docker-8:1-1182287-pool",
			},
			map[string]interface{}{
				"base_device_size_bytes":             int64(10740000000),
				"pool_blocksize_bytes":               int64(65540),
				"data_space_used_bytes":              int64(17300000000),
				"data_space_total_bytes":             int64(107400000000),
				"data_space_available_bytes":         int64(36530000000),
				"metadata_space_used_bytes":          int64(20970000),
				"metadata_space_total_bytes":         int64(2146999999),
				"metadata_space_available_bytes":     int64(2126999999),
				"thin_pool_minimum_free_space_bytes": int64(10740000000),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_container_cpu",
			map[string]string{
				"container_name":    "etcd2",
				"container_image":   "quay.io:4443/coreos/etcd",
				"cpu":               "cpu3",
				"container_version": "v3.3.25",
				"engine_host":       "absol",
				"ENVVAR1":           "loremipsum",
				"ENVVAR2":           "dolorsitamet",
				"ENVVAR3":           "=ubuntu:10.04",
				"ENVVAR7":           "ENVVAR8=ENVVAR9",
				"label1":            "test_value_1",
				"label2":            "test_value_2",
				"server_version":    "17.09.0-ce",
				"container_status":  "running",
			},
			map[string]interface{}{
				"usage_total":  uint64(1231652),
				"container_id": "b7dfbb9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296e2173",
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_container_mem",
			map[string]string{
				"engine_host":       "absol",
				"container_name":    "etcd2",
				"container_image":   "quay.io:4443/coreos/etcd",
				"container_version": "v3.3.25",
				"ENVVAR1":           "loremipsum",
				"ENVVAR2":           "dolorsitamet",
				"ENVVAR3":           "=ubuntu:10.04",
				"ENVVAR7":           "ENVVAR8=ENVVAR9",
				"label1":            "test_value_1",
				"label2":            "test_value_2",
				"server_version":    "17.09.0-ce",
				"container_status":  "running",
			},
			map[string]interface{}{
				"container_id":  "b7dfbb9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296e2173",
				"limit":         uint64(18935443456),
				"max_usage":     uint64(0),
				"usage":         uint64(0),
				"usage_percent": float64(0),
			},
			time.Unix(0, 0),
		),
	}

	// Collect data and check the result
	require.NoError(t, acc.GatherError(plugin.Gather))
	testutil.RequireMetricsSubset(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestDockerGatherSwarmInfo(t *testing.T) {
	// Setup client factory
	factory := newFactoryFromFiles("testdata", false)

	// Setup plugin
	plugin := &Docker{
		GatherServices: true,
		Log:            testutil.Logger{},
		newClient:      factory,
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Define the expected result
	expected := []telegraf.Metric{
		metric.New(
			"docker_swarm",
			map[string]string{
				"service_id":   "qolkls9g5iasdiuihcyz9rnx2",
				"service_name": "test1",
				"service_mode": "replicated",
			},
			map[string]interface{}{
				"tasks_running": int(2),
				"tasks_desired": uint64(2),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_swarm",
			map[string]string{
				"service_id":   "qolkls9g5iasdiuihcyz9rn3",
				"service_name": "test2",
				"service_mode": "global",
			},
			map[string]interface{}{
				"tasks_running": int(1),
				"tasks_desired": uint64(1),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_swarm",
			map[string]string{
				"service_id":   "rfmqydhe8cluzl9hayyrhw5ga",
				"service_name": "test3",
				"service_mode": "replicated_job",
			},
			map[string]interface{}{
				"tasks_running":     int(0),
				"max_concurrent":    uint64(2),
				"total_completions": uint64(2),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_swarm",
			map[string]string{
				"service_id":   "mp50lo68vqgkory4e26ts8f9d",
				"service_name": "test4",
				"service_mode": "global_job",
			},
			map[string]interface{}{
				"tasks_running": int(0),
			},
			time.Unix(0, 0),
		),
	}

	// Collect data and check the result
	require.NoError(t, acc.GatherError(plugin.Gather))
	testutil.RequireMetricsSubset(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestContainerStateFilter(t *testing.T) {
	var tests = []struct {
		name     string
		include  []string
		exclude  []string
		expected []string
	}{
		{
			name:     "default",
			expected: []string{"running"},
		},
		{
			name:     "include running",
			include:  []string{"running"},
			expected: []string{"running"},
		},
		{
			name:     "include glob",
			include:  []string{"r*"},
			expected: []string{"restarting", "running", "removing"},
		},
		{
			name:     "include all",
			include:  []string{"*"},
			expected: []string{"created", "restarting", "running", "removing", "paused", "exited", "dead"},
		},
		{
			name:    "exclude all",
			exclude: []string{"*"},
		},
		{
			name:     "exclude exited",
			include:  []string{"*"},
			exclude:  []string{"exited"},
			expected: []string{"created", "restarting", "running", "removing", "paused", "dead"},
		},
	}

	for _, tt := range tests {
		containerStates := []string{"created", "restarting", "running", "removing", "paused", "exited", "dead"}

		t.Run(tt.name, func(t *testing.T) {
			// Setup client factory
			data, err := readContainerData("testdata")
			require.NoError(t, err)
			// Get an ID to use for gather to complete
			var id string
			for k := range data.stats {
				id = k
				break
			}
			// Fake states data
			data.summaries = make([]container.Summary, 0, len(containerStates))
			for _, v := range containerStates {
				data.summaries = append(data.summaries, container.Summary{
					ID:    id,
					Names: []string{v},
					State: v,
				})
			}
			factory := func(string, *tls.Config) (dockerClient, error) {
				return newClientFromData(data, false), nil
			}

			// Setup plugin
			plugin := &Docker{
				ContainerStateInclude: tt.include,
				ContainerStateExclude: tt.exclude,
				Log:                   testutil.Logger{},
				newClient:             factory,
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Collect data and check the result
			require.NoError(t, acc.GatherError(plugin.Gather), data.summaries)
			actual := make([]string, 0, acc.NMetrics())
			for _, mt := range acc.Metrics {
				if name, ok := mt.Tags["container_name"]; ok {
					actual = append(actual, name)
				}
			}
			require.Subset(t, actual, tt.expected)
		})
	}
}

func TestContainerListRequestsAllContainers(t *testing.T) {
	// Setup factory that records the container-list options
	var actual container.ListOptions
	factory := func(string, *tls.Config) (dockerClient, error) {
		var client mockClient
		client.ContainerListF = func(options container.ListOptions) ([]container.Summary, error) {
			actual = options
			return nil, nil
		}
		return &client, nil
	}

	// Setup plugin
	plugin := &Docker{
		Log:       testutil.Logger{},
		newClient: factory,
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Collect data and check that all containers are requested
	require.NoError(t, acc.GatherError(plugin.Gather))
	require.True(t, actual.All, "ContainerList must request all containers so non-running states can be filtered client-side")
}

func TestNonRunningContainerEmitsStatusMetrics(t *testing.T) {
	// Setup client factory
	factory := func(string, *tls.Config) (dockerClient, error) {
		var client mockClient
		client.InfoF = func() (system.Info, error) {
			return system.Info{
				Name:          "absol",
				ServerVersion: "17.09.0-ce",
			}, nil
		}
		client.ContainerListF = func(container.ListOptions) ([]container.Summary, error) {
			return []container.Summary{
				{
					ID:    "abc123",
					Names: []string{"/stopped-container"},
					State: "exited",
				},
			}, nil
		}
		client.ContainerStatsF = func(string) (container.StatsResponseReader, error) {
			return container.StatsResponseReader{
				Body: io.NopCloser(strings.NewReader("")),
			}, nil
		}
		client.ContainerInspectF = func() (container.InspectResponse, error) {
			return container.InspectResponse{
				Config: &container.Config{},
				ContainerJSONBase: &container.ContainerJSONBase{
					State: &container.State{
						Status:     "exited",
						ExitCode:   137,
						StartedAt:  "2024-01-01T00:00:00Z",
						FinishedAt: "2024-01-01T01:00:00Z",
					},
				},
			}, nil
		}
		return &client, nil
	}

	// Setup client
	plugin := &Docker{
		ContainerStateInclude: []string{"exited"},
		Log:                   testutil.Logger{},
		newClient:             factory,
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Define expected results
	expected := []telegraf.Metric{
		metric.New(
			"docker",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
			},
			map[string]interface{}{
				"n_containers":            int64(0),
				"n_containers_paused":     int64(0),
				"n_containers_running":    int64(0),
				"n_containers_stopped":    int64(0),
				"n_cpus":                  int64(0),
				"n_images":                int64(0),
				"n_listener_events":       int64(0),
				"n_goroutines":            int64(0),
				"n_used_file_descriptors": int64(0),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
			},
			map[string]interface{}{
				"memory_total": int64(0),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_container_status",
			map[string]string{
				"container_name":    "stopped-container",
				"container_image":   "",
				"container_version": "unknown",
				"engine_host":       "absol",
				"server_version":    "17.09.0-ce",
				"container_status":  "exited",
			},
			map[string]interface{}{
				"oomkilled":     false,
				"pid":           0,
				"exitcode":      137,
				"restart_count": 0,
				"container_id":  "abc123",
				"started_at":    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano(),
				"finished_at":   time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC).UnixNano(),
				"uptime_ns":     int64(time.Hour),
			},
			time.Unix(0, 0),
		),
	}

	// Collect and check results
	require.NoError(t, acc.GatherError(plugin.Gather))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestContainerName(t *testing.T) {
	tests := []struct {
		name           string
		containerNames []string
		expected       string
	}{
		{
			name:           "container stats name is preferred",
			containerNames: []string{"/logspout/foo"},
			expected:       "logspout",
		},
		{
			name:           "container stats without name uses container list name",
			containerNames: []string{"/logspout"},
			expected:       "logspout",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup client factory
			data, err := readContainerData("testdata")
			require.NoError(t, err)
			// Get an ID to use for gather to complete
			var id string
			for k := range data.stats {
				id = k
				break
			}
			// Fake the container list
			data.summaries = []container.Summary{
				{
					ID:    id,
					Names: []string{"/logspout"},
					State: "running",
				},
			}
			factory := func(string, *tls.Config) (dockerClient, error) {
				return newClientFromData(data, false), nil
			}

			// Setup plugin
			plugin := &Docker{
				Log:       testutil.Logger{},
				newClient: factory,
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Collect data and check result
			require.NoError(t, acc.GatherError(plugin.Gather))

			for _, mt := range acc.Metrics {
				// This tag is set on all container measurements
				if mt.Measurement == "docker_container_mem" {
					require.Equal(t, tt.expected, mt.Tags["container_name"])
				}
			}
		})
	}
}

func TestHostnameFromID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			name:     "Real ID",
			id:       "565e3a55f5843cfdd4aa5659a1a75e4e78d47f73c3c483f782fe4a26fc8caa07",
			expected: "565e3a55f584",
		},
		{
			name:     "Short ID",
			id:       "shortid123",
			expected: "shortid123",
		},
		{
			name: "No ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, hostnameFromID(tt.id))
		})
	}
}

func TestGatherDiskUsage(t *testing.T) {
	// Setup the client factory
	factory := newFactoryFromFiles("testdata", false)

	// Setup plugin
	plugin := &Docker{
		StorageObjects: []string{"container"},
		Log:            testutil.Logger{},
		newClient:      factory,
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Define the expected result
	expected := []telegraf.Metric{
		metric.New(
			"docker_disk_usage",
			map[string]string{
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
			},
			map[string]interface{}{
				"layers_size": int64(1e10),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_disk_usage",
			map[string]string{
				"container_image":   "some_image",
				"container_version": "1.0.0-alpine",
				"engine_host":       "absol",
				"server_version":    "17.09.0-ce",
				"container_name":    "some_container",
			},
			map[string]interface{}{
				"size_root_fs": int64(123456789),
				"size_rw":      int64(0)},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_disk_usage",
			map[string]string{
				"image_id":       "some_imageid",
				"image_name":     "some_image_tag",
				"image_version":  "1.0.0-alpine",
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
			},
			map[string]interface{}{
				"size":        int64(123456789),
				"shared_size": int64(0)},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_disk_usage",
			map[string]string{
				"image_id":       "7f4a1cc74046",
				"image_name":     "telegraf",
				"image_version":  "latest",
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
			},
			map[string]interface{}{
				"size":        int64(425484494),
				"shared_size": int64(0)},
			time.Unix(0, 0),
		),
		metric.New(
			"docker_disk_usage",
			map[string]string{
				"volume_name":    "some_volume",
				"engine_host":    "absol",
				"server_version": "17.09.0-ce",
			},
			map[string]interface{}{
				"size": int64(123456789),
			},
			time.Unix(0, 0),
		),
	}

	// Collect data and check result
	require.NoError(t, acc.GatherError(plugin.Gather))
	testutil.RequireMetricsSubset(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestPodmanDetection(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		engine   string
		endpoint string
		binary   string
		expected bool
	}{
		{
			name:     "Docker engine",
			version:  "28.3.2",
			engine:   "docker-desktop",
			endpoint: "unix:///var/run/docker.sock",
			binary:   "docker-init",
			expected: false,
		},
		{
			name:     "Real Podman with version number",
			version:  "5.6.1",
			engine:   "localhost.localdomain",
			endpoint: "unix:///run/podman/podman.sock",
			binary:   "crun",
			expected: true,
		},
		{
			name:     "Podman with version string containing podman",
			version:  "4.9.4-podman",
			engine:   "localhost",
			endpoint: "unix:///run/podman/podman.sock",
			expected: true,
		},
		{
			name:     "Podman with podman in name",
			version:  "4.9.4",
			engine:   "podman-machine",
			endpoint: "unix:///var/run/docker.sock",
			expected: true,
		},
		{
			name:     "Podman detected by endpoint",
			version:  "5.2.0",
			engine:   "localhost",
			endpoint: "unix:///run/podman/podman.sock",
			expected: true,
		},
		{
			name:     "Podman with crun runtime",
			version:  "5.0.1",
			engine:   "myhost.local",
			endpoint: "unix:///var/run/container.sock",
			binary:   "crun",
			expected: true,
		},
		{
			name:     "Docker with crun (should not detect as Podman)",
			version:  "20.10.7",
			engine:   "docker-host",
			endpoint: "unix:///var/run/docker.sock",
			binary:   "crun",
			expected: false,
		},
		{
			name:     "Edge case - simple version with generic name",
			version:  "4.8.2",
			engine:   "host",
			endpoint: "unix:///var/run/container.sock",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup client factory and override test-data
			data, err := readContainerData("testdata")
			require.NoError(t, err)
			data.info = system.Info{
				Name:          tt.engine,
				ServerVersion: tt.version,
				InitBinary:    tt.binary,
			}
			factory := func(string, *tls.Config) (dockerClient, error) {
				return newClientFromData(data, false), nil
			}

			// Setup plugin
			plugin := &Docker{
				Endpoint:  tt.endpoint,
				Timeout:   config.Duration(5 * time.Second),
				Log:       testutil.Logger{},
				newClient: factory,
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			require.Equal(t, tt.expected, plugin.isPodman, "Podman detection mismatch")
		})
	}
}

func TestPodmanStatsCache(t *testing.T) {
	// Create a mock Docker plugin configured as Podman
	plugin := &Docker{
		PodmanCacheTTL: config.Duration(60 * time.Second),
		Log:            testutil.Logger{},
		statsCache:     make(map[string]*cachedContainerStats),
		isPodman:       true,
	}

	// Create test stats
	testID := "test-container-123"
	stats1 := &container.StatsResponse{
		CPUStats: container.CPUStats{
			CPUUsage: container.CPUUsage{
				TotalUsage: 1000,
			},
			SystemUsage: 2000,
		},
	}

	stats2 := &container.StatsResponse{
		CPUStats: container.CPUStats{
			CPUUsage: container.CPUUsage{
				TotalUsage: 2000,
			},
			SystemUsage: 4000,
		},
		PreCPUStats: container.CPUStats{}, // Will be filled by fixPodmanCPUStats
	}

	// First call should cache the stats
	plugin.fixPodmanCPUStats(testID, stats1)
	require.Contains(t, plugin.statsCache, testID)
	require.Equal(t, stats1, plugin.statsCache[testID].stats)

	// Second call should use cached stats as PreCPUStats
	plugin.fixPodmanCPUStats(testID, stats2)
	require.Equal(t, stats1.CPUStats, stats2.PreCPUStats)

	// Test cache cleanup
	plugin.statsCache["old-container"] = &cachedContainerStats{
		stats:     stats1,
		timestamp: time.Now().Add(-3 * time.Hour),
	}
	plugin.cleanupStaleCache()
	require.NotContains(t, plugin.statsCache, "old-container")
	require.Contains(t, plugin.statsCache, testID)
}

func TestStartupErrorBehaviorError(t *testing.T) {
	// Test that model.Start returns error when Ping fails with default "error" behavior
	// Uses the startup-error-behavior framework (TSD-006)
	plugin := &Docker{
		Timeout: config.Duration(100 * time.Millisecond),
		newClient: func(string, *tls.Config) (dockerClient, error) {
			return &mockClient{
				PingF: func() (types.Ping, error) {
					return types.Ping{}, errors.New("connection refused")
				},
			}, nil
		},
	}
	model := models.NewRunningInput(plugin, &models.InputConfig{
		Name:  "docker",
		Alias: "error-test",
	})
	model.StartupErrors.Set(0)
	require.NoError(t, model.Init())

	// Starting the plugin will fail with an error because Ping fails
	var acc testutil.Accumulator
	require.ErrorContains(t, model.Start(&acc), "failed to ping Docker daemon")
	model.Stop()
}

func TestStartupErrorBehaviorIgnore(t *testing.T) {
	// Test that model.Start returns fatal error with "ignore" behavior when Ping fails
	plugin := &Docker{
		Timeout: config.Duration(100 * time.Millisecond),
		newClient: func(string, *tls.Config) (dockerClient, error) {
			return &mockClient{
				PingF: func() (types.Ping, error) {
					return types.Ping{}, errors.New("connection refused")
				},
			}, nil
		},
	}
	model := models.NewRunningInput(plugin, &models.InputConfig{
		Name:                 "docker",
		Alias:                "ignore-test",
		StartupErrorBehavior: "ignore",
	})
	model.StartupErrors.Set(0)
	require.NoError(t, model.Init())

	// Starting the plugin will fail and model should convert to fatal error
	var acc testutil.Accumulator
	require.ErrorContains(t, model.Start(&acc), "failed to ping Docker daemon")
	model.Stop()
}

func TestStartSuccess(t *testing.T) {
	// Test that Start succeeds when Docker is available
	plugin := &Docker{
		Timeout: config.Duration(5 * time.Second),
		newClient: func(string, *tls.Config) (dockerClient, error) {
			return &mockClient{
				InfoF: func() (system.Info, error) {
					return system.Info{
						Name:          "docker-desktop",
						ServerVersion: "20.10.0",
					}, nil
				},
			}, nil
		},
		newEnvClient: func() (dockerClient, error) {
			return nil, errors.New("not using env client")
		},
	}
	model := models.NewRunningInput(plugin, &models.InputConfig{
		Name:  "docker",
		Alias: "success-test",
	})
	model.StartupErrors.Set(0)
	require.NoError(t, model.Init())

	var acc testutil.Accumulator
	require.NoError(t, model.Start(&acc))
	model.Stop()
}

// Internal

type mockClient struct {
	InfoF             func() (system.Info, error)
	ContainerListF    func(options container.ListOptions) ([]container.Summary, error)
	ContainerStatsF   func(containerID string) (container.StatsResponseReader, error)
	ContainerInspectF func() (container.InspectResponse, error)
	ServiceListF      func() ([]swarm.Service, error)
	TaskListF         func() ([]swarm.Task, error)
	NodeListF         func() ([]swarm.Node, error)
	DiskUsageF        func() (types.DiskUsage, error)
	ClientVersionF    func() string
	PingF             func() (types.Ping, error)
	CloseF            func() error
}

func (c *mockClient) Info(context.Context) (system.Info, error) {
	if c.InfoF == nil {
		return system.Info{}, nil
	}
	return c.InfoF()
}

func (c *mockClient) ContainerList(_ context.Context, options container.ListOptions) ([]container.Summary, error) {
	if c.ContainerListF == nil {
		return nil, errors.New("not implemented")
	}
	return c.ContainerListF(options)
}

func (c *mockClient) ContainerStats(_ context.Context, containerID string, _ bool) (container.StatsResponseReader, error) {
	if c.ContainerStatsF == nil {
		return container.StatsResponseReader{}, errors.New("not implemented")
	}
	return c.ContainerStatsF(containerID)
}

func (c *mockClient) ContainerInspect(context.Context, string) (container.InspectResponse, error) {
	if c.ContainerInspectF == nil {
		return container.InspectResponse{
			ContainerJSONBase: &container.ContainerJSONBase{},
		}, nil
	}
	return c.ContainerInspectF()
}

func (c *mockClient) ServiceList(context.Context, swarm.ServiceListOptions) ([]swarm.Service, error) {
	if c.ServiceListF == nil {
		return nil, errors.New("not implemented")
	}
	return c.ServiceListF()
}

func (c *mockClient) TaskList(context.Context, swarm.TaskListOptions) ([]swarm.Task, error) {
	if c.TaskListF == nil {
		return nil, errors.New("not implemented")
	}
	return c.TaskListF()
}

func (c *mockClient) NodeList(context.Context, swarm.NodeListOptions) ([]swarm.Node, error) {
	if c.NodeListF == nil {
		return nil, errors.New("not implemented")
	}
	return c.NodeListF()
}

func (c *mockClient) DiskUsage(context.Context, types.DiskUsageOptions) (types.DiskUsage, error) {
	if c.DiskUsageF == nil {
		return types.DiskUsage{}, errors.New("not implemented")
	}
	return c.DiskUsageF()
}

func (c *mockClient) ClientVersion() string {
	if c.ClientVersionF == nil {
		return "1.43"
	}
	return c.ClientVersionF()
}

func (c *mockClient) Ping(context.Context) (types.Ping, error) {
	if c.PingF == nil {
		return types.Ping{}, nil
	}
	return c.PingF()
}

func (c *mockClient) Close() error {
	if c.CloseF == nil {
		return nil
	}
	return c.CloseF()
}

type containerData struct {
	info         system.Info
	summaries    []container.Summary
	stats        map[string]container.StatsResponse
	statsWindows map[string]container.StatsResponse
	inspection   container.InspectResponse
	services     []swarm.Service
	tasks        []swarm.Task
	nodes        []swarm.Node
	disk         types.DiskUsage
}

func readContainerData(path string) (*containerData, error) {
	var in containerData

	// Read info
	if _, err := os.Stat(filepath.Join(path, "info.json")); err == nil {
		buf, err := os.ReadFile(filepath.Join(path, "info.json"))
		if err != nil {
			return nil, fmt.Errorf("reading info failed: %w", err)
		}
		if err := json.Unmarshal(buf, &in.info); err != nil {
			return nil, fmt.Errorf("parsing info failed: %w", err)
		}
	}

	// Read container list
	if _, err := os.Stat(filepath.Join(path, "list.json")); err == nil {
		buf, err := os.ReadFile(filepath.Join(path, "list.json"))
		if err != nil {
			return nil, fmt.Errorf("reading container list failed: %w", err)
		}
		if err := json.Unmarshal(buf, &in.summaries); err != nil {
			return nil, fmt.Errorf("parsing container list failed: %w", err)
		}
	}

	// Read container statistics data
	matches, err := filepath.Glob(filepath.Join(path, "stats_*.json"))
	if err != nil {
		return nil, fmt.Errorf("matching stats failed: %w", err)
	}
	in.stats = make(map[string]container.StatsResponse, len(matches))
	in.statsWindows = make(map[string]container.StatsResponse)
	for _, fn := range matches {
		buf, err := os.ReadFile(fn)
		if err != nil {
			return nil, fmt.Errorf("reading stats %q failed: %w", fn, err)
		}
		var stats container.StatsResponse
		if err := json.Unmarshal(buf, &stats); err != nil {
			return nil, fmt.Errorf("parsing stats %q failed: %w", fn, err)
		}
		id := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(fn), "stats_"), ".json")
		if strings.HasPrefix(id, "windows_") {
			in.statsWindows[strings.TrimPrefix(id, "windows_")] = stats
		} else {
			in.stats[id] = stats
		}
	}

	// Read container inspection data
	if _, err := os.Stat(filepath.Join(path, "inspect.json")); err == nil {
		buf, err := os.ReadFile(filepath.Join(path, "inspect.json"))
		if err != nil {
			return nil, fmt.Errorf("reading inspection data failed: %w", err)
		}
		if err := json.Unmarshal(buf, &in.inspection); err != nil {
			return nil, fmt.Errorf("parsing inspection data failed: %w", err)
		}
	}

	// Read service data
	if _, err := os.Stat(filepath.Join(path, "services.json")); err == nil {
		buf, err := os.ReadFile(filepath.Join(path, "services.json"))
		if err != nil {
			return nil, fmt.Errorf("reading services failed: %w", err)
		}
		if err := json.Unmarshal(buf, &in.services); err != nil {
			return nil, fmt.Errorf("parsing services failed: %w", err)
		}
	}

	// Read task data
	if _, err := os.Stat(filepath.Join(path, "tasks.json")); err == nil {
		buf, err := os.ReadFile(filepath.Join(path, "tasks.json"))
		if err != nil {
			return nil, fmt.Errorf("reading tasks failed: %w", err)
		}
		if err := json.Unmarshal(buf, &in.tasks); err != nil {
			return nil, fmt.Errorf("parsing tasks failed: %w", err)
		}
	}

	// Read node data
	if _, err := os.Stat(filepath.Join(path, "nodes.json")); err == nil {
		buf, err := os.ReadFile(filepath.Join(path, "nodes.json"))
		if err != nil {
			return nil, fmt.Errorf("reading nodes failed: %w", err)
		}
		if err := json.Unmarshal(buf, &in.nodes); err != nil {
			return nil, fmt.Errorf("parsing nodes failed: %w", err)
		}
	}

	// Read disk usage
	if _, err := os.Stat(filepath.Join(path, "disk.json")); err == nil {
		buf, err := os.ReadFile(filepath.Join(path, "disk.json"))
		if err != nil {
			return nil, fmt.Errorf("reading disk failed: %w", err)
		}
		if err := json.Unmarshal(buf, &in.disk); err != nil {
			return nil, fmt.Errorf("parsing disk failed: %w", err)
		}
	}

	return &in, nil
}

func newClientFromData(data *containerData, windows bool) *mockClient {
	var stats map[string]container.StatsResponse
	if windows {
		stats = data.statsWindows
	} else {
		stats = data.stats
	}

	return &mockClient{
		InfoF: func() (system.Info, error) {
			return data.info, nil
		},
		ContainerListF: func(container.ListOptions) ([]container.Summary, error) {
			return data.summaries, nil
		},
		ContainerStatsF: func(id string) (container.StatsResponseReader, error) {
			s, found := stats[id]
			if !found {
				return container.StatsResponseReader{}, fmt.Errorf("stats for %q not found", id)
			}
			buf, err := json.Marshal(s)
			if err != nil {
				return container.StatsResponseReader{}, fmt.Errorf("encoding stats for %q failed: %w", id, err)
			}
			return container.StatsResponseReader{Body: io.NopCloser(bytes.NewReader(buf))}, nil
		},
		ContainerInspectF: func() (container.InspectResponse, error) {
			return data.inspection, nil
		},
		ServiceListF: func() ([]swarm.Service, error) {
			return data.services, nil
		},
		TaskListF: func() ([]swarm.Task, error) {
			return data.tasks, nil
		},
		NodeListF: func() ([]swarm.Node, error) {
			return data.nodes, nil
		},
		DiskUsageF: func() (types.DiskUsage, error) {
			return data.disk, nil
		},
	}
}

//nolint:unparam // For now 'path' is always 'testdata' but this will change in a follow-up PR
func newFactoryFromFiles(path string, windows bool) func(string, *tls.Config) (dockerClient, error) {
	data, err := readContainerData(path)
	return func(string, *tls.Config) (dockerClient, error) {
		return newClientFromData(data, windows), err
	}
}
