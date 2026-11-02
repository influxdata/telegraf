package docker_log

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"testing/iotest"
	"time"

	"github.com/moby/moby/api/pkg/stdcopy"
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
					map[string]any{
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
					map[string]any{
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
	state := plugin.GetState().(map[string]time.Time)
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

	state = plugin.GetState().(map[string]time.Time)
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
	state := plugin.GetState().(map[string]time.Time)
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

	state = plugin.GetState().(map[string]time.Time)
	require.Contains(t, state, id)
	require.Equal(t, ts.UTC(), state[id])
}

func TestStatePersistenceEmptyLog(t *testing.T) {
	id := "1234567890"
	server := &mock.Server{
		Inspect: map[string]container.InspectResponse{id: {Config: &container.Config{Tty: true}}},
		List: []container.Summary{{
			ID:    id,
			Names: []string{"/" + id},
			Image: "influxdata/telegraf:1.11.0",
			State: "running",
		}},
		Logs: map[string]mock.Logs{id: {}},
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

	// Wait for the stream to end without any line
	require.NoError(t, plugin.Gather(&acc))
	require.Eventually(t, func() bool {
		plugin.mu.Lock()
		defer plugin.mu.Unlock()
		return len(plugin.containerList) == 0
	}, 5*time.Second, 50*time.Millisecond)
	require.Empty(t, acc.Errors)

	// A zero timestamp must not be recorded as it would replay the whole log
	require.Empty(t, plugin.GetState())
}

func TestGatherLegacyAPIVersion(t *testing.T) {
	server := &mock.Server{
		APIVersion: "1.39",
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
				Config: &container.Config{Tty: true},
			},
		},
		Logs: map[string]mock.Logs{
			"deadbeef": {Content: "2020-04-28T18:43:16.432691200Z hello\n"},
		},
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
	require.Empty(t, acc.Errors)
	require.Equal(t, "1.39", plugin.client.ClientVersion())
}

func TestGatherLegacyAPIVersionDaemonUnavailable(t *testing.T) {
	server := &mock.Server{APIVersion: "1.39"}
	addr := server.Start(t)
	defer server.Close()

	target, err := url.Parse(addr)
	require.NoError(t, err)
	backend := httputil.NewSingleHostReverseProxy(target)

	var unavailable atomic.Bool
	unavailable.Store(true)
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if unavailable.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		backend.ServeHTTP(w, r)
	}))
	defer proxy.Close()

	plugin := &DockerLogs{
		Endpoint: proxy.URL,
		Timeout:  config.Duration(time.Second * 5),
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.Error(t, plugin.Gather(&acc))

	unavailable.Store(false)
	require.NoError(t, plugin.Gather(&acc))
	require.Equal(t, "1.39", plugin.client.ClientVersion())
}

func TestGatherCurrentAPIVersionNotPinned(t *testing.T) {
	server := &mock.Server{APIVersion: "1.45"}
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

	original := plugin.client
	require.NoError(t, plugin.Gather(&acc))
	require.Same(t, original, plugin.client)
}

func TestTailStreamError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected error
	}{
		{
			name: "EOF",
			err:  io.EOF,
		},
		{
			name: "context canceled",
			err:  context.Canceled,
		},
		{
			name: "connection closed",
			err:  fmt.Errorf("read tcp: %w", net.ErrClosed),
		},
		{
			name:     "broken connection",
			err:      errors.New("broken"),
			expected: errors.New("broken"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := io.NopCloser(io.MultiReader(
				strings.NewReader("2020-04-28T18:43:16.432691200Z hello\n"),
				iotest.ErrReader(tt.err),
			))

			var acc testutil.Accumulator
			ts, err := tailStream(&acc, map[string]string{}, "deadbeef", reader, "tty")
			if tt.expected != nil {
				require.EqualError(t, err, tt.expected.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, time.Date(2020, 4, 28, 18, 43, 16, 432691200, time.UTC), ts)
			require.Equal(t, uint64(1), acc.NMetrics())
			require.Empty(t, acc.Errors)
		})
	}
}

func TestTailMultiplexedError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected error
	}{
		{
			name: "EOF",
			err:  io.EOF,
		},
		{
			name: "context canceled",
			err:  context.Canceled,
		},
		{
			name: "connection closed",
			err:  fmt.Errorf("read tcp: %w", net.ErrClosed),
		},
		{
			name:     "broken connection",
			err:      errors.New("broken"),
			expected: errors.New("broken"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writeMultiplexed(t, &buf, stdcopy.Stdout, "2020-04-28T18:43:16.432691200Z hello from stdout\n")
			writeMultiplexed(t, &buf, stdcopy.Stderr, "2020-04-28T18:43:17.432691200Z hello from stderr\n")
			reader := io.NopCloser(io.MultiReader(&buf, iotest.ErrReader(tt.err)))

			var acc testutil.Accumulator
			ts, err := tailMultiplexed(&acc, map[string]string{}, "deadbeef", reader)
			if tt.expected != nil {
				require.EqualError(t, err, tt.expected.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, time.Date(2020, 4, 28, 18, 43, 17, 432691200, time.UTC), ts)
			require.Equal(t, uint64(2), acc.NMetrics())
			require.Empty(t, acc.Errors)
		})
	}
}

func writeMultiplexed(t *testing.T, w io.Writer, stream stdcopy.StdType, content string) {
	t.Helper()

	header := [8]byte{0: byte(stream)}
	binary.BigEndian.PutUint32(header[4:], uint32(len(content)))
	_, err := w.Write(header[:])
	require.NoError(t, err)
	_, err = io.WriteString(w, content)
	require.NoError(t, err)
}
