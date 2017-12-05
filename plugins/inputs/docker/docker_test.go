package docker

import (
	"context"
	"crypto/tls"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/stretchr/testify/require"
)

type MockClient struct {
	InfoF             func(ctx context.Context) (types.Info, error)
	ContainerListF    func(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerStatsF   func(ctx context.Context, containerID string, stream bool) (types.ContainerStats, error)
	ContainerInspectF func(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ServiceListF      func(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error)
	TaskListF         func(ctx context.Context, options types.TaskListOptions) ([]swarm.Task, error)
	NodeListF         func(ctx context.Context, options types.NodeListOptions) ([]swarm.Node, error)
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

func (c *MockClient) ServiceList(
	ctx context.Context,
	options types.ServiceListOptions,
) ([]swarm.Service, error) {
	return c.ServiceListF(ctx, options)
}

func (c *MockClient) TaskList(
	ctx context.Context,
	options types.TaskListOptions,
) ([]swarm.Task, error) {
	return c.TaskListF(ctx, options)
}

func (c *MockClient) NodeList(
	ctx context.Context,
	options types.NodeListOptions,
) ([]swarm.Node, error) {
	return c.NodeListF(ctx, options)
}

var baseClient = MockClient{
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
	ServiceListF: func(context.Context, types.ServiceListOptions) ([]swarm.Service, error) {
		return ServiceList, nil
	},
	TaskListF: func(context.Context, types.TaskListOptions) ([]swarm.Task, error) {
		return TaskList, nil
	},
	NodeListF: func(context.Context, types.NodeListOptions) ([]swarm.Node, error) {
		return NodeList, nil
	},
}

func newClient(host string, tlsConfig *tls.Config) (Client, error) {
	return &baseClient, nil
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
				ServiceListF: func(context.Context, types.ServiceListOptions) ([]swarm.Service, error) {
					return ServiceList, nil
				},
				TaskListF: func(context.Context, types.TaskListOptions) ([]swarm.Task, error) {
					return TaskList, nil
				},
				NodeListF: func(context.Context, types.NodeListOptions) ([]swarm.Node, error) {
					return NodeList, nil
				},
			}, nil
		},
	}
	err := d.Gather(&acc)
	require.NoError(t, err)
}

