package docker_log

import (
	"context"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	common_docker "github.com/influxdata/telegraf/plugins/common/docker"
	common_tls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/testutil"
)

func TestGather(t *testing.T) {
	tests := []struct {
		name     string
		server   *common_docker.Server
		expected []telegraf.Metric
	}{
		{
			name:   "no containers",
			server: &common_docker.Server{},
		},
		{
			name: "one container tty",
			server: &common_docker.Server{
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
				Logs: map[string]common_docker.Logs{
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
					time.Unix(0, 1588099396432691200),
				),
			},
		},
		{
			name: "one container multiplex",
			server: &common_docker.Server{
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
				Logs: map[string]common_docker.Logs{
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
					time.Unix(0, 1588099336432691200),
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
				ClientConfig:     common_tls.ClientConfig{InsecureSkipVerify: true}, // Required as the test server has only a self-signed cert
				Timeout:          config.Duration(time.Second * 5),
				newClient:        newClient,
				containerList:    make(map[string]context.CancelFunc),
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Trigger a gather cycle which will make the logs to be "tracked"
			// and wait until we did see enough data
			require.NoError(t, plugin.Gather(&acc))
			require.Eventuallyf(t, func() bool {
				return acc.NMetrics() >= uint64(len(tt.expected))
			}, 3*time.Second, 100*time.Millisecond, "got %d metrics, expected %d", acc.NMetrics(), len(tt.expected))

			// Check the results
			require.Empty(t, acc.Errors)
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics())
		})
	}
}
