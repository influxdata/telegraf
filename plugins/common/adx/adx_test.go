package adx

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	serializers_json "github.com/influxdata/telegraf/plugins/serializers/json"
	"github.com/influxdata/telegraf/testutil"
)

func TestCreateAzureDataExplorerTable(t *testing.T) {
	serializer := &serializers_json.Serializer{}
	require.NoError(t, serializer.Init())
	plugin := AzureDataExplorer{
		Endpoint:        "someendpoint",
		Database:        "databasename",
		logger:          testutil.Logger{},
		MetricsGrouping: TablePerMetric,
		TableName:       "test1",
		CreateTables:    false,
		kustoClient:     kusto.NewMockClient(),
		metricIngestors: map[string]ingest.Ingestor{
			"test1": &fakeIngestor{},
		},
		IngestionType: QueuedIngestion,
	}

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	err := plugin.createAzureDataExplorerTable(context.Background(), "test1")

	output := buf.String()

	if err == nil && !strings.Contains(output, "skipped table creation") {
		t.Logf("FAILED : TestCreateAzureDataExplorerTable:  Should have skipped table creation.")
		t.Fail()
	}
}

func TestInit(t *testing.T) {
	testCases := []struct {
		name              string
		endpoint          string
		database          string
		metricsGrouping   string
		tableName         string
		timeout           config.Duration
		ingestionType     string
		expectedInitError string
	}{
		{
			name:            "Valid configuration",
			endpoint:        "someendpoint",
			database:        "databasename",
			metricsGrouping: TablePerMetric,
			timeout:         config.Duration(20 * time.Second),
			ingestionType:   QueuedIngestion,
		},
		{
			name:              "Empty endpoint",
			database:          "databasename",
			metricsGrouping:   TablePerMetric,
			expectedInitError: "endpoint configuration cannot be empty",
		},
		{
			name:              "Empty database",
			endpoint:          "someendpoint",
			metricsGrouping:   TablePerMetric,
			expectedInitError: "database configuration cannot be empty",
		},
		{
			name:              "SingleTable without table name",
			endpoint:          "someendpoint",
			database:          "databasename",
			metricsGrouping:   SingleTable,
			expectedInitError: "table name cannot be empty for SingleTable metrics grouping type",
		},
		{
			name:              "Invalid metrics grouping type",
			endpoint:          "someendpoint",
			database:          "databasename",
			metricsGrouping:   "invalidtype",
			expectedInitError: "metrics grouping type is not valid",
		},
		{
			name:              "Unknown ingestion type",
			endpoint:          "someendpoint",
			database:          "databasename",
			metricsGrouping:   TablePerMetric,
			ingestionType:     "unknown",
			expectedInitError: "unknown ingestion type \"unknown\"",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			plugin := AzureDataExplorer{
				Endpoint:        tC.endpoint,
				Database:        tC.database,
				MetricsGrouping: tC.metricsGrouping,
				TableName:       tC.tableName,
				Timeout:         tC.timeout,
				IngestionType:   tC.ingestionType,
				logger:          testutil.Logger{},
			}

			errorInit := plugin.Init()

			if tC.expectedInitError != "" {
				require.EqualError(t, errorInit, tC.expectedInitError)
			} else {
				require.NoError(t, errorInit)
			}
		})
	}
}

func TestInitBlankEndpointData(t *testing.T) {
	plugin := AzureDataExplorer{
		logger:          testutil.Logger{},
		kustoClient:     kusto.NewMockClient(),
		metricIngestors: map[string]ingest.Ingestor{},
	}

	errorInit := plugin.Init()
	require.Error(t, errorInit)
	require.Equal(t, "endpoint configuration cannot be empty", errorInit.Error())
}

func TestQueryConstruction(t *testing.T) {
	const tableName = "mytable"
	const expectedCreate = `.create-merge table ['mytable'] (['fields']:dynamic, ['name']:string, ['tags']:dynamic, ['timestamp']:datetime);`
	const expectedMapping = `` +
		`.create-or-alter table ['mytable'] ingestion json mapping 'mytable_mapping' '[{"column":"fields", ` +
		`"Properties":{"Path":"$[\'fields\']"}},{"column":"name", "Properties":{"Path":"$[\'name\']"}},{"column":"tags", ` +
		`"Properties":{"Path":"$[\'tags\']"}},{"column":"timestamp", "Properties":{"Path":"$[\'timestamp\']"}}]'`
	require.Equal(t, expectedCreate, createTableCommand(tableName).String())
	require.Equal(t, expectedMapping, createTableMappingCommand(tableName).String())
}

