package docker_log

import (
	"fmt"
	"maps"
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

func TestStatePersistence(t *testing.T) {
	// Create a log stream mimicing the actual behavior of the server
	logs := make(chan *mock.Logs, 10)

	// Create server
	id := "1234567890"
	server := &mock.Server{
		Inspect: map[string]container.InspectResponse{id: {Config: &container.Config{Tty: true}}},
		List: []container.Summary{{
			ID:    id,
			Names: []string{"/" + id},
			Image: "influxdata/telegraf:1.11.0",
			State: "running",
		}},
		LogStreams: map[string]chan *mock.Logs{id: logs},
	}
	addr := server.Start(t)
	defer server.Close()

	// Setup plugin
	plugin := &DockerLogs{
		Endpoint: addr,
		Timeout:  config.Duration(time.Second * 5),
	}
	require.NoError(t, plugin.Init())
	plugin.lastRecordMtx.Lock()
	state := maps.Clone(plugin.lastRecord)
	plugin.lastRecordMtx.Unlock()
	require.Empty(t, state)

	// Write a first log message
	ts, err := time.Parse(time.RFC3339Nano, "2020-04-28T18:43:16.432691200Z")
	require.NoError(t, err)
	logs <- &mock.Logs{Content: ts.Format(time.RFC3339Nano) + " hello\n"}

	// Start the plugin and gather
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(&acc))
	require.Eventually(t, func() bool {
		return acc.NMetrics() > 0
	}, 5*time.Second, 50*time.Millisecond)
	require.Empty(t, acc.Errors)

	// Gracefully stop the plugin and make sure the state was updated
	plugin.Stop()

	plugin.lastRecordMtx.Lock()
	state = maps.Clone(plugin.lastRecord)
	plugin.lastRecordMtx.Unlock()

	require.Contains(t, state, id)
	require.Equal(t, ts.UTC(), state[id])
}

func TestStatePersistenceMux(t *testing.T) {
	// Create a log stream mimicing the actual behavior of the server
	logs := make(chan *mock.Logs, 10)

	// Create server
	id := "1234567890"
	server := &mock.Server{
		Inspect: map[string]container.InspectResponse{id: {Config: &container.Config{}}},
		List: []container.Summary{{
			ID:    id,
			Names: []string{"/" + id},
			Image: "influxdata/telegraf:1.11.0",
			State: "running",
		}},
		LogStreams: map[string]chan *mock.Logs{id: logs},
	}
	addr := server.Start(t)
	defer server.Close()

	// Setup plugin
	plugin := &DockerLogs{
		Endpoint: addr,
		Timeout:  config.Duration(time.Second * 5),
	}
	require.NoError(t, plugin.Init())
	plugin.lastRecordMtx.Lock()
	state := maps.Clone(plugin.lastRecord)
	plugin.lastRecordMtx.Unlock()
	require.Empty(t, state)

	// Write a first log message
	ts, err := time.Parse(time.RFC3339Nano, "2020-04-28T18:43:16.432691200Z")
	require.NoError(t, err)
	logs <- &mock.Logs{
		Content:     ts.Format(time.RFC3339Nano) + " hello\n",
		Multiplexed: true,
	}

	// Start the plugin and gather
	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.NoError(t, plugin.Gather(&acc))
	require.Eventually(t, func() bool {
		return acc.NMetrics() > 0
	}, 5*time.Second, 50*time.Millisecond)
	require.Empty(t, acc.Errors)

	// Gracefully stop the plugin and make sure the state was updated
	plugin.Stop()

	plugin.lastRecordMtx.Lock()
	state = maps.Clone(plugin.lastRecord)
	plugin.lastRecordMtx.Unlock()

	require.Contains(t, state, id)
	require.Equal(t, ts.UTC(), state[id])
}
