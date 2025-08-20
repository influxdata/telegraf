package nats

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestConnectAndWriteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	natsServicePort := "4222"
	type testConfig struct {
		name                    string
		container               testutil.Container
		externalStream          nats.StreamConfig
		nats                    *NATS
		streamConfigCompareFunc func(*testing.T, *jetstream.StreamInfo)
		wantErr                 bool
	}
	testCases := []testConfig{
		{
			name: "valid without jetstream",
			container: testutil.Container{
				Image:        "nats:latest",
				ExposedPorts: []string{natsServicePort},
				WaitingFor:   wait.ForListeningPort(nat.Port(natsServicePort)),
			},
			nats: &NATS{
				Name:       "telegraf",
				Subject:    "telegraf",
				serializer: &influx.Serializer{},
				Log:        testutil.Logger{},
			},
		},
		{
			name: "valid without jetstream and with batch",
			container: testutil.Container{
				Image:        "nats:latest",
				ExposedPorts: []string{natsServicePort},
				WaitingFor:   wait.ForListeningPort(nat.Port(natsServicePort)),
			},
			nats: &NATS{
				Name:           "telegraf",
				Subject:        "telegraf",
				serializer:     &influx.Serializer{},
				Log:            testutil.Logger{},
				UseBatchFormat: true,
			},
		},
		{
			name: "valid with jetstream",
			container: testutil.Container{
				Image:        "nats:latest",
				ExposedPorts: []string{natsServicePort},
				Cmd:          []string{"--js"},
				WaitingFor:   wait.ForListeningPort(nat.Port(natsServicePort)),
			},
			nats: &NATS{
				Name:    "telegraf",
				Subject: "telegraf",
				Jetstream: &StreamConfig{
					Name: "my-telegraf-stream",
				},
				serializer: &influx.Serializer{},
				Log:        testutil.Logger{},
			},
			streamConfigCompareFunc: func(t *testing.T, si *jetstream.StreamInfo) {
				require.Equal(t, "my-telegraf-stream", si.Config.Name)
				require.Equal(t, []string{"telegraf"}, si.Config.Subjects)
			},
		},
		{
			name: "valid with jetstream and batch",
			container: testutil.Container{
				Image:        "nats:latest",
				ExposedPorts: []string{natsServicePort},
				Cmd:          []string{"--js"},
				WaitingFor:   wait.ForListeningPort(nat.Port(natsServicePort)),
			},
			nats: &NATS{
				Name:    "telegraf",
				Subject: "telegraf",
				Jetstream: &StreamConfig{
					Name: "my-telegraf-stream",
				},
				serializer:     &influx.Serializer{},
				Log:            testutil.Logger{},
				UseBatchFormat: true,
			},
			streamConfigCompareFunc: func(t *testing.T, si *jetstream.StreamInfo) {
				require.Equal(t, "my-telegraf-stream", si.Config.Name)
				require.Equal(t, []string{"telegraf"}, si.Config.Subjects)
			},
		},
		{
			name: "create stream with config",
			container: testutil.Container{
				Image:        "nats:latest",
				ExposedPorts: []string{natsServicePort},
				Cmd:          []string{"--js"},
				WaitingFor:   wait.ForListeningPort(nat.Port(natsServicePort)),
			},
			nats: &NATS{
				Name:    "telegraf",
				Subject: "my-tel-sub-outer",
				Jetstream: &StreamConfig{
					Name:              "telegraf-stream-with-cfg",
					Subjects:          []string{"my-tel-sub0", "my-tel-sub1", "my-tel-sub2"},
					Retention:         "workqueue",
					MaxConsumers:      10,
					Discard:           "new",
					Storage:           "memory",
					MaxMsgs:           100_000,
					MaxBytes:          104_857_600,
					MaxAge:            config.Duration(10 * time.Minute),
					Duplicates:        config.Duration(5 * time.Minute),
					MaxMsgSize:        120,
					MaxMsgsPerSubject: 500,
				},
				serializer: &influx.Serializer{},
				Log:        testutil.Logger{},
			},
			streamConfigCompareFunc: func(t *testing.T, si *jetstream.StreamInfo) {
				require.Equal(t, "telegraf-stream-with-cfg", si.Config.Name)
				require.Equal(t, []string{"my-tel-sub0", "my-tel-sub1", "my-tel-sub2", "my-tel-sub-outer"}, si.Config.Subjects)
				require.Equal(t, jetstream.WorkQueuePolicy, si.Config.Retention)
				require.Equal(t, 10, si.Config.MaxConsumers)
				require.Equal(t, jetstream.DiscardNew, si.Config.Discard)
				require.Equal(t, jetstream.MemoryStorage, si.Config.Storage)
				require.Equal(t, int64(100_000), si.Config.MaxMsgs)
				require.Equal(t, int64(104_857_600), si.Config.MaxBytes)
				require.Equal(t, 10*time.Minute, si.Config.MaxAge)
				require.Equal(t, 5*time.Minute, si.Config.Duplicates)
				require.Equal(t, int32(120), si.Config.MaxMsgSize)
				require.Equal(t, int64(500), si.Config.MaxMsgsPerSubject)
			},
		},
		{
			name: "stream missing with external jetstream",
			container: testutil.Container{
				Image:        "nats:latest",
				ExposedPorts: []string{natsServicePort},
				Cmd:          []string{"--js"},
				WaitingFor:   wait.ForListeningPort(nat.Port(natsServicePort)),
			},
			nats: &NATS{
				Name:    "telegraf",
				Subject: "telegraf",
				Jetstream: &StreamConfig{
					Name:                  "my-external-stream",
					DisableStreamCreation: true,
				},
				serializer: &influx.Serializer{},
				Log:        testutil.Logger{},
			},
			wantErr: true,
		},
		{
			name: "stream exists external jetstream",
			container: testutil.Container{
				Image:        "nats:latest",
				ExposedPorts: []string{natsServicePort},
				Cmd:          []string{"--js"},
				WaitingFor:   wait.ForListeningPort(nat.Port(natsServicePort)),
			},
			externalStream: nats.StreamConfig{
				Name:         "my-external-stream",
				Subjects:     []string{"telegraf", "telegraf2"},
				MaxConsumers: 6,
				MaxMsgs:      10101,
			},
			nats: &NATS{
				Name:    "telegraf",
				Subject: "telegraf",
				Jetstream: &StreamConfig{
					Name:                  "my-external-stream",
					DisableStreamCreation: true,
					MaxMsgs:               10,
					MaxConsumers:          100,
				},
				serializer: &influx.Serializer{},
				Log:        testutil.Logger{},
			},
			streamConfigCompareFunc: func(t *testing.T, si *jetstream.StreamInfo) {
				require.Equal(t, "my-external-stream", si.Config.Name)
				require.Equal(t, []string{"telegraf", "telegraf2"}, si.Config.Subjects)
				require.Equal(t, int(6), si.Config.MaxConsumers)
				require.Equal(t, int64(10101), si.Config.MaxMsgs)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.container.Start()
			require.NoError(t, err, "failed to start container")
			defer tc.container.Terminate()

			server := "nats://" + tc.container.Address + ":" + tc.container.Ports[natsServicePort]

			// Create the stream before starting the plugin to simulate
			// externally managed streams
			if len(tc.externalStream.Name) > 0 {
				createStream(t, server, &tc.externalStream)
			}

			tc.nats.Servers = []string{server}
			// Verify that we can connect to the NATS daemon
			require.NoError(t, tc.nats.Init())
			err = tc.nats.Connect()
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tc.nats.Jetstream != nil {
				stream, err := tc.nats.jetstreamClient.Stream(t.Context(), tc.nats.Jetstream.Name)
				require.NoError(t, err)
				si, err := stream.Info(t.Context())
				require.NoError(t, err)

				tc.streamConfigCompareFunc(t, si)
			}
			// Verify that we can successfully write a single metric to the NATS daemon
			err = tc.nats.Write(testutil.MockMetrics())
			require.NoError(t, err)

			// Verify that we can successfully write multiple metrics to the NATS daemon
			twoMetrics := []telegraf.Metric{testutil.TestMetric(1.0), testutil.TestMetric(2.0)}
			err = tc.nats.Write(twoMetrics)
			require.NoError(t, err)
		})
	}
}

