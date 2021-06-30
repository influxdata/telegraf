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
		expected           map[string]interface{}
		expectedWriteError string
	}{
		{
			name:        "Valid metric",
			inputMetric: testutil.MockMetrics(),
			client: &fakeClient{
				queries: make([]string, 0),
				internalMgmt: func(f *fakeClient, ctx context.Context, db string, query kusto.Stmt, options ...kusto.MgmtOption) (*kusto.RowIterator, error) {
					f.queries = append(f.queries, query.String())
					return &kusto.RowIterator{}, nil
				},
			},
			createIngestor: createFakeIngestor,
			expected: map[string]interface{}{
				"metricName":                "test1",
				"createTableCommand":        "",
				"createTableMappingCommand": "",
			},
		},
		{
			name:        "Error in Mgmt",
			inputMetric: testutil.MockMetrics(),
			client: &fakeClient{
				queries: make([]string, 0),
				internalMgmt: func(f *fakeClient, ctx context.Context, db string, query kusto.Stmt, options ...kusto.MgmtOption) (*kusto.RowIterator, error) {
					return nil, errors.New("Something went wrong")
				},
			},
			createIngestor: createFakeIngestor,
			expected: map[string]interface{}{
				"metricName":                "test1",
				"createTableCommand":        "",
				"createTableMappingCommand": "",
			},
			expectedWriteError: "creating table for \"test1\" failed: Something went wrong",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			serializer, err := telegrafJson.NewSerializer(time.Second)
			require.NoError(t, err)

			plugin := AzureDataExplorer{
				Endpoint:       "someendpoint",
				Database:       "databasename",
				Log:            testutil.Logger{},
				client:         tC.client,
				ingesters:      map[string]localIngestor{},
				createIngestor: tC.createIngestor,
				serializer:     serializer,
			}

			errorInWrite := plugin.Write(testutil.MockMetrics())

			if tC.expectedWriteError != "" {
				require.EqualError(t, errorInWrite, tC.expectedWriteError)
			} else {
				require.NoError(t, errorInWrite)

				expectedNameOfMetric := tC.expected["metricName"].(string)
				createdIngestor := plugin.ingesters[expectedNameOfMetric]
				require.NotNil(t, createdIngestor)
				createdFakeIngestor := createdIngestor.(*fakeIngestor)
				require.Equal(t, expectedNameOfMetric, createdFakeIngestor.actualOutputMetric["name"])

				createTableString := fmt.Sprintf(createTableCommandExpected, expectedNameOfMetric)
				require.Equal(t, createTableString, tC.client.queries[0])

				createTableMappingString := fmt.Sprintf(createTableMappingCommandExpected, expectedNameOfMetric, expectedNameOfMetric)
				require.Equal(t, createTableMappingString, tC.client.queries[1])
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

func createFakeIngestor(client localClient, database string, namespace string) (localIngestor, error) {
	return &fakeIngestor{}, nil
}
func (f *fakeIngestor) FromReader(ctx context.Context, reader io.Reader, options ...ingest.FileOption) (*ingest.Result, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Scan()
	firstLine := scanner.Text()
	err := json.Unmarshal([]byte(firstLine), &f.actualOutputMetric)
	if err != nil {
		return nil, err
	}
	return &ingest.Result{}, nil
}
