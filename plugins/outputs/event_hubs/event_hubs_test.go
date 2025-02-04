package event_hubs

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azurite"
	"github.com/testcontainers/testcontainers-go/wait"

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

	// Setup the Azure Event Hub emulator environment
	// See https://learn.microsoft.com/en-us/azure/event-hubs/test-locally-with-event-hub-emulator
	ctx := context.Background()
	azuriteContainer, err := azurite.Run(ctx, "mcr.microsoft.com/azure-storage/azurite:3.28.0")
	require.NoError(t, err, "failed to start Azurite container")
	defer func() {
		if err := testcontainers.TerminateContainer(azuriteContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	blobPort, err := azuriteContainer.MappedPort(ctx, azurite.BlobPort)
	require.NoError(t, err)

	metadataPort, err := azuriteContainer.MappedPort(ctx, azurite.TablePort)
	require.NoError(t, err)

	cfgfile, err := filepath.Abs(filepath.Join("testdata", "Config.json"))
	require.NoError(t, err, "getting absolute path for config")
	emulator := testutil.Container{
		Image: "mcr.microsoft.com/azure-messaging/eventhubs-emulator:latest",
		Env: map[string]string{
			"BLOB_SERVER":     "host.docker.internal:" + blobPort.Port(),
			"METADATA_SERVER": "host.docker.internal:" + metadataPort.Port(),
			"ACCEPT_EULA":     "Y",
		},
		Files: map[string]string{
			"/Eventhubs_Emulator/ConfigFiles/Config.json": cfgfile,
		},
		HostAccessPorts: []int{blobPort.Int(), metadataPort.Int()},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.ExtraHosts = append(hc.ExtraHosts, "host.docker.internal:host-gateway")
		},
		ExposedPorts: []string{"5672"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort(nat.Port("5672")),
		),
	}
	require.NoError(t, emulator.Start(), "failed to start Azure Event Hub emulator container")
	defer emulator.Terminate()

	conn := "Endpoint=sb://" + emulator.Address + ":" + emulator.Ports["5672"] + ";"
	conn += "SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=SAS_KEY_VALUE;UseDevelopmentEmulator=true;EntityPath=test"

	// Setup plugin and connect
	serializer := &json.Serializer{}
	require.NoError(t, serializer.Init())

	plugin := &EventHubs{
		ConnectionString: conn,
		Timeout:          config.Duration(60 * time.Second),
	}
	plugin.SetSerializer(serializer)
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

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
