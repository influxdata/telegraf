package azure_data_explorer

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
	"github.com/influxdata/telegraf"
	telegrafJson "github.com/influxdata/telegraf/plugins/serializers/json"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestWriteWithType(t *testing.T) {
	metricName := "test1"
	fakeClient := kusto.NewMockClient()
	expectedResultMap := map[string]string{metricName: `{"fields":{"value":1},"name":"test1","tags":{"tag1":"value1"},"timestamp":1257894000}`}
	mockMetrics := testutil.MockMetrics()
	// Multi tables
	mockMetricsMulti := []telegraf.Metric{
		testutil.TestMetric(1.0, "test2"),
		testutil.TestMetric(2.0, "test3"),
	}
	expectedResultMap2 := map[string]string{"test2": `{"fields":{"value":1.0},"name":"test2","tags":{"tag1":"value1"},"timestamp":1257894000}`, "test3": `{"fields":{"value":2.0},"name":"test3","tags":{"tag1":"value1"},"timestamp":1257894000}`}
	// List of tests
	testCases := []struct {
		name                      string
		inputMetric               []telegraf.Metric
		metricsGrouping           string
		tableNameToExpectedResult map[string]string
		expectedWriteError        string
		createTables              bool
		ingestionType             string
	}{
		{
			name:                      "Valid metric",
			inputMetric:               mockMetrics,
			createTables:              true,
			metricsGrouping:           tablePerMetric,
			tableNameToExpectedResult: expectedResultMap,
		},
		{
			name:                      "Don't create tables'",
			inputMetric:               mockMetrics,
			createTables:              false,
			metricsGrouping:           tablePerMetric,
			tableNameToExpectedResult: expectedResultMap,
		},
		{
			name:                      "SingleTable metric grouping type",
			inputMetric:               mockMetrics,
			createTables:              true,
			metricsGrouping:           singleTable,
			tableNameToExpectedResult: expectedResultMap,
		},
		{
			name:                      "Valid metric managed ingestion",
			inputMetric:               mockMetrics,
			createTables:              true,
			metricsGrouping:           tablePerMetric,
			tableNameToExpectedResult: expectedResultMap,
			ingestionType:             managedIngestion,
		},
		{
			name:                      "Table per metric type",
			inputMetric:               mockMetricsMulti,
			createTables:              true,
			metricsGrouping:           tablePerMetric,
			tableNameToExpectedResult: expectedResultMap2,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			serializer, err := telegrafJson.NewSerializer(time.Second, "", "")
			require.NoError(t, err)
			for tableName, jsonValue := range testCase.tableNameToExpectedResult {
				ingestionType := "queued"
				if testCase.ingestionType != "" {
					ingestionType = testCase.ingestionType
				}
				mockIngestor := &mockIngestor{}
				plugin := AzureDataExplorer{
					Endpoint:        "someendpoint",
					Database:        "databasename",
					Log:             testutil.Logger{},
					IngestionType:   ingestionType,
					MetricsGrouping: testCase.metricsGrouping,
					TableName:       tableName,
					CreateTables:    testCase.createTables,
					kustoClient:     fakeClient,
					metricIngestors: map[string]ingest.Ingestor{
						tableName: mockIngestor,
					},
					serializer: serializer,
				}
				err := plugin.Write(testCase.inputMetric)
				if testCase.expectedWriteError != "" {
					require.EqualError(t, err, testCase.expectedWriteError)
					continue
				}
				require.NoError(t, err)
				createdIngestor := plugin.metricIngestors[tableName]
				if testCase.metricsGrouping == singleTable {
					createdIngestor = plugin.metricIngestors[tableName]
				}
				records := mockIngestor.records[0] // the first element
				require.NotNil(t, createdIngestor)
				require.JSONEq(t, jsonValue, records)
			}
		})
	}
}

func TestInitBlankEndpointData(t *testing.T) {
	plugin := AzureDataExplorer{
		Log:             testutil.Logger{},
		kustoClient:     kusto.NewMockClient(),
		metricIngestors: map[string]ingest.Ingestor{},
	}

	errorInit := plugin.Init()
	require.Error(t, errorInit)
	require.Equal(t, "Endpoint configuration cannot be empty", errorInit.Error())
}

type mockIngestor struct {
	records []string
}

func (m *mockIngestor) FromReader(ctx context.Context, reader io.Reader, options ...ingest.FileOption) (*ingest.Result, error) {
	bufbytes, _ := io.ReadAll(reader)
	metricjson := string(bufbytes)
	m.SetRecords(strings.Split(metricjson, "\n"))
	return &ingest.Result{}, nil
}

func (m *mockIngestor) FromFile(ctx context.Context, fPath string, options ...ingest.FileOption) (*ingest.Result, error) {
	return &ingest.Result{}, nil
}

func (m *mockIngestor) SetRecords(records []string) {
	m.records = records
}

// Name receives a copy of Foo since it doesn't need to modify it.
func (m *mockIngestor) Records() []string {
	return m.records
}

func (m *mockIngestor) Close() error {
	return nil
}
