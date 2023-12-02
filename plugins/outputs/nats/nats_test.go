package nats

import (
	"context"
	_ "embed"
	"fmt"
	"os"
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
			name: "valid with jetstream(stream created)",
			container: testutil.Container{
				Image:        "nats:latest",
				ExposedPorts: []string{natsServicePort},
				Cmd:          []string{"--js"},
				WaitingFor:   wait.ForListeningPort(nat.Port(natsServicePort)),
			},
			nats: &NATS{
				Name:    "telegraf",
				Subject: "telegraf",
				Stream:  "telegraf-stream",
				Jetstream: &JetstreamConfigWrapper{
					StreamConfig: jetstream.StreamConfig{
						Name: "this will be ignored",
					},
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
				Subject: "my-tel-sub",
				Stream:  "telegraf-stream-with-cfg",
				Jetstream: &JetstreamConfigWrapper{
					StreamConfig: jetstream.StreamConfig{
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

			if tc.nats.Stream != "" {
				stream, err := tc.nats.jetstreamClient.Stream(context.Background(), tc.nats.Stream)
				require.NoError(t, err)
				si, err := stream.Info(context.Background())
				require.NoError(t, err)
				require.Equal(t, tc.nats.Stream, tc.nats.Jetstream.Name)
				// compare only relevant fields, since defaults for fields like max_bytes is not 0
				fieldsEqualHelper(t, tc.nats.Jetstream.StreamConfig, si.Config, tc.streamConfigCompareFields...)
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

	if valA.Type() != valB.Type() {
		t.Error("Both parameters must be of the same type")
		return
	}

	for _, fieldName := range fieldNames {
		fieldA := valA.FieldByName(fieldName)
		fieldB := valB.FieldByName(fieldName)

		require.Equal(t, fieldA.Interface(), fieldB.Interface(), "Field %s should be equal", fieldName)
	}
}

func Test_extractNestedTable(t *testing.T) {
	tests := []struct {
		name    string
		tomlMap map[string]interface{}
		keys    []string
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "Valid Nested Table",
			tomlMap: map[string]interface{}{
				"outer": map[string]interface{}{
					"name": "outer",
					"inner": map[string]interface{}{
						"field1": "abc",
						"field2": "pqr",
					},
				},
			},
			keys: []string{"outer", "inner"},
			want: map[string]interface{}{"field1": "abc", "field2": "pqr"},
		},
		{
			name:    "Invalid Key",
			tomlMap: map[string]interface{}{"key": "value"},
			keys:    []string{"nonexistent"},
			wantErr: true,
		},
		{
			name:    "Non-Table Value",
			tomlMap: map[string]interface{}{"key": "value"},
			keys:    []string{"key"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := extractNestedTable(tt.tomlMap, tt.keys...)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, actual)
		})
	}
}

func TestConfigParsing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get all testcase directories
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Register the plugin
	outputs.Add("nats", func() telegraf.Output {
		return &NATS{}
	})

	for _, f := range folders {
		t.Run(f.Name(), func(t *testing.T) {
			testcasePath := filepath.Join("testcases", f.Name())
			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(testcasePath))
			require.Len(t, cfg.Outputs, 1)
		})
	}
}