func TestContainerLabels(t *testing.T) {
	var tests = []struct {
		name      string
		container types.Container
		include   []string
		exclude   []string
		expected  map[string]string
	}{
		{
			name: "Nil filters matches all",
			container: types.Container{
				Labels: map[string]string{
					"a": "x",
				},
			},
			include: nil,
			exclude: nil,
			expected: map[string]string{
				"a": "x",
			},
		},
		{
			name: "Empty filters matches all",
			container: types.Container{
				Labels: map[string]string{
					"a": "x",
				},
			},
			include: []string{},
			exclude: []string{},
			expected: map[string]string{
				"a": "x",
			},
		},
		{
			name: "Must match include",
			container: types.Container{
				Labels: map[string]string{
					"a": "x",
					"b": "y",
				},
			},
			include: []string{"a"},
			exclude: []string{},
			expected: map[string]string{
				"a": "x",
			},
		},
		{
			name: "Must not match exclude",
			container: types.Container{
				Labels: map[string]string{
					"a": "x",
					"b": "y",
				},
			},
			include: []string{},
			exclude: []string{"b"},
			expected: map[string]string{
				"a": "x",
			},
		},
		{
			name: "Include Glob",
			container: types.Container{
				Labels: map[string]string{
					"aa": "x",
					"ab": "y",
					"bb": "z",
				},
			},
			include: []string{"a*"},
			exclude: []string{},
			expected: map[string]string{
				"aa": "x",
				"ab": "y",
			},
		},
		{
			name: "Exclude Glob",
			container: types.Container{
				Labels: map[string]string{
					"aa": "x",
					"ab": "y",
					"bb": "z",
				},
			},
			include: []string{},
			exclude: []string{"a*"},
			expected: map[string]string{
				"bb": "z",
			},
		},
		{
			name: "Excluded Includes",
			container: types.Container{
				Labels: map[string]string{
					"aa": "x",
					"ab": "y",
					"bb": "z",
				},
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
			var acc testutil.Accumulator

			newClientFunc := func(host string, tlsConfig *tls.Config) (Client, error) {
				client := baseClient
				client.ContainerListF = func(context.Context, types.ContainerListOptions) ([]types.Container, error) {
					return []types.Container{tt.container}, nil
				}
				return &client, nil
			}

			d := Docker{
				newClient:    newClientFunc,
				LabelInclude: tt.include,
				LabelExclude: tt.exclude,
			}

			err := d.Gather(&acc)
			require.NoError(t, err)

			// Grab tags from a container metric
			var actual map[string]string
			for _, metric := range acc.Metrics {
				if metric.Measurement == "docker_container_cpu" {
					actual = metric.Tags
				}
			}

			for k, v := range tt.expected {
				require.Equal(t, v, actual[k])
			}
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
			name: "Nil filters matches all",
			containers: [][]string{
				{"/etcd"},
				{"/etcd2"},
			},
			include:  nil,
			exclude:  nil,
			expected: []string{"etcd", "etcd2"},
		},
		{
			name: "Empty filters matches all",
			containers: [][]string{
				{"/etcd"},
				{"/etcd2"},
			},
			include:  []string{},
			exclude:  []string{},
			expected: []string{"etcd", "etcd2"},
		},
		{
			name: "Match all containers",
			containers: [][]string{
				{"/etcd"},
				{"/etcd2"},
			},
			include:  []string{"*"},
			exclude:  []string{},
			expected: []string{"etcd", "etcd2"},
		},
		{
			name: "Include prefix match",
			containers: [][]string{
				{"/etcd"},
				{"/etcd2"},
			},
			include:  []string{"etc*"},
			exclude:  []string{},
			expected: []string{"etcd", "etcd2"},
		},
		{
			name: "Exact match",
			containers: [][]string{
				{"/etcd"},
				{"/etcd2"},
			},
			include:  []string{"etcd"},
			exclude:  []string{},
			expected: []string{"etcd"},
		},
		{
			name: "Star matches zero length",
			containers: [][]string{
				{"/etcd"},
				{"/etcd2"},
			},
			include:  []string{"etcd2*"},
			exclude:  []string{},
			expected: []string{"etcd2"},
		},
		{
			name: "Exclude matches all",
			containers: [][]string{
				{"/etcd"},
				{"/etcd2"},
			},
			include:  []string{},
			exclude:  []string{"etc*"},
			expected: []string{},
		},
		{
			name: "Exclude single",
			containers: [][]string{
				{"/etcd"},
				{"/etcd2"},
			},
			include:  []string{},
			exclude:  []string{"etcd"},
			expected: []string{"etcd2"},
		},
		{
			name: "Exclude all",
			containers: [][]string{
				{"/etcd"},
				{"/etcd2"},
			},
			include:  []string{"*"},
			exclude:  []string{"*"},
			expected: []string{},
		},
		{
			name: "Exclude item matching include",
			containers: [][]string{
				{"acme"},
				{"foo"},
				{"acme-test"},
			},
			include:  []string{"acme*"},
			exclude:  []string{"*test*"},
			expected: []string{"acme"},
		},
		{
			name: "Exclude item no wildcards",
			containers: [][]string{
				{"acme"},
				{"acme-test"},
			},
			include:  []string{"acme*"},
			exclude:  []string{"test"},
			expected: []string{"acme", "acme-test"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var acc testutil.Accumulator

			newClientFunc := func(host string, tlsConfig *tls.Config) (Client, error) {
				client := baseClient
				client.ContainerListF = func(context.Context, types.ContainerListOptions) ([]types.Container, error) {
					var containers []types.Container
					for _, names := range tt.containers {
						containers = append(containers, types.Container{
							Names: names,
						})
					}
					return containers, nil
				}
				return &client, nil
			}

			d := Docker{
				newClient:        newClientFunc,
				ContainerInclude: tt.include,
				ContainerExclude: tt.exclude,
			}

			err := d.Gather(&acc)
			require.NoError(t, err)

			// Set of expected names
			var expected = make(map[string]bool)
			for _, v := range tt.expected {
				expected[v] = true
			}

			// Set of actual names
			var actual = make(map[string]bool)
			for _, metric := range acc.Metrics {
				if name, ok := metric.Tags["container_name"]; ok {
					actual[name] = true
				}
			}

			require.Equal(t, expected, actual)
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

func TestDockerGatherSwarmInfo(t *testing.T) {
	var acc testutil.Accumulator
	d := Docker{
		newClient: newClient,
	}

	err := acc.GatherError(d.Gather)
	require.NoError(t, err)

	d.gatherSwarmInfo(&acc)

	// test docker_container_net measurement
	acc.AssertContainsTaggedFields(t,
		"docker_swarm",
		map[string]interface{}{
			"tasks_running": int(2),
			"tasks_desired": uint64(2),
		},
		map[string]string{
			"service_id":   "qolkls9g5iasdiuihcyz9rnx2",
			"service_name": "test1",
			"service_mode": "replicated",
		},
	)

	acc.AssertContainsTaggedFields(t,
		"docker_swarm",
		map[string]interface{}{
			"tasks_running": int(1),
			"tasks_desired": int(1),
		},
		map[string]string{
			"service_id":   "qolkls9g5iasdiuihcyz9rn3",
			"service_name": "test2",
			"service_mode": "global",
		},
	)
}
