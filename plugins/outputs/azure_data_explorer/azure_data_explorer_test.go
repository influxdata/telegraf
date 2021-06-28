package azure_data_explorer

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var logger testutil.Logger = testutil.Logger{}
var actualOutputMetric map[string]interface{}
var queriesSentToAzureDataExplorer = make([]string, 0)

const createTableCommandExpected = `.create-merge table ['%s']  (['fields']:dynamic, ['name']:string, ['tags']:dynamic, ['timestamp']:datetime);`
const createTableMappingCommandExpected = `.create-or-alter table ['%s'] ingestion json mapping '%s_mapping' '[{"column":"fields", "Properties":{"Path":"$[\'fields\']"}},{"column":"name", "Properties":{"Path":"$[\'name\']"}},{"column":"tags", "Properties":{"Path":"$[\'tags\']"}},{"column":"timestamp", "Properties":{"Path":"$[\'timestamp\']"}}]'`

func TestWrite(t *testing.T) {
	plugin := AzureDataExplorer{
		Endpoint:     "someendpoint",
		Database:     "databasename",
		ClientID:     "longclientid",
		ClientSecret: "longclientsecret",
		TenantID:     "longtenantid",
		Log:          logger,
		client:       &kusto.Client{},
		ingesters:    map[string]localIngestor{},
	}

	createClient = createFakeClient
	createIngestor = createFakeIngestor

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	require.NoError(t, plugin.Write(testutil.MockMetrics()))

	expectedNameOfMetric := "test1"
	require.Equal(t, expectedNameOfMetric, actualOutputMetric["name"])

	createTableString := fmt.Sprintf(createTableCommandExpected, expectedNameOfMetric)
	require.Equal(t, createTableString, queriesSentToAzureDataExplorer[0])

	createTableMappingString := fmt.Sprintf(createTableMappingCommandExpected, expectedNameOfMetric, expectedNameOfMetric)
	require.Equal(t, createTableMappingString, queriesSentToAzureDataExplorer[1])
}

func TestWriteBlankEndpoint(t *testing.T) {
	plugin := AzureDataExplorer{
		Endpoint:     "",
		Database:     "",
		ClientID:     "",
		ClientSecret: "",
		TenantID:     "",
		Log:          logger,
		client:       &kusto.Client{},
		ingesters:    map[string]localIngestor{},
	}

	createClient = createFakeClient
	createIngestor = createFakeIngestor

	errorInit := plugin.Init()
	require.Error(t, errorInit)
	require.Equal(t, "Endpoint configuration cannot be empty", errorInit.Error())
}

func TestWriteErrorInMgmt(t *testing.T) {
	plugin := AzureDataExplorer{
		Endpoint:     "s",
		Database:     "s",
		ClientID:     "s",
		ClientSecret: "s",
		TenantID:     "s",
		Log:          logger,
		client:       &kusto.Client{},
		ingesters:    map[string]localIngestor{},
	}

	createClient = func(endpoint string, clientID string, clientSecret string, tenantID string) (localClient, error) {
		return &fakeClientMgmtProduceError{}, nil
	}

	createIngestor = createFakeIngestor

	errorInit := plugin.Init()
	if errorInit != nil {
		t.Errorf(errorInit.Error())
	}

	errorConnect := plugin.Connect()
	if errorConnect != nil {
		t.Errorf(errorConnect.Error())
	}

	errorWrite := plugin.Write(testutil.MockMetrics())
	require.Error(t, errorWrite)
	require.Equal(t, "Something went wrong", errorWrite.Error())
}

func createFakeIngestor(client localClient, database string, namespace string) (localIngestor, error) {
	return &fakeIngestor{}, nil
}

func createFakeClient(endpoint string, clientID string, clientSecret string, tenantID string) (localClient, error) {
	return &fakeClient{}, nil
}

type fakeClient struct{}

func (f *fakeClient) Mgmt(ctx context.Context, db string, query kusto.Stmt, options ...kusto.MgmtOption) (*kusto.RowIterator, error) {
	queriesSentToAzureDataExplorer = append(queriesSentToAzureDataExplorer, query.String())
	return &kusto.RowIterator{}, nil
}

type fakeClientMgmtProduceError struct{}

func (f *fakeClientMgmtProduceError) Mgmt(ctx context.Context, db string, query kusto.Stmt, options ...kusto.MgmtOption) (*kusto.RowIterator, error) {
	return nil, errors.New("Something went wrong")
}

type fakeIngestor struct{}

func (f *fakeIngestor) FromReader(ctx context.Context, reader io.Reader, options ...ingest.FileOption) (*ingest.Result, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Scan()
	firstLine := scanner.Text()
	err := json.Unmarshal([]byte(firstLine), &actualOutputMetric)
	if err != nil {
		return nil, err
	}
	return &ingest.Result{}, nil
}
