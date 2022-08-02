package azure_data_explorer

import (
	"context"
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
	fakeClient := kusto.NewMockClient()
	expectedResultMap := map[string]string{metricName: `{"fields":{"value":1},"name":"test1","tags":{"tag1":"value1"},"timestamp":1257894000}`}
	mockMetrics := testutil.MockMetrics()
	// Multi tables
	mockMetrics2 := testutil.TestMetric(1.0, "test2")
	mockMetrics3 := testutil.TestMetric(2.0, "test3")
	mockMetricsMulti := make([]telegraf.Metric, 2)
	mockMetricsMulti[0] = mockMetrics2
	mockMetricsMulti[1] = mockMetrics3
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
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
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
					client:          fakeClient,
					ingestors: map[string]ingest.Ingestor{
						tableName: mockIngestor,
					},
					serializer: serializer,
				}
				errorInWrite := plugin.Write(testCase.inputMetric)
				if testCase.expectedWriteError != "" {
					require.EqualError(t, errorInWrite, testCase.expectedWriteError)
				} else {
					require.NoError(t, errorInWrite)
					createdIngestor := plugin.ingestors[tableName]
					if testCase.metricsGrouping == singleTable {
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

func TestSampleConfig(t *testing.T) {
	fakeClient := kusto.NewMockClient()
	t.Parallel()
	plugin := AzureDataExplorer{
		Log:       testutil.Logger{},
		Endpoint:  "someendpoint",
		Database:  "databasename",
		client:    fakeClient,
		ingestors: make(map[string]ingest.Ingestor),
	}
	sampleConfig := plugin.SampleConfig()
	require.NotNil(t, sampleConfig)
	b, err := ioutil.ReadFile("sample.conf") // just pass the file name
	require.Nil(t, err)                      // read should not error out
	expectedString := string(b)
	require.Equal(t, expectedString, sampleConfig)
}

func TestInitValidations(t *testing.T) {
	fakeClient := kusto.NewMockClient()
	testCases := []struct {
		name          string             // name of the test
		adx           *AzureDataExplorer // the struct to test
		expectedError string             // the error to expect
	}{
		{
			name: "empty_endpoint_configuration",
			adx: &AzureDataExplorer{
				Log:       testutil.Logger{},
				Endpoint:  "",
				Database:  "databasename",
				client:    fakeClient,
				ingestors: make(map[string]ingest.Ingestor),
			},
			expectedError: "endpoint configuration cannot be empty",
		},
		{
			name: "empty_database_configuration",
			adx: &AzureDataExplorer{
				Log:       testutil.Logger{},
				Endpoint:  "endpoint",
				Database:  "",
				client:    fakeClient,
				ingestors: make(map[string]ingest.Ingestor),
			},
			expectedError: "database configuration cannot be empty",
		},
		{
			name: "empty_table_configuration",
			adx: &AzureDataExplorer{
				Log:             testutil.Logger{},
				Endpoint:        "endpoint",
				Database:        "database",
				MetricsGrouping: "SingleTable",
				client:          fakeClient,
				ingestors:       make(map[string]ingest.Ingestor),
			},
			expectedError: "table name cannot be empty for SingleTable metrics grouping type",
		},
		{
			name: "incorrect_metrics_grouping",
			adx: &AzureDataExplorer{
				Log:             testutil.Logger{},
				Endpoint:        "endpoint",
				Database:        "database",
				MetricsGrouping: "MultiTable",
				client:          fakeClient,
				ingestors:       make(map[string]ingest.Ingestor),
			},
			expectedError: "metrics grouping type is not valid",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			err := testCase.adx.Init()
			require.Error(t, err)
			require.Equal(t, testCase.expectedError, err.Error())
		})
	}
}

func TestConnect(t *testing.T) {
	t.Parallel()
	fakeClient := kusto.NewMockClient()
	plugin := AzureDataExplorer{
		Log:       testutil.Logger{},
		Endpoint:  "https://sometestcluster.dummyregion.env.kusto.windows.net",
		Database:  "databasename",
		client:    fakeClient,
		ingestors: make(map[string]ingest.Ingestor),
	}

	connection := plugin.Connect()
	require.Error(t, connection)
	require.Equal(t, "MSI not available", connection.Error())
}

func TestInit(t *testing.T) {
	t.Parallel()
	fakeClient := kusto.NewMockClient()
	plugin := AzureDataExplorer{
		Log:       testutil.Logger{},
		Endpoint:  "someendpoint",
		Database:  "databasename",
		client:    fakeClient,
		ingestors: make(map[string]ingest.Ingestor),
	}
	initResponse := plugin.Init()
	require.Equal(t, initResponse, nil)
}

func TestCreateRealIngestorManaged(t *testing.T) {
	t.Parallel()
	kustoLocalClient := kusto.NewMockClient()
	localIngestor, err := createIngestorByTable(kustoLocalClient, "telegrafdb", "metrics", "managed")
	require.Nil(t, err)
	require.NotNil(t, localIngestor)
}

func TestCreateRealIngestorQueued(t *testing.T) {
	t.Parallel()
	kustoLocalClient := kusto.NewMockClient()
	localIngestor, err := createIngestorByTable(kustoLocalClient, "telegrafdb", "metrics", "queued")
	require.Nil(t, err)
	require.NotNil(t, localIngestor)
}

func TestInvalidIngestorType(t *testing.T) {
	t.Parallel()
	kustoLocalClient := kusto.NewMockClient()
	localIngestor, err := createIngestorByTable(kustoLocalClient, "telegrafdb", "metrics", "streaming")
	require.NotNil(t, err)
	require.Nil(t, localIngestor)
	require.Equal(t, "ingestion_type has to be one of managed or queued", err.Error())
}

func TestClose(t *testing.T) {
	t.Parallel()
	fakeClient := kusto.NewMockClient()
	adx := AzureDataExplorer{
		Log:       testutil.Logger{},
		Endpoint:  "someendpoint",
		Database:  "databasename",
		client:    fakeClient,
		ingestors: make(map[string]ingest.Ingestor),
	}
	err := adx.Close()
	require.Nil(t, err)
	// client becomes nil in the end
	require.Nil(t, adx.client)
	require.Nil(t, adx.ingestors)
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
