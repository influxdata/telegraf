package azure_data_explorer

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	adx_common "github.com/influxdata/telegraf/plugins/common/adx"
	serializers_json "github.com/influxdata/telegraf/plugins/serializers/json"
	"github.com/influxdata/telegraf/testutil"
)

func TestWrite(t *testing.T) {
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
			metricsGrouping: adx_common.TablePerMetric,
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
			metricsGrouping: adx_common.TablePerMetric,
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
			metricsGrouping: adx_common.SingleTable,
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
			metricsGrouping: adx_common.TablePerMetric,
			ingestionType:   adx_common.ManagedIngestion,
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
			serializer := &serializers_json.Serializer{}
			require.NoError(t, serializer.Init())

			ingestionType := "queued"
			if tC.ingestionType != "" {
				ingestionType = tC.ingestionType
			}

			localFakeIngestor := &fakeIngestor{}
			plugin := AzureDataExplorer{
				Config: adx_common.Config{
					Endpoint:        "someendpoint",
					Database:        "databasename",
					MetricsGrouping: tC.metricsGrouping,
					TableName:       tC.tableName,
					CreateTables:    tC.createTables,
					IngestionType:   ingestionType,
				},
				serializer: serializer,
				Log:        testutil.Logger{},
			}
			plugin.Connect()
			plugin.Client.SetLogger(plugin.Log)
			plugin.Client.SetKustoClient(kusto.NewMockClient())
			plugin.Client.SetMetricsIngestors(map[string]ingest.Ingestor{
				tC.tableName: localFakeIngestor,
			})
			errorInWrite := plugin.Write(testutil.MockMetrics())

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
				require.InDelta(t, expectedTime, createdFakeIngestor.actualOutputMetric["timestamp"], testutil.DefaultDelta)
			}
		})
	}
}

func TestWriteWithType(t *testing.T) {
	metricName := "test1"
	fakeClient := kusto.NewMockClient()
	expectedResultMap := map[string]string{metricName: `{"fields":{"value":1},"name":"test1","tags":{"tag1":"value1"},"timestamp":1257894000}`}
	mockMetrics := testutil.MockMetrics()
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
			metricsGrouping:           adx_common.TablePerMetric,
			tableNameToExpectedResult: expectedResultMap,
		},
		{
			name:                      "Don't create tables'",
			inputMetric:               mockMetrics,
			createTables:              false,
			metricsGrouping:           adx_common.TablePerMetric,
			tableNameToExpectedResult: expectedResultMap,
		},
		{
			name:                      "SingleTable metric grouping type",
			inputMetric:               mockMetrics,
			createTables:              true,
			metricsGrouping:           adx_common.SingleTable,
			tableNameToExpectedResult: expectedResultMap,
		},
		{
			name:                      "Valid metric managed ingestion",
			inputMetric:               mockMetrics,
			createTables:              true,
			metricsGrouping:           adx_common.TablePerMetric,
			tableNameToExpectedResult: expectedResultMap,
			ingestionType:             adx_common.ManagedIngestion,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			serializer := &serializers_json.Serializer{}
			require.NoError(t, serializer.Init())
			for tableName, jsonValue := range testCase.tableNameToExpectedResult {
				ingestionType := "queued"
				if testCase.ingestionType != "" {
					ingestionType = testCase.ingestionType
				}
				mockIngestor := &mockIngestor{}
				plugin := AzureDataExplorer{
					Config: adx_common.Config{
						Endpoint:        "someendpoint",
						Database:        "databasename",
						IngestionType:   ingestionType,
						MetricsGrouping: testCase.metricsGrouping,
						TableName:       tableName,
						CreateTables:    testCase.createTables,
					},
					serializer: serializer,
					Log:        testutil.Logger{},
				}
				plugin.Connect()
				plugin.Client.SetLogger(plugin.Log)
				plugin.Client.SetKustoClient(fakeClient)
				plugin.Client.SetMetricsIngestors(map[string]ingest.Ingestor{
					tableName: mockIngestor,
				})

				err := plugin.Write(testCase.inputMetric)
				if testCase.expectedWriteError != "" {
					require.EqualError(t, err, testCase.expectedWriteError)
					continue
				}
				require.NoError(t, err)

				records := mockIngestor.records[0] // the first element
				require.JSONEq(t, jsonValue, records)
			}
		})
	}
}

func TestInit(t *testing.T) {
	plugin := AzureDataExplorer{
		Log:    testutil.Logger{},
		Client: &adx_common.Client{},
		Config: adx_common.Config{
			Endpoint: "someendpoint",
		},
	}

	err := plugin.Init()
	require.NoError(t, err)
}

func TestConnectBlankEndpointData(t *testing.T) {
	plugin := AzureDataExplorer{
		Log:    testutil.Logger{},
		Client: &adx_common.Client{},
		Config: adx_common.Config{
			Endpoint: "",
		},
	}
	err := plugin.Connect()
	require.Error(t, err)
	require.Equal(t, "Error creating new client. Error: endpoint configuration cannot be empty", err.Error())
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

func (*fakeIngestor) FromFile(context.Context, string, ...ingest.FileOption) (*ingest.Result, error) {
	return &ingest.Result{}, nil
}

func (*fakeIngestor) Close() error {
	return nil
}

type mockIngestor struct {
	records []string
}

func (m *mockIngestor) FromReader(_ context.Context, reader io.Reader, _ ...ingest.FileOption) (*ingest.Result, error) {
	bufbytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	metricjson := string(bufbytes)
	m.SetRecords(strings.Split(metricjson, "\n"))
	return &ingest.Result{}, nil
}

func (*mockIngestor) FromFile(context.Context, string, ...ingest.FileOption) (*ingest.Result, error) {
	return &ingest.Result{}, nil
}

func (m *mockIngestor) SetRecords(records []string) {
	m.records = records
}

// Name receives a copy of Foo since it doesn't need to modify it.
func (m *mockIngestor) Records() []string {
	return m.records
}

func (*mockIngestor) Close() error {
	return nil
}