type fakeIngestor struct {
	actualOutputMetric map[string]interface{}
}

func (f *fakeIngestor) FromReader(_ context.Context, reader io.Reader, _ ...ingest.FileOption) (*ingest.Result, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Scan()
	firstLine := scanner.Text()
	err := json.Unmarshal([]byte(firstLine), &f.actualOutputMetric)
	if err != nil {
		return nil, err
	}
	return &ingest.Result{}, nil
}

func (f *fakeIngestor) FromFile(_ context.Context, _ string, _ ...ingest.FileOption) (*ingest.Result, error) {
	return &ingest.Result{}, nil
}

func (f *fakeIngestor) Close() error {
	return nil
}

func TestGetMetricIngestor(t *testing.T) {
	plugin := AzureDataExplorer{
		logger:        testutil.Logger{},
		IngestionType: QueuedIngestion,
		kustoClient:   kusto.NewMockClient(),
	}

	plugin.metricIngestors = map[string]ingest.Ingestor{
		"test1": &fakeIngestor{},
	}

	ingestor, err := plugin.GetMetricIngestor(context.Background(), "test1")
	if err != nil {
		t.Errorf("Error getting ingestor: %v", err)
	}
	require.NotNil(t, ingestor)
}

func TestGetMetricIngestorNoIngester(t *testing.T) {
	plugin := AzureDataExplorer{
		logger:        testutil.Logger{},
		IngestionType: QueuedIngestion,
		kustoClient:   kusto.NewMockClient(),
	}

	plugin.metricIngestors = map[string]ingest.Ingestor{}

	ingestor, err := plugin.GetMetricIngestor(context.Background(), "test1")
	if err != nil {
		t.Errorf("Error getting ingestor: %v", err)
	}
	require.NotNil(t, ingestor)
}

func TestPushMetrics(t *testing.T) {
	plugin := AzureDataExplorer{
		logger:        testutil.Logger{},
		IngestionType: QueuedIngestion,
		kustoClient:   kusto.NewMockClient(),
	}

	plugin.metricIngestors = map[string]ingest.Ingestor{
		"test1": &fakeIngestor{},
	}

	metrics := []byte(`[{"fields": {"value": 1}, "name": "test1", "tags": {"tag1": "value1"}, "timestamp": "2021-01-01T00:00:00Z"}]`)
	err := plugin.PushMetrics(ingest.FileFormat(ingest.JSON), "test1", metrics)
	if err != nil {
		t.Errorf("Error pushing metrics: %v", err)
	}
}

func TestConnect(t *testing.T) {
	testCases := []struct {
		name          string
		endpoint      string
		expectedError string
		expectedPanic bool
	}{
		{
			name:          "Valid connection",
			endpoint:      "https://valid.endpoint",
			expectedError: "",
			expectedPanic: false,
		},
		{
			name:          "Invalid connection",
			endpoint:      "",
			expectedError: "error: Connection string cannot be empty",
			expectedPanic: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			plugin := AzureDataExplorer{
				Endpoint: tC.endpoint,
				logger:   testutil.Logger{},
			}

			if tC.expectedPanic {
				require.PanicsWithValue(t, tC.expectedError, func() {
					err := plugin.Connect()
					require.NoError(t, err)
				})
			} else {
				require.NotPanics(t, func() {
					err := plugin.Connect()
					require.NoError(t, err)
					require.NotNil(t, plugin.kustoClient)
					require.NotNil(t, plugin.metricIngestors)
				})
			}
		})
	}
}

func TestClose(t *testing.T) {
	plugin := AzureDataExplorer{
		logger:        testutil.Logger{},
		IngestionType: QueuedIngestion,
		kustoClient:   kusto.NewMockClient(),
	}

	err := plugin.Close()
	require.NoError(t, err)
}
