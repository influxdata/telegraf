package azure_data_explorer

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
	"github.com/influxdata/telegraf"
	telegrafJson "github.com/influxdata/telegraf/plugins/serializers/json"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const createTableCommandExpected = `.create-merge table ['%s']  (['fields']:dynamic, ['name']:string, ['tags']:dynamic, ['timestamp']:datetime);`
const createTableMappingCommandExpected = `.create-or-alter table ['%s'] ingestion json mapping '%s_mapping' '[{"column":"fields", "Properties":{"Path":"$[\'fields\']"}},{"column":"name", "Properties":{"Path":"$[\'name\']"}},{"column":"tags", "Properties":{"Path":"$[\'tags\']"}},{"column":"timestamp", "Properties":{"Path":"$[\'timestamp\']"}}]'`

func TestWrite(t *testing.T) {
	testCases := []struct {
		name               string
		inputMetric        []telegraf.Metric
		client             *fakeClient
		createIngestor     ingestorFactory
		metricsGrouping    string
		tableName          string
		expected           map[string]interface{}
		expectedWriteError string
		createTables       bool
	}{
		{
			name:         "Valid metric",
			inputMetric:  testutil.MockMetrics(),
			createTables: true,
			client: &fakeClient{
				queries: make([]string, 0),
				internalMgmt: func(f *fakeClient, ctx context.Context, db string, query kusto.Stmt, options ...kusto.MgmtOption) (*kusto.RowIterator, error) {
					f.queries = append(f.queries, query.String())
					return &kusto.RowIterator{}, nil
				},
			},
			createIngestor:  createFakeIngestor,
			metricsGrouping: tablePerMetric,
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
			name:         "Don't create tables'",
			inputMetric:  testutil.MockMetrics(),
			createTables: false,
			client: &fakeClient{
				queries: make([]string, 0),
				internalMgmt: func(f *fakeClient, ctx context.Context, db string, query kusto.Stmt, options ...kusto.MgmtOption) (*kusto.RowIterator, error) {
					require.Fail(t, "Mgmt shouldn't be called when create_tables is false")
					f.queries = append(f.queries, query.String())
					return &kusto.RowIterator{}, nil
				},
			},
			createIngestor:  createFakeIngestor,
			metricsGrouping: tablePerMetric,
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
			name:         "Error in Mgmt",
			inputMetric:  testutil.MockMetrics(),
			createTables: true,
			client: &fakeClient{
				queries: make([]string, 0),
				internalMgmt: func(f *fakeClient, ctx context.Context, db string, query kusto.Stmt, options ...kusto.MgmtOption) (*kusto.RowIterator, error) {
					return nil, errors.New("Something went wrong")
				},
			},
			createIngestor:  createFakeIngestor,
			metricsGrouping: tablePerMetric,
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
			expectedWriteError: "creating table for \"test1\" failed: Something went wrong",
		},
		{
			name:         "SingleTable metric grouping type",
			inputMetric:  testutil.MockMetrics(),
			createTables: true,
			client: &fakeClient{
				queries: make([]string, 0),
				internalMgmt: func(f *fakeClient, ctx context.Context, db string, query kusto.Stmt, options ...kusto.MgmtOption) (*kusto.RowIterator, error) {
					f.queries = append(f.queries, query.String())
					return &kusto.RowIterator{}, nil
				},
			},
			createIngestor:  createFakeIngestor,
			metricsGrouping: singleTable,
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
			serializer, err := telegrafJson.NewSerializer(time.Second, "")
			require.NoError(t, err)

			plugin := AzureDataExplorer{
				Endpoint:        "someendpoint",
				Database:        "databasename",
				Log:             testutil.Logger{},
				MetricsGrouping: tC.metricsGrouping,
				TableName:       tC.tableName,
				CreateTables:    tC.createTables,
				client:          tC.client,
				ingesters:       map[string]localIngestor{},
				createIngestor:  tC.createIngestor,
				serializer:      serializer,
			}

			errorInWrite := plugin.Write(testutil.MockMetrics())

			if tC.expectedWriteError != "" {
				require.EqualError(t, errorInWrite, tC.expectedWriteError)
			} else {
				require.NoError(t, errorInWrite)

				expectedNameOfMetric := tC.expected["metricName"].(string)
				expectedNameOfTable := expectedNameOfMetric
				createdIngestor := plugin.ingesters[expectedNameOfMetric]

				if tC.metricsGrouping == singleTable {
					expectedNameOfTable = tC.tableName
					createdIngestor = plugin.ingesters[expectedNameOfTable]
				}

				require.NotNil(t, createdIngestor)
				createdFakeIngestor := createdIngestor.(*fakeIngestor)
				require.Equal(t, expectedNameOfMetric, createdFakeIngestor.actualOutputMetric["name"])

				expectedFields := tC.expected["fields"].(map[string]interface{})
				require.Equal(t, expectedFields, createdFakeIngestor.actualOutputMetric["fields"])

				expectedTags := tC.expected["tags"].(map[string]interface{})
				require.Equal(t, expectedTags, createdFakeIngestor.actualOutputMetric["tags"])

				expectedTime := tC.expected["timestamp"].(float64)
				require.Equal(t, expectedTime, createdFakeIngestor.actualOutputMetric["timestamp"])

				if tC.createTables {
					createTableString := fmt.Sprintf(createTableCommandExpected, expectedNameOfTable)
					require.Equal(t, createTableString, tC.client.queries[0])

					createTableMappingString := fmt.Sprintf(createTableMappingCommandExpected, expectedNameOfTable, expectedNameOfTable)
					require.Equal(t, createTableMappingString, tC.client.queries[1])
				} else {
					require.Empty(t, tC.client.queries)
				}
			}
		})
	}
}

func TestInitBlankEndpoint(t *testing.T) {
	plugin := AzureDataExplorer{
		Log:            testutil.Logger{},
		client:         &fakeClient{},
		ingesters:      map[string]localIngestor{},
		createIngestor: createFakeIngestor,
	}

	errorInit := plugin.Init()
	require.Error(t, errorInit)
	require.Equal(t, "Endpoint configuration cannot be empty", errorInit.Error())
}

type fakeClient struct {
	queries      []string
	internalMgmt func(client *fakeClient, ctx context.Context, db string, query kusto.Stmt, options ...kusto.MgmtOption) (*kusto.RowIterator, error)
}

func (f *fakeClient) Mgmt(ctx context.Context, db string, query kusto.Stmt, options ...kusto.MgmtOption) (*kusto.RowIterator, error) {
	return f.internalMgmt(f, ctx, db, query, options...)
}

type fakeIngestor struct {
	actualOutputMetric map[string]interface{}
}

func createFakeIngestor(localClient, string, string) (localIngestor, error) {
	return &fakeIngestor{}, nil
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
