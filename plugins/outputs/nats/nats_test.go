package nats

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestConnectAndWriteNATSIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	natsServicePort := "4222"
	type testConfig struct {
		name        string
		container   testutil.Container
		nats        *NATS
		matchConfig bool // flag to check if we need to check the whole config(used in case of json config)
		wantErr     bool
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
				Jetstream: &JetstreamConfig{
					Stream:           "telegraf-stream",
					AutoCreateStream: true,
				},
				serializer: &influx.Serializer{},
				Log:        testutil.Logger{},
			},
		},
		{
			name: "stream does not exist",
			container: testutil.Container{
				Image:        "nats:latest",
				ExposedPorts: []string{natsServicePort},
				Cmd:          []string{"--js"},
				WaitingFor:   wait.ForListeningPort(nat.Port(natsServicePort)),
			},
			nats: &NATS{
				Name:    "telegraf",
				Subject: "telegraf",
				Jetstream: &JetstreamConfig{
					Stream: "does-not-exist",
				},
				serializer: &influx.Serializer{},
				Log:        testutil.Logger{},
			},
			wantErr: true,
		},
		{
			name: "create stream via json config",
			container: testutil.Container{
				Image:        "nats:latest",
				ExposedPorts: []string{natsServicePort},
				Cmd:          []string{"--js"},
				WaitingFor:   wait.ForListeningPort(nat.Port(natsServicePort)),
			},
			nats: &NATS{
				Name:    "telegraf",
				Subject: "my-tel-sub",
				Jetstream: &JetstreamConfig{
					Stream: "telegraf-stream-via-cfg",
					// we are adding duplicate_window, max_msg_size, max_msgs_per_subject in the config,
					// since their defaults are different (2mins, -1, -1) respectively
					// max_age and duplicate_window are of types time.Duration and if provided using
					// strings such as 1h or 20m, the unmarshalling fails. This is also a limitation in nats-cli!
					StreamJSON: `
						{
							"retention": "workqueue",
							"max_consumers": 10,
							"discard": "old",
							"storage": "file",
							"max_msgs": 100000,
							"max_bytes": 104857600,
							"max_age": 86400000000000,
							"num_replicas": 1,
							"duplicate_window": 180000000000,
							"max_msg_size": 1200,
							"max_msgs_per_subject": 500
						}
					`,
					AutoCreateStream: true,
				},
				serializer: &influx.Serializer{},
				Log:        testutil.Logger{},
			},
			matchConfig: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.container.Start()
			require.NoError(t, err, "failed to start container")
			defer tc.container.Terminate()

			server := []string{fmt.Sprintf("nats://%s:%s", tc.container.Address, tc.container.Ports[natsServicePort])}
			tc.nats.Servers = server
			// Verify that we can connect to the NATS daemon
			err = tc.nats.Connect()
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tc.nats.Jetstream != nil {
				stream, err := tc.nats.jetstreamClient.Stream(context.Background(), tc.nats.Jetstream.Stream)
				require.NoError(t, err)
				si, err := stream.Info(context.Background())
				require.NoError(t, err)

				if tc.matchConfig {
					require.EqualValues(t, tc.nats.Jetstream.streamConfig, si.Config)
				}
			}
			// Verify that we can successfully write data to the NATS daemon
			err = tc.nats.Write(testutil.MockMetrics())
			require.NoError(t, err)
		})
	}
}
