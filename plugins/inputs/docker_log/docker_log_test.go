package docker_log

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"os"
	"path"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

type MockClient struct {
	ContainerListF    func(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerInspectF func(ctx context.Context, containerID string) (types.ContainerJSON, error)
	ContainerLogsF    func(ctx context.Context, containerID string, options types.ContainerLogsOptions) (io.ReadCloser, error)
}

func (c *MockClient) ContainerList(
	ctx context.Context,
	options types.ContainerListOptions,
) ([]types.Container, error) {
	return c.ContainerListF(ctx, options)
}

func (c *MockClient) ContainerInspect(
	ctx context.Context,
	containerID string,
) (types.ContainerJSON, error) {
	return c.ContainerInspectF(ctx, containerID)
}

func (c *MockClient) ContainerLogs(
	ctx context.Context,
	containerID string,
	options types.ContainerLogsOptions,
) (io.ReadCloser, error) {
	return c.ContainerLogsF(ctx, containerID, options)
}

type Response struct {
	io.Reader
}

func (r *Response) Close() error {
	return nil
}

func MustParse(layout, value string) time.Time {
	tm, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return tm
}

func Test(t *testing.T) {
	tests := []struct {
		name     string
		client   *MockClient
		expected []telegraf.Metric
	}{
		{
			name: "no containers",
			client: &MockClient{
				ContainerListF: func(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
					return nil, nil
				},
			},
		},
		{
			name: "one container tty",
			client: &MockClient{
				ContainerListF: func(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
					return []types.Container{
						{
							ID:    "deadbeef",
							Names: []string{"/telegraf"},
							Image: "influxdata/telegraf:1.11.0",
						},
					}, nil
				},
				ContainerInspectF: func(ctx context.Context, containerID string) (types.ContainerJSON, error) {
					return types.ContainerJSON{
						Config: &container.Config{
							Tty: true,
						},
					}, nil
				},
				ContainerLogsF: func(ctx context.Context, containerID string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
					return &Response{Reader: bytes.NewBuffer([]byte("2020-04-28T18:43:16.432691200Z hello\n"))}, nil
				},
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
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
					MustParse(time.RFC3339Nano, "2020-04-28T18:43:16.432691200Z"),
				),
			},
		},
		{
			name: "one container multiplex",
			client: &MockClient{
				ContainerListF: func(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
					return []types.Container{
						{
							ID:    "deadbeef",
							Names: []string{"/telegraf"},
							Image: "influxdata/telegraf:1.11.0",
						},
					}, nil
				},
				ContainerInspectF: func(ctx context.Context, containerID string) (types.ContainerJSON, error) {
					return types.ContainerJSON{
						Config: &container.Config{
							Tty: false,
						},
					}, nil
				},
				ContainerLogsF: func(ctx context.Context, containerID string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
					var buf bytes.Buffer
					w := stdcopy.NewStdWriter(&buf, stdcopy.Stdout)
					w.Write([]byte("2020-04-28T18:42:16.432691200Z hello from stdout"))
					return &Response{Reader: &buf}, nil
				},
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
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
					MustParse(time.RFC3339Nano, "2020-04-28T18:42:16.432691200Z"),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var acc testutil.Accumulator
			plugin := &DockerLogs{
				Timeout:           internal.Duration{Duration: time.Second * 5},
				newClient:         func(string, *tls.Config) (Client, error) { return tt.client, nil },
				containerList:     make(map[string]context.CancelFunc),
				IncludeSourceTag:  true,
				OffsetStoragePath: ".",
				OffsetFlush:       internal.Duration{Duration: 1 * time.Second},
				offsetChan:        make(chan offsetData),
			}

			err := plugin.Init()
			require.NoError(t, err)

			//Remove offset files
			containers, _ := tt.client.ContainerListF(nil, types.ContainerListOptions{})
			for _, cont := range containers {
				os.RemoveAll(path.Join(plugin.OffsetStoragePath, cont.ID))
			}

			err = plugin.Gather(&acc)
			require.NoError(t, err)

			acc.Wait(len(tt.expected))
			plugin.Stop()

			require.Nil(t, acc.Errors) // no errors during gathering

			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics())

			//Remove offset files
			for _, cont := range containers {
				os.RemoveAll(path.Join(plugin.OffsetStoragePath, cont.ID))
			}
		})
	}
}

