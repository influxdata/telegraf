package event_hubs

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azure/eventhubs"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/json"
	"github.com/influxdata/telegraf/testutil"
)

func TestEmulatorIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Require the developers to explicitly accept the EULA of the emulator
	if os.Getenv("AZURE_EVENT_HUBS_EMULATOR_ACCEPT_EULA") != "yes" {
		t.Skip(`
			Skipping due to unexcepted EULA. To run this test, please check the EULA of the emulator
			at https://github.com/Azure/azure-event-hubs-emulator-installer/blob/main/EMULATOR_EULA.md
			and accept it by setting the environment variable AZURE_EVENT_HUBS_EMULATOR_ACCEPT_EULA
			to 'yes'.
		`)
	}

	// Load the configuration for the Event-Hubs instance
	emulatorConfig, err := os.ReadFile(filepath.Join("testdata", "Config.json"))
	require.NoError(t, err, "reading config failed")

	// Setup the Azure Event Hub emulator environment
	// See https://learn.microsoft.com/en-us/azure/event-hubs/test-locally-with-event-hub-emulator
	emulator, err := eventhubs.Run(
		t.Context(),
		"mcr.microsoft.com/azure-messaging/eventhubs-emulator:2.1.0",
		eventhubs.WithAcceptEULA(),
		eventhubs.WithConfig(bytes.NewReader(emulatorConfig)),
	)
	require.NoError(t, err, "failed to start container")
	defer emulator.Terminate(t.Context()) //nolint:errcheck // Can't do anything anyway

	conn, err := emulator.ConnectionString(t.Context())
	require.NoError(t, err, "getting connection string failed")
	conn += "EntityPath=test"

	// Setup plugin and connect
	serializer := &json.Serializer{}
	require.NoError(t, serializer.Init())

	plugin := &EventHubs{
		ConnectionString: conn,
		Timeout:          config.Duration(3 * time.Second),
		Log:              testutil.Logger{},
	}
	plugin.SetSerializer(serializer)
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// Make sure we are connected
	require.Eventually(t, func() bool {
		return plugin.Write(testutil.MockMetrics()) == nil
	}, 3*time.Second, 500*time.Millisecond)

	input := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{
				"source":   "foo",
				"division": "A",
				"type":     "temperature",
			},
			map[string]interface{}{
				"value": 23,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"test",
			map[string]string{
				"source":   "foo",
				"division": "A",
				"type":     "humidity",
			},
			map[string]interface{}{
				"value": 59,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"test",
			map[string]string{
				"source":   "bar",
				"division": "B",
				"type":     "temperature",
			},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"test",
			map[string]string{
				"source":   "bar",
				"division": "B",
				"type":     "humidity",
			},
			map[string]interface{}{
				"value": 87,
			},
			time.Unix(0, 0),
		),
	}

	require.NoError(t, plugin.Write(input))
}

func TestReconnectIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Require the developers to explicitly accept the EULA of the emulator
	if os.Getenv("AZURE_EVENT_HUBS_EMULATOR_ACCEPT_EULA") != "yes" {
		t.Skip(`
			Skipping due to unexcepted EULA. To run this test, please check the EULA of the emulator
			at https://github.com/Azure/azure-event-hubs-emulator-installer/blob/main/EMULATOR_EULA.md
			and accept it by setting the environment variable AZURE_EVENT_HUBS_EMULATOR_ACCEPT_EULA
			to 'yes'.
		`)
	}

	// Load the configuration for the Event-Hubs instance
	emulatorConfig, err := os.ReadFile(filepath.Join("testdata", "Config.json"))
	require.NoError(t, err, "reading config failed")

	// Setup the Azure Event Hub emulator environment
	// See https://learn.microsoft.com/en-us/azure/event-hubs/test-locally-with-event-hub-emulator
	emulator, err := eventhubs.Run(
		t.Context(),
		"mcr.microsoft.com/azure-messaging/eventhubs-emulator:2.1.0",
		eventhubs.WithAcceptEULA(),
		eventhubs.WithConfig(bytes.NewReader(emulatorConfig)),
	)
	require.NoError(t, err, "failed to start container")
	defer emulator.Terminate(t.Context()) //nolint:errcheck // Can't do anything anyway

	conn, err := emulator.ConnectionString(t.Context())
	require.NoError(t, err, "getting connection string failed")
	conn += "EntityPath=test"

	// Setup plugin and connect
	serializer := &json.Serializer{}
	require.NoError(t, serializer.Init())

	plugin := &EventHubs{
		ConnectionString: conn,
		Timeout:          config.Duration(3 * time.Second),
		Log:              testutil.Logger{},
	}
	plugin.SetSerializer(serializer)
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// Make sure we are connected
	require.Eventually(t, func() bool {
		return plugin.Write(testutil.MockMetrics()) == nil
	}, 3*time.Second, 500*time.Millisecond)

	input := []telegraf.Metric{
		metric.New(
			"test",
			map[string]string{
				"source":   "foo",
				"division": "A",
				"type":     "temperature",
			},
			map[string]interface{}{
				"value": 23,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"test",
			map[string]string{
				"source":   "foo",
				"division": "A",
				"type":     "humidity",
			},
			map[string]interface{}{
				"value": 59,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"test",
			map[string]string{
				"source":   "bar",
				"division": "B",
				"type":     "temperature",
			},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(0, 0),
		),
		metric.New(
			"test",
			map[string]string{
				"source":   "bar",
				"division": "B",
				"type":     "humidity",
			},
			map[string]interface{}{
				"value": 87,
			},
			time.Unix(0, 0),
		),
	}

	// This write should succeed as we should be able to connect to the
	// container
	require.NoError(t, plugin.Write(input))

	// Instantiate a docker client to be able to pause/resume the container
	client, err := testcontainers.NewDockerClientWithOpts(t.Context())
	require.NoError(t, err, "creating docker client failed")

	// Pause the container to simulate connection loss. Subsequent writes
	// should fail until the container is resumed
	require.NoError(t, client.ContainerPause(t.Context(), emulator.GetContainerID()))
	require.ErrorIs(t, plugin.Write(input), context.DeadlineExceeded)

	// Resume the container to check if the plugin reconnects
	require.NoError(t, client.ContainerUnpause(t.Context(), emulator.GetContainerID()))
	require.NoError(t, plugin.Write(input))
}
