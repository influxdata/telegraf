package docker

import (
	"context"
	"crypto/tls"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/require"
)

type MockClient struct {
	InfoF             func(ctx context.Context) (types.Info, error)
	ContainerListF    func(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerStatsF   func(ctx context.Context, containerID string, stream bool) (types.ContainerStats, error)
	ContainerInspectF func(ctx context.Context, containerID string) (types.ContainerJSON, error)
}

func (c *MockClient) Info(ctx context.Context) (types.Info, error) {
	return c.InfoF(ctx)
}

func (c *MockClient) ContainerList(
	ctx context.Context,
	options types.ContainerListOptions,
) ([]types.Container, error) {
	return c.ContainerListF(ctx, options)
}

func (c *MockClient) ContainerStats(
	ctx context.Context,
	containerID string,
	stream bool,
) (types.ContainerStats, error) {
	return c.ContainerStatsF(ctx, containerID, stream)
}

func (c *MockClient) ContainerInspect(
	ctx context.Context,
	containerID string,
) (types.ContainerJSON, error) {
	return c.ContainerInspectF(ctx, containerID)
}

func newClient(host string, tlsConfig *tls.Config) (Client, error) {
	return &MockClient{
		InfoF: func(context.Context) (types.Info, error) {
			return info, nil
		},
		ContainerListF: func(context.Context, types.ContainerListOptions) ([]types.Container, error) {
			return containerList, nil
		},
		ContainerStatsF: func(context.Context, string, bool) (types.ContainerStats, error) {
			return containerStats(), nil
		},
		ContainerInspectF: func(context.Context, string) (types.ContainerJSON, error) {
			return containerInspect, nil
		},
	}, nil
}

func TestDockerGatherContainerStats(t *testing.T) {
	var acc testutil.Accumulator
	stats := testStats()

	tags := map[string]string{
		"container_name":  "redis",
		"container_image": "redis/image",
	}

	gatherContainerStats(stats, &acc, tags, "123456789", true, true, "linux")

	// test docker_container_net measurement
	netfields := map[string]interface{}{
		"rx_dropped":   uint64(1),
		"rx_bytes":     uint64(2),
		"rx_errors":    uint64(3),
		"tx_packets":   uint64(4),
		"tx_dropped":   uint64(1),
		"rx_packets":   uint64(2),
		"tx_errors":    uint64(3),
		"tx_bytes":     uint64(4),
		"container_id": "123456789",
	}
	nettags := copyTags(tags)
	nettags["network"] = "eth0"
	acc.AssertContainsTaggedFields(t, "docker_container_net", netfields, nettags)

	netfields = map[string]interface{}{
		"rx_dropped":   uint64(6),
		"rx_bytes":     uint64(8),
		"rx_errors":    uint64(10),
		"tx_packets":   uint64(12),
		"tx_dropped":   uint64(6),
		"rx_packets":   uint64(8),
		"tx_errors":    uint64(10),
		"tx_bytes":     uint64(12),
		"container_id": "123456789",
	}
	nettags = copyTags(tags)
	nettags["network"] = "total"
	acc.AssertContainsTaggedFields(t, "docker_container_net", netfields, nettags)

	// test docker_blkio measurement
	blkiotags := copyTags(tags)
	blkiotags["device"] = "6:0"
	blkiofields := map[string]interface{}{
		"io_service_bytes_recursive_read": uint64(100),
		"io_serviced_recursive_write":     uint64(101),
		"container_id":                    "123456789",
	}
	acc.AssertContainsTaggedFields(t, "docker_container_blkio", blkiofields, blkiotags)

	blkiotags = copyTags(tags)
	blkiotags["device"] = "total"
	blkiofields = map[string]interface{}{
		"io_service_bytes_recursive_read": uint64(100),
		"io_serviced_recursive_write":     uint64(302),
		"container_id":                    "123456789",
	}
	acc.AssertContainsTaggedFields(t, "docker_container_blkio", blkiofields, blkiotags)

	// test docker_container_mem measurement
	memfields := map[string]interface{}{
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
	}

	acc.AssertContainsTaggedFields(t, "docker_container_mem", memfields, tags)

	// test docker_container_cpu measurement
	cputags := copyTags(tags)
	cputags["cpu"] = "cpu-total"
	cpufields := map[string]interface{}{
		"usage_total":                  uint64(500),
		"usage_in_usermode":            uint64(100),
		"usage_in_kernelmode":          uint64(200),
		"usage_system":                 uint64(100),
		"throttling_periods":           uint64(1),
		"throttling_throttled_periods": uint64(0),
		"throttling_throttled_time":    uint64(0),
		"usage_percent":                float64(400.0),
		"container_id":                 "123456789",
	}
	acc.AssertContainsTaggedFields(t, "docker_container_cpu", cpufields, cputags)

	cputags["cpu"] = "cpu0"
	cpu0fields := map[string]interface{}{
		"usage_total":  uint64(1),
		"container_id": "123456789",
	}
	acc.AssertContainsTaggedFields(t, "docker_container_cpu", cpu0fields, cputags)

	cputags["cpu"] = "cpu1"
	cpu1fields := map[string]interface{}{
		"usage_total":  uint64(1002),
		"container_id": "123456789",
	}
	acc.AssertContainsTaggedFields(t, "docker_container_cpu", cpu1fields, cputags)

	// Those tagged filed should not be present because of offline CPUs
	cputags["cpu"] = "cpu2"
	cpu2fields := map[string]interface{}{
		"usage_total":  uint64(0),
		"container_id": "123456789",
	}
	acc.AssertDoesNotContainsTaggedFields(t, "docker_container_cpu", cpu2fields, cputags)

	cputags["cpu"] = "cpu3"
	cpu3fields := map[string]interface{}{
		"usage_total":  uint64(0),
		"container_id": "123456789",
	}
	acc.AssertDoesNotContainsTaggedFields(t, "docker_container_cpu", cpu3fields, cputags)
}

func TestDocker_WindowsMemoryContainerStats(t *testing.T) {
	var acc testutil.Accumulator

	d := Docker{
		newClient: func(string, *tls.Config) (Client, error) {
			return &MockClient{
				InfoF: func(ctx context.Context) (types.Info, error) {
					return info, nil
				},
				ContainerListF: func(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
					return containerList, nil
				},
				ContainerStatsF: func(ctx context.Context, containerID string, stream bool) (types.ContainerStats, error) {
					return containerStatsWindows(), nil
				},
				ContainerInspectF: func(ctx context.Context, containerID string) (types.ContainerJSON, error) {
					return containerInspect, nil
				},
			}, nil
		},
	}
	err := d.Gather(&acc)
	require.NoError(t, err)
}

func TestDockerGatherLabels(t *testing.T) {
	var gatherLabelsTests = []struct {
		include     []string
		exclude     []string
		expected    []string
		notexpected []string
	}{
		{[]string{}, []string{}, []string{"label1", "label2"}, []string{}},
		{[]string{"*"}, []string{}, []string{"label1", "label2"}, []string{}},
		{[]string{"lab*"}, []string{}, []string{"label1", "label2"}, []string{}},
		{[]string{"label1"}, []string{}, []string{"label1"}, []string{"label2"}},
		{[]string{"label1*"}, []string{}, []string{"label1"}, []string{"label2"}},
		{[]string{}, []string{"*"}, []string{}, []string{"label1", "label2"}},
		{[]string{}, []string{"lab*"}, []string{}, []string{"label1", "label2"}},
		{[]string{}, []string{"label1"}, []string{"label2"}, []string{"label1"}},
		{[]string{"*"}, []string{"*"}, []string{}, []string{"label1", "label2"}},
	}

	for _, tt := range gatherLabelsTests {
		t.Run("", func(t *testing.T) {
			var acc testutil.Accumulator
			d := Docker{
				newClient: newClient,
			}

			for _, label := range tt.include {
				d.LabelInclude = append(d.LabelInclude, label)
			}
			for _, label := range tt.exclude {
				d.LabelExclude = append(d.LabelExclude, label)
			}

			err := d.Gather(&acc)
			require.NoError(t, err)

			for _, label := range tt.expected {
				if !acc.HasTag("docker_container_cpu", label) {
					t.Errorf("Didn't get expected label of %s.  Test was:  Include: %s  Exclude %s",
						label, tt.include, tt.exclude)
				}
			}

			for _, label := range tt.notexpected {
				if acc.HasTag("docker_container_cpu", label) {
					t.Errorf("Got unexpected label of %s.  Test was:  Include: %s  Exclude %s",
						label, tt.include, tt.exclude)
				}
			}
		})
	}
}

func TestContainerNames(t *testing.T) {
	var gatherContainerNames = []struct {
		include     []string
		exclude     []string
		expected    []string
		notexpected []string
	}{
		{[]string{}, []string{}, []string{"etcd", "etcd2"}, []string{}},
		{[]string{"*"}, []string{}, []string{"etcd", "etcd2"}, []string{}},
		{[]string{"etc*"}, []string{}, []string{"etcd", "etcd2"}, []string{}},
		{[]string{"etcd"}, []string{}, []string{"etcd"}, []string{"etcd2"}},
		{[]string{"etcd2*"}, []string{}, []string{"etcd2"}, []string{"etcd"}},
		{[]string{}, []string{"etc*"}, []string{}, []string{"etcd", "etcd2"}},
		{[]string{}, []string{"etcd"}, []string{"etcd2"}, []string{"etcd"}},
		{[]string{"*"}, []string{"*"}, []string{"etcd", "etcd2"}, []string{}},
		{[]string{}, []string{"*"}, []string{""}, []string{"etcd", "etcd2"}},
	}

	for _, tt := range gatherContainerNames {
		t.Run("", func(t *testing.T) {
			var acc testutil.Accumulator

			d := Docker{
				newClient:        newClient,
				ContainerInclude: tt.include,
				ContainerExclude: tt.exclude,
			}

			err := d.Gather(&acc)
			require.NoError(t, err)

			for _, metric := range acc.Metrics {
				if metric.Measurement == "docker_container_cpu" {
					if val, ok := metric.Tags["container_name"]; ok {
						var found bool = false
						for _, cname := range tt.expected {
							if val == cname {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("Got unexpected container of %s. Test was -> Include: %s, Exclude: %s", val, tt.include, tt.exclude)
						}
					}
				}
			}

			for _, metric := range acc.Metrics {
				if metric.Measurement == "docker_container_cpu" {
					if val, ok := metric.Tags["container_name"]; ok {
						var found bool = false
						for _, cname := range tt.notexpected {
							if val == cname {
								found = true
								break
							}
						}
						if found {
							t.Errorf("Got unexpected container of %s. Test was -> Include: %s, Exclude: %s", val, tt.include, tt.exclude)
						}
					}
				}
			}
		})
	}
}

func TestDockerGatherInfo(t *testing.T) {
	var acc testutil.Accumulator
	d := Docker{
		newClient: newClient,
		TagEnvironment: []string{"ENVVAR1", "ENVVAR2", "ENVVAR3", "ENVVAR5",
			"ENVVAR6", "ENVVAR7", "ENVVAR8", "ENVVAR9"},
	}

	err := acc.GatherError(d.Gather)
	require.NoError(t, err)

	acc.AssertContainsTaggedFields(t,
		"docker",
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
		map[string]string{"engine_host": "absol"},
	)

	acc.AssertContainsTaggedFields(t,
		"docker_data",
		map[string]interface{}{
			"used":      int64(17300000000),
			"total":     int64(107400000000),
			"available": int64(36530000000),
		},
		map[string]string{
			"unit":        "bytes",
			"engine_host": "absol",
		},
	)
	acc.AssertContainsTaggedFields(t,
		"docker_container_cpu",
		map[string]interface{}{
			"usage_total":  uint64(1231652),
			"container_id": "b7dfbb9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296e2173",
		},
		map[string]string{
			"container_name":    "etcd2",
			"container_image":   "quay.io:4443/coreos/etcd",
			"cpu":               "cpu3",
			"container_version": "v2.2.2",
			"engine_host":       "absol",
			"ENVVAR1":           "loremipsum",
			"ENVVAR2":           "dolorsitamet",
			"ENVVAR3":           "=ubuntu:10.04",
			"ENVVAR7":           "ENVVAR8=ENVVAR9",
			"label1":            "test_value_1",
			"label2":            "test_value_2",
		},
	)
	acc.AssertContainsTaggedFields(t,
		"docker_container_mem",
		map[string]interface{}{
			"container_id":  "b7dfbb9478a6ae55e237d4d74f8bbb753f0817192b5081334dc78476296e2173",
			"limit":         uint64(18935443456),
			"max_usage":     uint64(0),
			"usage":         uint64(0),
			"usage_percent": float64(0),
		},
		map[string]string{
			"engine_host":       "absol",
			"container_name":    "etcd2",
			"container_image":   "quay.io:4443/coreos/etcd",
			"container_version": "v2.2.2",
			"ENVVAR1":           "loremipsum",
			"ENVVAR2":           "dolorsitamet",
			"ENVVAR3":           "=ubuntu:10.04",
			"ENVVAR7":           "ENVVAR8=ENVVAR9",
			"label1":            "test_value_1",
			"label2":            "test_value_2",
		},
	)
}
