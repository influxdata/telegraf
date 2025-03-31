package azure_data_explorer

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	common_adx "github.com/influxdata/telegraf/plugins/common/adx"
	serializers_json "github.com/influxdata/telegraf/plugins/serializers/json"
	"github.com/influxdata/telegraf/testutil"
)

func TestWrite(t *testing.T) {
	testCases := []struct {
		name               string
		inputMetric        []telegraf.Metric
		metricsGrouping    string
		tableName          string
		expectedWriteError string
		createTables       bool
		ingestionType      string
	}{
		{
			name:            "Valid metric",
			inputMetric:     testutil.MockMetrics(),
			createTables:    true,
			tableName:       "test1",
			metricsGrouping: common_adx.TablePerMetric,
		},
		{
			name:            "Don't create tables'",
			inputMetric:     testutil.MockMetrics(),
			createTables:    false,
			tableName:       "test1",
			metricsGrouping: common_adx.TablePerMetric,
		},
		{
			name:            "SingleTable metric grouping type",
			inputMetric:     testutil.MockMetrics(),
			createTables:    true,
			tableName:       "test1",
			metricsGrouping: common_adx.SingleTable,
		},
		{
			name:            "Valid metric managed ingestion",
			inputMetric:     testutil.MockMetrics(),
			createTables:    true,
			tableName:       "test1",
			metricsGrouping: common_adx.TablePerMetric,
			ingestionType:   common_adx.ManagedIngestion,
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
				Config: common_adx.Config{
					Endpoint:        "https://someendpoint.kusto.net",
					Database:        "databasename",
					MetricsGrouping: tC.metricsGrouping,
					TableName:       tC.tableName,
					CreateTables:    tC.createTables,
					IngestionType:   ingestionType,
				},
				serializer: serializer,
				Log:        testutil.Logger{},
			}
			plugin.client = NewMockClient(&plugin.Config, plugin.Log, map[string]ingest.Ingestor{
				tC.tableName: localFakeIngestor,
			})
			errorInWrite := plugin.Write(testutil.MockMetrics())

			if tC.expectedWriteError != "" {
				require.EqualError(t, errorInWrite, tC.expectedWriteError)
			} else {
				require.NoError(t, errorInWrite)
				// Moved metric data level test to commons
			}
		})
	}
}

func TestWriteWithType(t *testing.T) {
	metricName := "test1"
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
			metricsGrouping:           common_adx.TablePerMetric,
			tableNameToExpectedResult: expectedResultMap,
		},
		{
			name:                      "Don't create tables'",
			inputMetric:               mockMetrics,
			createTables:              false,
			metricsGrouping:           common_adx.TablePerMetric,
			tableNameToExpectedResult: expectedResultMap,
		},
		{
			name:                      "SingleTable metric grouping type",
			inputMetric:               mockMetrics,
			createTables:              true,
			metricsGrouping:           common_adx.SingleTable,
			tableNameToExpectedResult: expectedResultMap,
		},
		{
			name:                      "Valid metric managed ingestion",
			inputMetric:               mockMetrics,
			createTables:              true,
			metricsGrouping:           common_adx.TablePerMetric,
			tableNameToExpectedResult: expectedResultMap,
			ingestionType:             common_adx.ManagedIngestion,
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
					Config: common_adx.Config{
						Endpoint:        "someendpoint",
						Database:        "databasename",
						IngestionType:   ingestionType,
						MetricsGrouping: testCase.metricsGrouping,
						TableName:       tableName,
						CreateTables:    testCase.createTables,
						Timeout:         config.Duration(20 * time.Second),
					},
					serializer: serializer,
					Log:        testutil.Logger{},
				}

				plugin.client = NewMockClient(&plugin.Config, plugin.Log, map[string]ingest.Ingestor{tableName: mockIngestor})

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
		client: &common_adx.Client{},
		Config: common_adx.Config{
			Endpoint: "someendpoint",
		},
	}

	err := plugin.Init()
	require.NoError(t, err)
}

func TestConnectBlankEndpointData(t *testing.T) {
	plugin := AzureDataExplorer{
		Log: testutil.Logger{},
	}
	require.ErrorContains(t, plugin.Connect(), "endpoint configuration cannot be empty")
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

// MockClient is a mock implementation of the Client struct for testing purposes
type MockClient struct {
	cfg       *common_adx.Config
	conn      *kusto.ConnectionStringBuilder
	client    *kusto.Client
	ingestors map[string]ingest.Ingestor
	logger    telegraf.Logger
}

// NewMockClient creates a new instance of MockClient
func NewMockClient(cfg *common_adx.Config, logger telegraf.Logger, ingestor map[string]ingest.Ingestor) *MockClient {
	return &MockClient{
		cfg:       cfg,
		conn:      kusto.NewConnectionStringBuilder(cfg.Endpoint).WithDefaultAzureCredential(),
		client:    kusto.NewMockClient(),
		ingestors: ingestor,
		logger:    logger,
	}
}

// Mock implementation of the Close method
func (*MockClient) Close() error {
	// Mock behavior for closing the client
	return nil
}

// Mock implementation of the PushMetrics method
func (m *MockClient) PushMetrics(format ingest.FileOption, tableName string, metrics []byte) error {
	// Mock behavior for pushing metrics
	ctx := context.Background()
	metricIngestor := m.ingestors[tableName]
	reader := bytes.NewReader(metrics)
	mapping := ingest.IngestionMappingRef(tableName+"_mapping", ingest.JSON)
	if metricIngestor != nil {
		if _, err := metricIngestor.FromReader(ctx, reader, format, mapping); err != nil {
			return fmt.Errorf("sending ingestion request to Azure Data Explorer for table %q failed: %w", tableName, err)
		}
	}
	return nil
}