func TestConfigParsing(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{name: "Valid Default", path: filepath.Join("testcases", "no-js.conf")},
		{name: "Valid Default with Batch", path: filepath.Join("testcases", "no-js-batch.conf")},
		{name: "Valid JS", path: filepath.Join("testcases", "js-default.conf")},
		{name: "Valid JS Config", path: filepath.Join("testcases", "js-config.conf")},
		{name: "Valid JS Async Publish", path: filepath.Join("testcases", "js-async-pub.conf")},
		{name: "Subjects warning", path: filepath.Join("testcases", "js-subjects.conf")},
		{name: "Invalid JS", path: filepath.Join("testcases", "js-no-stream.conf"), wantErr: true},
		{name: "JS with layout", path: filepath.Join("testcases", "js-layout.conf")},
		{name: "Invalid JS with layout", path: filepath.Join("testcases", "js-layout-nosub.conf"), wantErr: true},
	}

	// Register the plugin
	outputs.Add("nats", func() telegraf.Output {
		return &NATS{}
	})
	srl := &influx.Serializer{}
	require.NoError(t, srl.Init())

	// Run tests using the table-driven approach
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(tc.path))
			require.Len(t, cfg.Outputs, 1)
			err := cfg.Outputs[0].Init()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWriteWithLayoutIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	natsServicePort := "4222"

	container := testutil.Container{
		Image:        "nats:latest",
		ExposedPorts: []string{natsServicePort},
		Cmd:          []string{"--js"},
		WaitingFor:   wait.ForListeningPort(nat.Port(natsServicePort)),
	}
	require.NoError(t, container.Start(), "failed to start container")
	defer container.Terminate()

	tests := []struct {
		name             string
		subject          string
		sendMetrics      []telegraf.Metric
		expectedSubjects []string
	}{
		{
			name:    "subject layout with tags",
			subject: "my-subject.metrics.{{ .Name }}.{{ .Tag \"tag1\" }}.{{ .Tag \"tag2\" }}",
			sendMetrics: []telegraf.Metric{metric.New(
				"test1",
				map[string]string{"tag1": "foo", "tag2": "bar"},
				map[string]interface{}{"value": 1.0},
				time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			)},
			expectedSubjects: []string{
				"my-subject.metrics.test1.foo.bar",
			},
		},
		{
			name:    "subject layout with field name",
			subject: "my-subject.metrics.{{ .Tag \"tag1\" }}.{{ .Tag \"tag2\" }}.{{ .Name }}.{{ .Field \"value\" }}",
			sendMetrics: []telegraf.Metric{metric.New(
				"test1",
				map[string]string{"tag1": "foo", "tag2": "bar"},
				map[string]interface{}{"value": 1.0},
				time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			)},
			expectedSubjects: []string{
				"my-subject.metrics.foo.bar.test1.1",
			},
		},
	}

	server := []string{fmt.Sprintf("nats://%s:%s", container.Address, container.Ports[natsServicePort])}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plugin := &NATS{
				Name: "telegraf",
				Jetstream: &StreamConfig{
					Name:     "my-telegraf-stream",
					Subjects: []string{"my-subject.>"},
				},
				serializer: &influx.Serializer{},
				Log:        testutil.Logger{},
				Subject:    tc.subject,
				Servers:    server,
			}

			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Connect())
			defer plugin.Close()
			// Get the stream to check for subjects and messages
			stream, err := plugin.jetstreamClient.Stream(t.Context(), plugin.Jetstream.Name)
			require.NoError(t, err)

			// Validate the stream properties
			info, err := stream.Info(t.Context())
			require.NoError(t, err)
			require.Equal(t, "my-telegraf-stream", info.Config.Name)
			require.Len(t, info.Config.Subjects, 1)
			require.Equal(t, "my-subject.>", info.Config.Subjects[0])

			// Write the metrics
			require.NoError(t, plugin.Write(tc.sendMetrics))
			metricCount := len(tc.sendMetrics)

			// Access the JetStream and validate the number of messages
			// as well as the created subjects
			js, err := plugin.conn.JetStream()
			require.NoError(t, err)

			// Make sure to erase all streams created by the plugin for the
			// next run to avoid side effects
			defer js.PurgeStream(plugin.Jetstream.Name) //nolint:errcheck // Ignore the returned error as we cannot do anything about it anyway

			require.Len(t, plugin.Jetstream.Subjects, 1)
			sub, err := js.PullSubscribe(plugin.Jetstream.Subjects[0], "")
			require.NoError(t, err)

			msgs, err := sub.Fetch(metricCount, nats.MaxWait(1*time.Second))
			require.NoError(t, err)

			require.Len(t, msgs, metricCount, "unexpected number of messages")

			actual := make([]string, 0, metricCount)
			for _, msg := range msgs {
				actual = append(actual, msg.Subject)
			}
			require.Equal(t, tc.expectedSubjects, actual)
		})
	}
}

func createStream(t *testing.T, server string, cfg *nats.StreamConfig) {
	t.Helper()

	// Connect to NATS server
	conn, err := nats.Connect(server)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, conn.Drain(), "draining failed")
	}()

	// Create the stream in the JetStream context
	js, err := conn.JetStream()
	require.NoError(t, err)
	_, err = js.AddStream(cfg)
	require.NoError(t, err)
}
