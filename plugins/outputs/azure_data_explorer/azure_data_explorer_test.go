package azure_data_explorer

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
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

func TestWrite(t *testing.T) {
	metricName := "test1"
	mockClient := kusto.NewMockClient()
	expectedResultMap := map[string]string{metricName: `{"fields":{"value":1},"name":"test1","tags":{"tag1":"value1"},"timestamp":1257894000}`}
	mockMetrics := testutil.MockMetrics()
	// Multi tables
	mockMetrics2 := testutil.TestMetric(1.0, "test2")
	mockMetrics3 := testutil.TestMetric(2.0, "test3")
	mockMetricsMulti := make([]telegraf.Metric, 2)
	mockMetricsMulti[0] = mockMetrics2
	mockMetricsMulti[1] = mockMetrics3
	expectedResultMap2 := map[string]string{"test2": `{"fields":{"value":1.0},"name":"test2","tags":{"tag1":"value1"},"timestamp":1257894000}`, "test3": `{"fields":{"value":2.0},"name":"test3","tags":{"tag1":"value1"},"timestamp":1257894000}`}

	testCases := []struct {
		name                      string
		inputMetric               []telegraf.Metric
		metricsGrouping           string
		tableNameToExpectedResult map[string]string
		expectedWriteError        string
		createTables              bool
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
			name:                      "Table per metric type",
			inputMetric:               mockMetricsMulti,
			createTables:              true,
			metricsGrouping:           tablePerMetric,
			tableNameToExpectedResult: expectedResultMap2,
		},
	}

	for _, tC := range testCases {
		tC := tC
		t.Run(tC.name, func(t *testing.T) {
			//t.Parallel()
			serializer, err := telegrafJson.NewSerializer(time.Second, "", "")
			require.NoError(t, err)
			for tableName, jsonValue := range tC.tableNameToExpectedResult {
				mockIngestor := &mockIngestor{}
				plugin := AzureDataExplorer{
					Endpoint:        "someendpoint",
					Database:        "databasename",
					Log:             testutil.Logger{},
					MetricsGrouping: tC.metricsGrouping,
					TableName:       tableName,
					CreateTables:    tC.createTables,
					client:          mockClient,
					ingestors: map[string]ingest.Ingestor{
						tableName: mockIngestor,
					},
					serializer: serializer,
				}

				errorInWrite := plugin.Write(tC.inputMetric)

				if tC.expectedWriteError != "" {
					require.EqualError(t, errorInWrite, tC.expectedWriteError)
				} else {
					require.NoError(t, errorInWrite)
					createdIngestor := plugin.ingestors[tableName]
					if len(mockIngestor.records) == 0 {
						fmt.Println(tC.name)
					}
					if tC.metricsGrouping == singleTable {
						createdIngestor = plugin.ingestors[tableName]
					}
					records := mockIngestor.records[0] // the first element
					require.NotNil(t, createdIngestor)
					require.JSONEq(t, jsonValue, records)
				}
			}
		})
	}
}

func TestInitBlankEndpoint(t *testing.T) {
	mockClient := kusto.NewMockClient()
	plugin := AzureDataExplorer{
		Log:       testutil.Logger{},
		client:    mockClient,
		ingestors: make(map[string]ingest.Ingestor),
	}

	errorInit := plugin.Init()
	require.Error(t, errorInit)
	require.Equal(t, "Endpoint configuration cannot be empty", errorInit.Error())
}

type mockIngestor struct {
	records []string
}

func (m *mockIngestor) FromReader(ctx context.Context, reader io.Reader, options ...ingest.FileOption) (*ingest.Result, error) {
	bufbytes, _ := ioutil.ReadAll(reader)
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
