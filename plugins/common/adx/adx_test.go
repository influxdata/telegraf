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
	"github.com/influxdata/telegraf/config"
	serializers_json "github.com/influxdata/telegraf/plugins/serializers/json"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
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

	ingestor, err := plugin.getMetricIngestor(context.Background(), "test1")
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

	ingestor, err := plugin.getMetricIngestor(context.Background(), "test1")
	if err != nil {
		t.Errorf("Error getting ingestor: %v", err)
	}
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
		name               string
		inputMetric        []telegraf.Metric
		metricsGrouping    string
		tableName          string
		expected           map[string]interface{}
		expectedWriteError string
		createTables       bool
		ingestionType      string
	}{
		{
			name:            "Valid metric",
			inputMetric:     testutil.MockMetrics(),
			createTables:    true,
			tableName:       "test1",
			metricsGrouping: TablePerMetric,
			expected: map[string]interface{}{
				"metricName": "test1",
				"fields": map[string]interface{}{
					"value": 1.0,
				},
				"tags": map[string]interface{}{
					"tag1": "value1",
				},
				"timestamp": float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).UnixNano() / int64(time.Second)),
			},
		},
		{
			name:            "Don't create tables'",
			inputMetric:     testutil.MockMetrics(),
			createTables:    false,
			tableName:       "test1",
			metricsGrouping: TablePerMetric,
			expected: map[string]interface{}{
				"metricName": "test1",
				"fields": map[string]interface{}{
					"value": 1.0,
				},
				"tags": map[string]interface{}{
					"tag1": "value1",
				},
				"timestamp": float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).UnixNano() / int64(time.Second)),
			},
		},
		{
			name:            "SingleTable metric grouping type",
			inputMetric:     testutil.MockMetrics(),
			createTables:    true,
			tableName:       "test1",
			metricsGrouping: SingleTable,
			expected: map[string]interface{}{
				"metricName": "test1",
				"fields": map[string]interface{}{
					"value": 1.0,
				},
				"tags": map[string]interface{}{
					"tag1": "value1",
				},
				"timestamp": float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).UnixNano() / int64(time.Second)),
			},
		},
		{
			name:            "Valid metric managed ingestion",
			inputMetric:     testutil.MockMetrics(),
			createTables:    true,
			tableName:       "test1",
			metricsGrouping: TablePerMetric,
			ingestionType:   ManagedIngestion,
			expected: map[string]interface{}{
				"metricName": "test1",
				"fields": map[string]interface{}{
					"value": 1.0,
				},
				"tags": map[string]interface{}{
					"tag1": "value1",
				},
				"timestamp": float64(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).UnixNano() / int64(time.Second)),
			},
		},
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

			localFakeIngestor := &fakeIngestor{}
			client := Client{
				cfg: &Config{
					Endpoint:        "https://someendpoint.kusto.net",
					Database:        "databasename",
					MetricsGrouping: tC.metricsGrouping,
					TableName:       tC.tableName,
					CreateTables:    tC.createTables,
					IngestionType:   ingestionType,
					Timeout:         config.Duration(20 * time.Second),
				},
				ingestors: map[string]ingest.Ingestor{tC.tableName: localFakeIngestor},
				logger:    testutil.Logger{},
			}

			tableMetricGroups := make(map[string][]byte)
			mockmetrics := testutil.MockMetrics()
			for _, m := range mockmetrics {
				metricInBytes, err := serializer.Serialize(m)
				require.NoError(t, err)
				tableMetricGroups[m.Name()] = append(tableMetricGroups[m.Name()], metricInBytes...)
			}

			format := ingest.FileFormat(ingest.JSON)
			for tableName, tableMetrics := range tableMetricGroups {
				errorInWrite := client.PushMetrics(format, tableName, tableMetrics)

				if tC.expectedWriteError != "" {
					require.EqualError(t, errorInWrite, tC.expectedWriteError)
				} else {
					require.NoError(t, errorInWrite)

					expectedNameOfMetric := tC.expected["metricName"].(string)

					createdFakeIngestor := localFakeIngestor

					require.Equal(t, expectedNameOfMetric, createdFakeIngestor.actualOutputMetric["name"])

					expectedFields := tC.expected["fields"].(map[string]interface{})
					require.Equal(t, expectedFields, createdFakeIngestor.actualOutputMetric["fields"])

					expectedTags := tC.expected["tags"].(map[string]interface{})
					require.Equal(t, expectedTags, createdFakeIngestor.actualOutputMetric["tags"])

					expectedTime := tC.expected["timestamp"].(float64)
					timestampStr := createdFakeIngestor.actualOutputMetric["timestamp"].(string)
					parsedTime, err := time.Parse(time.RFC3339Nano, timestampStr)
					parsedTimeFloat := float64(parsedTime.UnixNano()) / 1e9
					require.NoError(t, err)
					require.InDelta(t, expectedTime, parsedTimeFloat, testutil.DefaultDelta)
				}
			}
		})
	}
}

func TestClose(t *testing.T) {
	plugin := Client{
		logger: testutil.Logger{},
		cfg: &Config{
			IngestionType: QueuedIngestion,
		},
		client: kusto.NewMockClient(),
	}

	err := plugin.Close()
	require.NoError(t, err)
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
