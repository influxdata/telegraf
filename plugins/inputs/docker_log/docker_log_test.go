package docker_log

import (
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/docker/mock"
	"github.com/influxdata/telegraf/testutil"
)

func TestGather(t *testing.T) {
	tests := []struct {
		name     string
		server   *mock.Server
		expected []telegraf.Metric
	}{
		{
			name:   "no containers",
			server: &mock.Server{},
		},
		{
			name: "one container tty",
			server: &mock.Server{
				List: []container.Summary{
					{
						ID:    "deadbeef",
						Names: []string{"/telegraf"},
						Image: "influxdata/telegraf:1.11.0",
						State: "running",
					},
				},
				Inspect: map[string]container.InspectResponse{
					"deadbeef": {
						ID: "deadbeef",
						Config: &container.Config{
							Tty: true,
						},
					},
				},
				Logs: map[string]mock.Logs{
					"deadbeef": {Content: "2020-04-28T18:43:16.432691200Z hello\n"},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"docker_log",
					map[string]string{
						"container_name":    "telegraf",
						"container_image":   "influxdata/telegraf",
						"container_version": "1.11.0",
						"stream":            "tty",
						"source":            "deadbeef",
					},
					map[string]interface{}{
						"container_id": "deadbeef",
						"message":      "hello",
					},
					time.Date(2020, 4, 28, 18, 43, 16, 432691200, time.UTC),
				),
			},
		},
		{
			name: "one container multiplex",
			server: &mock.Server{
				List: []container.Summary{
					{
						ID:    "deadbeef",
						Names: []string{"/telegraf"},
						Image: "influxdata/telegraf:1.11.0",
						State: "running",
					},
				},
				Inspect: map[string]container.InspectResponse{
					"deadbeef": {
						Config: &container.Config{
							Tty: false,
						},
					},
				},
				Logs: map[string]mock.Logs{
					"deadbeef": {
						Content:     "2020-04-28T18:42:16.432691200Z hello from stdout\n",
						Multiplexed: true,
					},
				},
			},
			expected: []telegraf.Metric{
				metric.New(
					"docker_log",
					map[string]string{
						"container_name":    "telegraf",
						"container_image":   "influxdata/telegraf",
						"container_version": "1.11.0",
						"stream":            "stdout",
						"source":            "deadbeef",
					},
					map[string]interface{}{
						"container_id": "deadbeef",
						"message":      "hello from stdout",
					},
					time.Date(2020, 4, 28, 18, 42, 16, 432691200, time.UTC),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup a server mocking the docker daemon responses
			addr := tt.server.Start(t)
			defer tt.server.Close()

			// Setup the plugin
			plugin := &DockerLogs{
				Endpoint:         addr,
				IncludeSourceTag: true,
				Timeout:          config.Duration(time.Second * 5),
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Trigger a gather cycle which will make the logs to be "tracked"
			// and wait until we did see enough data
			require.NoError(t, plugin.Gather(&acc))
			require.Eventually(t, func() bool {
				return acc.NMetrics() >= uint64(len(tt.expected))
			}, 3*time.Second, 100*time.Millisecond)

			// Check the results
			require.Empty(t, acc.Errors)
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics())
		})
	}
}
