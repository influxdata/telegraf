package nats

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestConnectAndWriteNATSIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	natsServicePort := "4222"
	type testConfig struct {
		name                      string
		container                 testutil.Container
		nats                      *NATS
		streamConfigCompareFields []string
		wantErr                   bool
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
			streamConfigCompareFields: []string{"Name", "Subjects"},
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
				Subject: "my-tel-sub2",
				Jetstream: &StreamConfig{
					Name:              "telegraf-stream-with-cfg",
					Subjects:          []string{"my-tel-sub0", "my-tel-sub1", "my-tel-sub2"},
					Retention:         jetstream.WorkQueuePolicy,
					MaxConsumers:      10,
					Discard:           jetstream.DiscardOld,
					Storage:           jetstream.FileStorage,
					MaxMsgs:           100000,
					MaxBytes:          104857600,
					MaxAge:            86400000000000,
					Replicas:          1,
					Duplicates:        180000000000,
					MaxMsgSize:        120,
					MaxMsgsPerSubject: 500,
				},
				serializer: &influx.Serializer{},
				Log:        testutil.Logger{},
			},
			streamConfigCompareFields: []string{
				"Name",
				"Subjects",
				"Retention",
				"MaxConsumers",
				"Discard",
				"Storage",
				"MaxMsgs",
				"MaxBytes",
				"MaxAge",
				"Replicas",
				"Duplicates",
				"MaxMsgSize",
				"MaxMsgsPerSubject"},
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
				stream, err := tc.nats.jetstreamClient.Stream(context.Background(), tc.nats.Jetstream.Name)
				require.NoError(t, err)
				si, err := stream.Info(context.Background())
				require.NoError(t, err)
				// compare only relevant fields, since defaults for fields like max_bytes is not 0
				fieldsEqualHelper(t, *tc.nats.Jetstream, si.Config, tc.streamConfigCompareFields...)
			}
			// Verify that we can successfully write data to the NATS daemon
			err = tc.nats.Write(testutil.MockMetrics())
			require.NoError(t, err)
		})
	}
}

func fieldsEqualHelper(t *testing.T, a, b interface{}, fieldNames ...string) {
	valA := reflect.ValueOf(a)
	valB := reflect.ValueOf(b)

	if valA.Kind() != reflect.Struct || valB.Kind() != reflect.Struct {
		t.Error("Both parameters must be structs")
		return
	}

	for _, fieldName := range fieldNames {
		fieldA := valA.FieldByName(fieldName)
		fieldB := valB.FieldByName(fieldName)

		require.Equal(t, fieldA.Interface(), fieldB.Interface(), "Field %s should be equal", fieldName)
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