func TestResumeFromOffset(t *testing.T) {
	client := &MockClient{
		ContainerListF: func(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
			return []types.Container{
				{
					ID:    "badcafe",
					Names: []string{"/telegraf"},
					Image: "influxdata/telegraf:1.11.0",
				},
			}, nil
		},
		ContainerInspectF: func(ctx context.Context, containerID string) (types.ContainerJSON, error) {
			return types.ContainerJSON{
				Config: &container.Config{
					Tty: true,
				},
			}, nil
		},
		ContainerLogsF: func(ctx context.Context, containerID string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
			data := []byte("2020-04-28T18:43:16.000000000Z message 1\n2020-04-28T18:43:17.000000000Z message 2")
			returnedData := make([]byte, 0, len(data))
			s := bufio.NewScanner(bytes.NewReader(data))
			for s.Scan() {
				parts := bytes.SplitN(s.Bytes(), []byte(" "), 2)
				ts := MustParse(time.RFC3339Nano, string(parts[0]))
				if ts.After(MustParse(time.RFC3339Nano, options.Since)) {
					returnedData = append(returnedData, s.Bytes()...)
					returnedData = append(returnedData, byte('\n'))
				}
			}

			return &Response{Reader: bytes.NewBuffer(returnedData)}, nil
		},
	}
	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"docker_log",
			map[string]string{
				"container_name":    "telegraf",
				"container_image":   "influxdata/telegraf",
				"container_version": "1.11.0",
				"stream":            "tty",
				"source":            "badcafe",
			},
			map[string]interface{}{
				"container_id": "badcafe",
				"message":      "message 1",
			},
			MustParse(time.RFC3339Nano, "2020-04-28T18:43:16.000000000Z"),
		),
		testutil.MustMetric(
			"docker_log",
			map[string]string{
				"container_name":    "telegraf",
				"container_image":   "influxdata/telegraf",
				"container_version": "1.11.0",
				"stream":            "tty",
				"source":            "badcafe",
			},
			map[string]interface{}{
				"container_id": "badcafe",
				"message":      "message 2",
			},
			MustParse(time.RFC3339Nano, "2020-04-28T18:43:17.000000000Z"),
		),
	}
	acc := testutil.Accumulator{}

	plugin := &DockerLogs{
		Timeout:           internal.Duration{Duration: time.Second * 5},
		newClient:         func(string, *tls.Config) (Client, error) { return client, nil },
		containerList:     make(map[string]context.CancelFunc),
		IncludeSourceTag:  true,
		OffsetStoragePath: ".",
		OffsetFlush:       internal.Duration{Duration: 1 * time.Second},
		offsetChan:        make(chan offsetData),
	}

	cont, _ := client.ContainerListF(nil, types.ContainerListOptions{})

	err := plugin.Init()
	require.NoError(t, err)

	os.RemoveAll(path.Join(plugin.OffsetStoragePath, cont[0].ID))
	//Create offset file
	plugin.flushOffsetToFs(cont[0].ID, metrics[0].Time().UnixNano()+1)

	err = plugin.Gather(&acc)
	require.NoError(t, err)

	acc.Wait(len(metrics) - 1) //Should be 1 metric only
	plugin.Stop()

	require.Nil(t, acc.Errors) // no errors during gathering
	//First metric should be filtered based on time-stamp
	testutil.RequireMetricsEqual(t, []telegraf.Metric{metrics[1]}, acc.GetTelegrafMetrics())

	//Examine offset file
	_, tsUnix := plugin.loadOffsetFormFs(cont[0].ID)

	//in the offset file it should be last message timestamp +1 ns
	require.Equal(t, tsUnix, metrics[1].Time().UnixNano()+1)

	//Remove offset file
	os.RemoveAll(path.Join(plugin.OffsetStoragePath, cont[0].ID))
}

func TestDeliverAllMessageNoOffset(t *testing.T) {
	client := &MockClient{
		ContainerListF: func(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
			return []types.Container{
				{
					ID:    "badf00d",
					Names: []string{"/telegraf"},
					Image: "influxdata/telegraf:1.11.0",
				},
			}, nil
		},
		ContainerInspectF: func(ctx context.Context, containerID string) (types.ContainerJSON, error) {
			return types.ContainerJSON{
				Config: &container.Config{
					Tty: true,
				},
			}, nil
		},
		ContainerLogsF: func(ctx context.Context, containerID string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
			return &Response{Reader: bytes.NewBuffer([]byte("2020-04-28T18:43:16.000000000Z message 1\n2020-04-28T18:43:17.000000000Z message 2"))}, nil
		},
	}
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"docker_log",
			map[string]string{
				"container_name":    "telegraf",
				"container_image":   "influxdata/telegraf",
				"container_version": "1.11.0",
				"stream":            "tty",
				"source":            "badf00d",
			},
			map[string]interface{}{
				"container_id": "badf00d",
				"message":      "message 1",
			},
			MustParse(time.RFC3339Nano, "2020-04-28T18:43:16.000000000Z"),
		),
		testutil.MustMetric(
			"docker_log",
			map[string]string{
				"container_name":    "telegraf",
				"container_image":   "influxdata/telegraf",
				"container_version": "1.11.0",
				"stream":            "tty",
				"source":            "badf00d",
			},
			map[string]interface{}{
				"container_id": "badf00d",
				"message":      "message 2",
			},
			MustParse(time.RFC3339Nano, "2020-04-28T18:43:17.000000000Z"),
		),
	}
	acc := testutil.Accumulator{}

	plugin := &DockerLogs{
		Timeout:          internal.Duration{Duration: time.Second * 5},
		newClient:        func(string, *tls.Config) (Client, error) { return client, nil },
		containerList:    make(map[string]context.CancelFunc),
		IncludeSourceTag: true,
		FromBeginning:    true,
		OffsetFlush:      internal.Duration{Duration: 10 * time.Second},
		offsetChan:       make(chan offsetData),
	}

	err := plugin.Init()
	require.NoError(t, err)

	cont, _ := client.ContainerListF(nil, types.ContainerListOptions{})
	os.RemoveAll(path.Join(plugin.OffsetStoragePath, cont[0].ID))

	err = plugin.Gather(&acc)
	require.NoError(t, err)

	acc.Wait(len(expected))
	plugin.Stop()

	require.Nil(t, acc.Errors) // no errors during gathering

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics()) //2 metrics delivered

	//Examine offset file
	_, tsUnix := plugin.loadOffsetFormFs(expected[0].Tags()["source"])

	//in the offset file it should be last message timestamp +1 ns
	require.Equal(t, tsUnix, expected[len(expected)-1].Time().UnixNano()+1)

	//Remove offset files
	os.RemoveAll(path.Join(plugin.OffsetStoragePath, cont[0].ID))
}
