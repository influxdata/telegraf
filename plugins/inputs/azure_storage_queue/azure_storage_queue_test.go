package azure_storage_queue

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azurite"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
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
	emulator, err := azurite.Run(
		t.Context(),
		"mcr.microsoft.com/azure-storage/azurite:3.28.0",
		azurite.WithInMemoryPersistence(64.0),
	)
	require.NoError(t, err, "failed to start Azurite container")
	defer testcontainers.TerminateContainer(emulator) //nolint:errcheck // Ignore error as we can't do anything about it

	endpoint := emulator.MustServiceURL(t.Context(), azurite.QueueService) + "/" + azurite.AccountName

	// Create two queues and push some messages to get data
	credentials, err := azqueue.NewSharedKeyCredential(azurite.AccountName, azurite.AccountKey)
	require.NoError(t, err)

	client, err := azqueue.NewServiceClientWithSharedKeyCredential(endpoint, credentials, nil)
	require.NoError(t, err)

	// Remember the oldest messages
	oldest := make(map[string]time.Time, 2)

	// Add five messages to test queue one
	_, err = client.CreateQueue(t.Context(), "test-one", nil)
	require.NoError(t, err)

	qc := client.NewQueueClient("test-one")
	for i := range 5 {
		msg := fmt.Sprintf(`{"count": %d, "message": "foobar"}`, i)
		resp, err := qc.EnqueueMessage(t.Context(), msg, nil)
		require.NoError(t, err)
		if i == 0 {
			oldest["test-one"] = *resp.Date
			time.Sleep(time.Second)
		}
	}

	// Add three messages to test queue two
	_, err = client.CreateQueue(t.Context(), "test-two", nil)
	require.NoError(t, err)

	qc = client.NewQueueClient("test-two")
	for i := range 3 {
		msg := fmt.Sprintf(`{"count": %d, "message": "tiger"}`, i)
		resp, err := qc.EnqueueMessage(t.Context(), msg, nil)
		require.NoError(t, err)
		if i == 0 {
			oldest["test-two"] = *resp.Date
			time.Sleep(time.Second)
		}
	}

	// Setup plugin
	plugin := &AzureStorageQueue{
		EndpointURL:          endpoint,
		StorageAccountName:   azurite.AccountName,
		StorageAccountKey:    azurite.AccountKey,
		PeekOldestMessageAge: true,
		Log:                  &testutil.Logger{},
	}

	require.NoError(t, plugin.Init())

	// Make sure we are connected
	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	expected := []telegraf.Metric{
		metric.New(
			"azure_storage_queues",
			map[string]string{
				"account": azurite.AccountName,
				"queue":   "test-one",
			},
			map[string]interface{}{
				"oldest_message_age_ns": int64(0),
				"size":                  int64(5),
			},
			time.Unix(0, 0),
		),
		metric.New(
			"azure_storage_queues",
			map[string]string{
				"account": azurite.AccountName,
				"queue":   "test-two",
			},
			map[string]interface{}{
				"oldest_message_age_ns": int64(0),
				"size":                  int64(3),
			},
			time.Unix(0, 0),
		),
	}

	// Test the metrics
	options := []cmp.Option{
		testutil.IgnoreTime(),
		testutil.IgnoreFields("oldest_message_age_ns"),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, options...)

	// Test the oldest-message values
	for _, m := range actual {
		q, found := m.GetTag("queue")
		require.True(t, found)

		actualAge, found := m.GetField("oldest_message_age_ns")
		require.True(t, found)

		expectedAge := m.Time().Sub(oldest[q])
		require.Equal(t, expectedAge.Nanoseconds(), actualAge)
	}
}
