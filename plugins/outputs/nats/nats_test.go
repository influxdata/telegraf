package nats

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
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
		setupCmds               []string
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
					Name: "my-external-stream",
				},
				ExternalStreamConfig: true,
				serializer:           &influx.Serializer{},
				Log:                  testutil.Logger{},
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
			setupCmds: []string{
				"nats",
				"stream",
				"add",
				"my-external-stream",
				"--subjects",
				"telegraf,telegraf2",
				"--defaults",
			},
			nats: &NATS{
				Name:    "telegraf",
				Subject: "telegraf",
				Jetstream: &StreamConfig{
					Name: "my-external-stream",
				},
				ExternalStreamConfig: true,
				serializer:           &influx.Serializer{},
				Log:                  testutil.Logger{},
			},
			streamConfigCompareFunc: func(t *testing.T, si *jetstream.StreamInfo) {
				require.Equal(t, "my-external-stream", si.Config.Name)
				require.Equal(t, []string{"telegraf", "telegraf2"}, si.Config.Subjects)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.container.Start()
			require.NoError(t, err, "failed to start container")
			defer tc.container.Terminate()

			// If nats cli setup commands are required they need to run
			// in a nats cli container. The server does not contain the
			// nats cli tool.
			if len(tc.setupCmds) > 0 {
				hostPort := tc.container.Ports[natsServicePort]
				srvPort, err := strconv.Atoi(hostPort)
				require.NoError(t, err, "failed to convert port to int")

				// The HostAccessPorts is not setup until after the container is started
				// so we override the entrypoint to sleep until the container is started
				// and then run the setup commands
				natsCli := testutil.Container{
					Image:           "bitnami/natscli",
					Cmd:             []string{"10"},
					Entrypoint:      []string{"sleep"},
					HostAccessPorts: []int{srvPort},
					Env: map[string]string{
						"NATS_URL": fmt.Sprintf("nats://%s:%s", testcontainers.HostInternal, tc.container.Ports[natsServicePort]),
					},
				}
				err = natsCli.Start()
				require.NoError(t, err, "failed to start cli container")
				natsCli.PrintLogs()

				exitCode, output, err := natsCli.Exec(tc.setupCmds)
				if exitCode != 0 {
					log.Printf("failed to setup container: %v", output)
					_, err = io.Copy(os.Stdout, output)
					if err != nil {
						require.NoError(t, err, "failed to copy output")
					}
				}

				require.NoError(t, err, "failed to setup container")
			}

			server := []string{fmt.Sprintf("nats://%s:%s", tc.container.Address, tc.container.Ports[natsServicePort])}
			tc.nats.Servers = server
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
