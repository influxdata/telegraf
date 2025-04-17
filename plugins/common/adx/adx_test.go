package adx

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	serializers_json "github.com/influxdata/telegraf/plugins/serializers/json"
	"github.com/influxdata/telegraf/testutil"
)

func TestInitBlankEndpointData(t *testing.T) {
	plugin := Config{
		Endpoint: "",
		Database: "mydb",
	}

	_, err := plugin.NewClient("TestKusto.Telegraf", nil)
	require.Error(t, err)
	require.Equal(t, "endpoint configuration cannot be empty", err.Error())
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

func TestGetMetricIngestor(t *testing.T) {
	plugin := Client{
		logger: testutil.Logger{},
		client: kusto.NewMockClient(),
		cfg: &Config{
			Database:      "mydb",
			IngestionType: QueuedIngestion,
		},
		ingestors: map[string]ingest.Ingestor{"test1": &fakeIngestor{}},
	}

	ingestor, err := plugin.getMetricIngestor(t.Context(), "test1")
	require.NoError(t, err)
	require.NotNil(t, ingestor)
}

func TestGetMetricIngestorNoIngester(t *testing.T) {
	plugin := Client{
		logger: testutil.Logger{},
		client: kusto.NewMockClient(),
		cfg: &Config{
			IngestionType: QueuedIngestion,
		},
		ingestors: map[string]ingest.Ingestor{"test1": &fakeIngestor{}},
	}

	ingestor, err := plugin.getMetricIngestor(t.Context(), "test1")
	require.NoError(t, err)
	require.NotNil(t, ingestor)
}

func TestPushMetrics(t *testing.T) {
	plugin := Client{
		logger: testutil.Logger{},
		client: kusto.NewMockClient(),
		cfg: &Config{
			Database:      "mydb",
			Endpoint:      "https://ingest-test.westus.kusto.windows.net",
			IngestionType: QueuedIngestion,
		},
		ingestors: map[string]ingest.Ingestor{"test1": &fakeIngestor{}},
	}

	metrics := []byte(`{"fields": {"value": 1}, "name": "test1", "tags": {"tag1": "value1"}, "timestamp": "2021-01-01T00:00:00Z"}`)
	require.NoError(t, plugin.PushMetrics(ingest.FileFormat(ingest.JSON), "test1", metrics))
}

func TestPushMetricsOutputs(t *testing.T) {
	testCases := []struct {
		name            string
		inputMetric     []telegraf.Metric
		metricsGrouping string
		createTables    bool
		ingestionType   string
	}{
		{
			name:            "Valid metric",
			inputMetric:     testutil.MockMetrics(),
			createTables:    true,
			metricsGrouping: TablePerMetric,
		},
		{
			name:            "Don't create tables'",
			inputMetric:     testutil.MockMetrics(),
			createTables:    false,
			metricsGrouping: TablePerMetric,
		},
		{
			name:            "SingleTable metric grouping type",
			inputMetric:     testutil.MockMetrics(),
			createTables:    true,
			metricsGrouping: SingleTable,
		},
		{
			name:            "Valid metric managed ingestion",
			inputMetric:     testutil.MockMetrics(),
			createTables:    true,
			metricsGrouping: TablePerMetric,
			ingestionType:   ManagedIngestion,
		},
	}
	var expectedMetric = map[string]interface{}{
		"metricName": "test1",
		"fields": map[string]interface{}{
			"value": 1.0,
		},
		"tags": map[string]interface{}{
			"tag1": "value1",
		},
		"timestamp": float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).UnixNano() / int64(time.Second)),
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			ingestionType := "queued"
			if tC.ingestionType != "" {
				ingestionType = tC.ingestionType
			}

			serializer := &serializers_json.Serializer{
				TimestampUnits:  config.Duration(time.Nanosecond),
				TimestampFormat: time.RFC3339Nano,
			}

			cfg := &Config{
				Endpoint:        "https://someendpoint.kusto.net",
				Database:        "databasename",
				MetricsGrouping: tC.metricsGrouping,
				TableName:       "test1",
				CreateTables:    tC.createTables,
				IngestionType:   ingestionType,
				Timeout:         config.Duration(20 * time.Second),
			}
			client, err := cfg.NewClient("telegraf", &testutil.Logger{})
			require.NoError(t, err)

			// Inject the ingestor
			ingestor := &fakeIngestor{}
			client.ingestors["test1"] = ingestor

			tableMetricGroups := make(map[string][]byte)
			mockmetrics := testutil.MockMetrics()
			for _, m := range mockmetrics {
				metricInBytes, err := serializer.Serialize(m)
				require.NoError(t, err)
				tableMetricGroups[m.Name()] = append(tableMetricGroups[m.Name()], metricInBytes...)
			}

			format := ingest.FileFormat(ingest.JSON)
			for tableName, tableMetrics := range tableMetricGroups {
				require.NoError(t, client.PushMetrics(format, tableName, tableMetrics))
				createdFakeIngestor := ingestor
				require.EqualValues(t, expectedMetric["metricName"], createdFakeIngestor.actualOutputMetric["name"])
				require.EqualValues(t, expectedMetric["fields"], createdFakeIngestor.actualOutputMetric["fields"])
				require.EqualValues(t, expectedMetric["tags"], createdFakeIngestor.actualOutputMetric["tags"])
				timestampStr := createdFakeIngestor.actualOutputMetric["timestamp"].(string)
				parsedTime, err := time.Parse(time.RFC3339Nano, timestampStr)
				parsedTimeFloat := float64(parsedTime.UnixNano()) / 1e9
				require.NoError(t, err)
				require.InDelta(t, expectedMetric["timestamp"].(float64), parsedTimeFloat, testutil.DefaultDelta)
			}
		})
	}
}

func TestAlreadyClosed(t *testing.T) {
	plugin := Client{
		logger: testutil.Logger{},
		cfg: &Config{
			IngestionType: QueuedIngestion,
		},
		client: kusto.NewMockClient(),
	}
	require.NoError(t, plugin.Close())
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

func (*fakeIngestor) FromFile(_ context.Context, _ string, _ ...ingest.FileOption) (*ingest.Result, error) {
	return &ingest.Result{}, nil
}

func (*fakeIngestor) Close() error {
	return nil
}
