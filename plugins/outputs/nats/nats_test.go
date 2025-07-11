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
			// Verify that we can successfully write data to the NATS daemon
			err = tc.nats.Write(testutil.MockMetrics())
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
func TestWriteWithLayoutIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	natsServicePort := "4222"
	type testConfig struct {
		name                    string
		container               testutil.Container
		nats                    *NATS
		streamConfigCompareFunc func(*testing.T, *jetstream.StreamInfo)
		tags                    map[string]string
		fields                  map[string]interface{}
		msgCount                int
		expectedSubjects        []string
	}
	testCases := []testConfig{
		{
			name: "subject layout with tags",
			container: testutil.Container{
				Image:        "nats:latest",
				ExposedPorts: []string{natsServicePort},
				Cmd:          []string{"--js"},
				WaitingFor:   wait.ForListeningPort(nat.Port(natsServicePort)),
			},
			nats: &NATS{
				Name:    "telegraf",
				Subject: "my-subject.metrics.{{ .Name }}.{{ .Tag \"tag1\" }}.{{ .Tag \"tag2\" }}",
				Jetstream: &StreamConfig{
					Name:     "my-telegraf-stream",
					Subjects: []string{"my-subject.>"},
				},
				serializer: &influx.Serializer{},
				Log:        testutil.Logger{},
			},
			streamConfigCompareFunc: func(t *testing.T, si *jetstream.StreamInfo) {
				require.Equal(t, "my-telegraf-stream", si.Config.Name)
				require.Equal(t, []string{"my-subject.>"}, si.Config.Subjects)
			},
			tags: map[string]string{
				"tag1": "foo",
				"tag2": "bar",
			},
			msgCount: 1,
			expectedSubjects: []string{
				"my-subject.metrics.test1.foo.bar",
			},
		},
		{
			name: "subject layout with field name",
			container: testutil.Container{
				Image:        "nats:latest",
				ExposedPorts: []string{natsServicePort},
				Cmd:          []string{"--js"},
				WaitingFor:   wait.ForListeningPort(nat.Port(natsServicePort)),
			},
			nats: &NATS{
				Name:    "telegraf",
				Subject: "my-subject.metrics.{{ .Tag \"tag1\" }}.{{ .Tag \"tag2\" }}.{{ .Name }}.{{ .Tag \"FieldName\" }}",
				Jetstream: &StreamConfig{
					Name:     "my-telegraf-stream",
					Subjects: []string{"my-subject.>"},
				},
				serializer: &influx.Serializer{},
				Log:        testutil.Logger{},
			},
			streamConfigCompareFunc: func(t *testing.T, si *jetstream.StreamInfo) {
				require.Equal(t, "my-telegraf-stream", si.Config.Name)
				require.Equal(t, []string{"my-subject.>"}, si.Config.Subjects)
			},
			tags: map[string]string{
				"tag1": "foo",
				"tag2": "bar",
			},
			fields: map[string]interface{}{
				"cpu_usage":  1,
				"cpu_idle":   2,
				"cpu_system": 3,
			},
			msgCount: 4, // Our 3 fields plus the default value field
			expectedSubjects: []string{
				"my-subject.metrics.foo.bar.test1.cpu_usage",
				"my-subject.metrics.foo.bar.test1.cpu_idle",
				"my-subject.metrics.foo.bar.test1.cpu_system",
				"my-subject.metrics.foo.bar.test1.value",
			},
		},
		{
			name: "subject layout missing end tag",
			container: testutil.Container{
				Image:        "nats:latest",
				ExposedPorts: []string{natsServicePort},
				Cmd:          []string{"--js"},
				WaitingFor:   wait.ForListeningPort(nat.Port(natsServicePort)),
			},
			nats: &NATS{
				Name:    "telegraf",
				Subject: "my-subject.{{ .Name }}.{{ .Tag \"tag1\" }}.{{ .Tag \"tag2\" }}",
				Jetstream: &StreamConfig{
					Name:     "my-telegraf-stream",
					Subjects: []string{"my-subject.>"},
				},
				serializer: &influx.Serializer{},
				Log:        testutil.Logger{},
			},
			streamConfigCompareFunc: func(t *testing.T, si *jetstream.StreamInfo) {
				require.Equal(t, "my-telegraf-stream", si.Config.Name)
				require.Equal(t, []string{"my-subject.>"}, si.Config.Subjects)
			},
			tags: map[string]string{
				"tag1": "foo",
			},
			msgCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, tc.container.Start(), "failed to start container")
			defer tc.container.Terminate()

			server := []string{fmt.Sprintf("nats://%s:%s", tc.container.Address, tc.container.Ports[natsServicePort])}
			tc.nats.Servers = server
			require.NoError(t, tc.nats.Init())
			require.NoError(t, tc.nats.Connect())

			stream, err := tc.nats.jetstreamClient.Stream(t.Context(), tc.nats.Jetstream.Name)
			require.NoError(t, err)
			si, err := stream.Info(t.Context())
			require.NoError(t, err)

			tc.streamConfigCompareFunc(t, si)
			// Verify that we can successfully write data to the NATS daemon
			metric := testutil.MockMetrics()
			for _, m := range metric {
				for k, v := range tc.tags {
					m.AddTag(k, v)
				}
				for k, v := range tc.fields {
					m.AddField(k, v)
				}
			}

			require.NoError(t, tc.nats.Write(metric))

			foundSubjects := make([]string, 0)
			if tc.nats.Jetstream != nil {
				js, err := tc.nats.conn.JetStream()
				require.NoError(t, err)
				sub, err := js.PullSubscribe(tc.nats.Jetstream.Subjects[0], "")
				require.NoError(t, err)

				msgs, _ := sub.Fetch(100, nats.MaxWait(1*time.Second))

				require.Len(t, msgs, tc.msgCount, "unexpected number of messages")
				for _, msg := range msgs {
					foundSubjects = append(foundSubjects, msg.Subject)
				}
			}

			require.NoError(t, validateSubjects(tc.expectedSubjects, foundSubjects))
		}) // end of test case
	}
}

// validateSubjects checks that:
// - All entries in generated exist in expected.
// - All expected values appear at least once in generated.
// Returns error if either condition fails.
func validateSubjects(expected, generated []string) error {
	expectedSet := make(map[string]struct{}, len(expected))
	for _, e := range expected {
		expectedSet[e] = struct{}{}
	}

	seen := make(map[string]bool)

	for _, g := range generated {
		if _, ok := expectedSet[g]; !ok {
			return fmt.Errorf("invalid entry found: %q is not in expected list", g)
		}
		seen[g] = true
	}

	for _, e := range expected {
		if !seen[e] {
			return fmt.Errorf("missing expected entry: %q not found in generated list", e)
		}
	}

	return nil
}
