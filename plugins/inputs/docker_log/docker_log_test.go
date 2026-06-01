package docker_log

import (
	"fmt"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

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

func TestGatherConcurrentState(t *testing.T) {
	// Spawn many containers so their tailing goroutines update the shared
	// last-record state concurrently. Run with -race to detect unsynchronized
	// access to the state map.
	const count = 64
	server := &mock.Server{
		Inspect: make(map[string]container.InspectResponse, count),
		Logs:    make(map[string]mock.Logs, count),
	}
	for i := range count {
		id := fmt.Sprintf("container%03d", i)
		server.List = append(server.List, container.Summary{
			ID:    id,
			Names: []string{"/" + id},
			Image: "influxdata/telegraf:1.11.0",
			State: "running",
		})
		server.Inspect[id] = container.InspectResponse{Config: &container.Config{Tty: true}}
		server.Logs[id] = mock.Logs{Content: "2020-04-28T18:43:16.432691200Z hello\n"}
	}
	addr := server.Start(t)
	defer server.Close()

	plugin := &DockerLogs{
		Endpoint: addr,
		Timeout:  config.Duration(time.Second * 5),
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(&acc))
	require.Eventually(t, func() bool {
		return acc.NMetrics() >= uint64(count)
	}, 5*time.Second, 50*time.Millisecond)
	require.Empty(t, acc.Errors)
}

func TestTailLogsNoDuplicateIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name          string
		fromBeginning bool
	}{
		{name: "telegraf-docker-log-from-end", fromBeginning: false},
		{name: "telegraf-docker-log-from-beginning", fromBeginning: true},
	}

	// Continuously emit a uniquely numbered line so every record has a distinct
	// timestamp and we can detect any line that gets collected more than once.
	const script = "i=1; while true; do echo \"log line $i\"; i=$((i+1)); sleep 0.3; done"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cntnr := testutil.Container{
				Name:       tt.name,
				Image:      "alpine:3",
				Entrypoint: []string{"/bin/sh", "-c", script},
				WaitingFor: wait.ForLog("log line 1"),
			}
			require.NoError(t, cntnr.Start(), "failed to start container")
			defer cntnr.Terminate()

			// Restrict the plugin to our container so it ignores everything else
			// running on the daemon (e.g. the testcontainers reaper).
			plugin := &DockerLogs{
				Endpoint:         "ENV",
				FromBeginning:    tt.fromBeginning,
				ContainerInclude: []string{tt.name},
				Timeout:          config.Duration(5 * time.Second),
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// First cycle reads the lines produced so far and persists the offset.
			require.NoError(t, plugin.Gather(&acc))
			require.Eventually(t, func() bool {
				return acc.NMetrics() > 0
			}, 10*time.Second, 200*time.Millisecond)
			collected := acc.NMetrics()

			// Keep gathering until lines produced after the first cycle are
			// collected, proving the offset advances without over-skipping new
			// records. The container guards a second goroutine while the first
			// is still running, so repeated Gather calls cannot double-read.
			require.Eventually(t, func() bool {
				require.NoError(t, plugin.Gather(&acc))
				return acc.NMetrics() > collected
			}, 20*time.Second, 300*time.Millisecond)

			// Docker's "since" filter is inclusive of the boundary timestamp, so
			// the last line of one cycle must not be re-emitted by the next one.
			counts := make(map[string]int)
			for _, m := range acc.GetTelegrafMetrics() {
				msg, ok := m.GetField("message")
				require.True(t, ok)
				counts[msg.(string)]++
			}
			for msg, n := range counts {
				require.Equalf(t, 1, n, "log line %q was collected %d times", msg, n)
			}
		})
	}
}
