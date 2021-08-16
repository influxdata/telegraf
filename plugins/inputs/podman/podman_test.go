package podman

import (
	"context"
	"testing"
	"time"

	"github.com/containers/podman/v3/libpod/define"
	"github.com/containers/podman/v3/pkg/domain/entities"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

type MockClient struct {
	InfoF           func() (*define.Info, error)
	ContainerListF  func(filters map[string][]string) ([]entities.ListContainer, error)
	ContainerStatsF func(string) (*define.ContainerStats, error)
	BackgroundF     func() context.Context
}

func (c *MockClient) Info() (*define.Info, error) {
	return c.InfoF()
}

func (c *MockClient) ContainerList(
	ctx context.Context,
	filters map[string][]string,
) ([]entities.ListContainer, error) {
	return c.ContainerListF(filters)
}

func (c *MockClient) ContainerStats(
	ctx context.Context,
	containerID string,
) (*define.ContainerStats, error) {
	return c.ContainerStatsF(containerID)
}

func (c *MockClient) Background() context.Context {
	return c.BackgroundF()
}

var baseClient = MockClient{
	InfoF: func() (*define.Info, error) {
		return &info, nil
	},
	ContainerListF: func(filters map[string][]string) ([]entities.ListContainer, error) {
		return containerList, nil
	},
	ContainerStatsF: func(containerID string) (*define.ContainerStats, error) {
		if containerID == container_test_1 {
			return &containerStats_nginx, nil
		}
		return &containerStats_blissful_lewin, nil
	},
	BackgroundF: func() context.Context {
		return context.Background()
	},
}

func TestPodmanGather(t *testing.T) {
	var acc testutil.Accumulator

	p := &Podman{
		Log:    testutil.Logger{},
		client: &baseClient,
	}
	err := p.Gather(&acc)
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"podman",
			map[string]string{
				"engine_host":    "fedora",
				"server_version": "3.2.0",
			},
			map[string]interface{}{
				"n_containers":         int64(2),
				"n_containers_paused":  int64(0),
				"n_containers_running": int64(1),
				"n_containers_stopped": int64(1),
				"n_cpus":               int64(8),
				"n_images":             int64(10),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"podman",
			map[string]string{
				"engine_host":    "fedora",
				"server_version": "3.2.0",
			},
			map[string]interface{}{
				"memory_total": int64(0),
			},
			time.Unix(0, 0),
		),
	}
	actual := FilterMetrics(acc.GetTelegrafMetrics(), func(m telegraf.Metric) bool {
		return m.Name() == "podman"
	})

	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
}

func TestGatherContainerStats(t *testing.T) {
	var tests = []struct {
		name      string
		container string
		expected  []telegraf.Metric
	}{
		{
			name:      "Test ngnix container stats",
			container: container_test_2,
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"podman_container_stats",
					map[string]string{
						"container_image":   "docker.io/library/nginx",
						"container_name":    "nginx",
						"container_version": "latest",
						"engine_host":       "fedora",
						"server_version":    "3.2.0",
						"pod_name":          "",
					},
					map[string]interface{}{
						"container_id": string("9a4f6929b45ee0171b781233ce4c68acd2b7ede4fdf8d1dbe17edc3b07446854"),
						"cpu":          float64(2.1863381926526583e-09),
						"mem_limit":    uint64(7966027776),
						"mem_usage":    uint64(4014080),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:      "Test blissful_lewin container stats",
			container: container_test_1,
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"podman_container_stats",
					map[string]string{
						"container_image":   "docker.io/library/ubuntu",
						"container_name":    "blissful_lewin",
						"container_version": "latest",
						"engine_host":       "fedora",
						"server_version":    "3.2.0",
						"pod_name":          "elastic_pare",
					},
					map[string]interface{}{
						"container_id": string("59897a61355010568bb67c3c4150163b7246648ceae6f64fac77da590dacdc3d"),
						"cpu":          float64(3.687353584388549e-08),
						"mem_limit":    uint64(7966027776),
						"mem_usage":    uint64(3330048),
					},
					time.Unix(0, 0),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				acc testutil.Accumulator
				p   = Podman{
					Log:              testutil.Logger{},
					client:           &baseClient,
					ContainerExclude: []string{tt.container},
				}
			)

			err := p.Gather(&acc)
			require.NoError(t, err)

			actual := FilterMetrics(acc.GetTelegrafMetrics(), func(m telegraf.Metric) bool {
				return m.Name() == "podman_container_stats"
			})
			testutil.RequireMetricsEqual(t, tt.expected, actual, testutil.IgnoreTime())
		})
	}
}

func FilterMetrics(metrics []telegraf.Metric, f func(telegraf.Metric) bool) []telegraf.Metric {
	results := []telegraf.Metric{}
	for _, m := range metrics {
		if f(m) {
			results = append(results, m)
		}
	}
	return results
}

/*
func TestPodman(t *testing.T) {
	// Get Podman socket location
	sock_dir := os.Getenv("XDG_RUNTIME_DIR")
	socket := "unix:" + sock_dir + "/podman/podman.sock"
	var acc testutil.Accumulator
	p := &Podman{
		Log:      testutil.Logger{},
		Endpoint: socket,
	}
	err := p.Gather(&acc)
	if err != nil {
		log.Fatal(err)
	}
	acc.Wait(1)
	log.Println(acc.GetTelegrafMetrics())
}
*/
