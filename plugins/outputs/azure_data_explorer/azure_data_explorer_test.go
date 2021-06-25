package azure_data_explorer

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
	"github.com/influxdata/telegraf/testutil"
)

var logger testutil.Logger = testutil.Logger{}
var actualOutputMetric map[string]interface{}
var queriesSentToAzureDataExplorer = make([]string, 0)

const createTableCommandExpected = `.create-merge table ['%s']  (['fields']:dynamic, ['name']:string, ['tags']:dynamic, ['timestamp']:datetime);`
const createTableMappingCommandExpected = `.create-or-alter table ['%s'] ingestion json mapping '%s_mapping' '[{"column":"fields", "Properties":{"Path":"$[\'fields\']"}},{"column":"name", "Properties":{"Path":"$[\'name\']"}},{"column":"tags", "Properties":{"Path":"$[\'tags\']"}},{"column":"timestamp", "Properties":{"Path":"$[\'timestamp\']"}}]'`

func TestWrite(t *testing.T) {
	azureDataExplorerOutput := AzureDataExplorer{
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

	errorInit := azureDataExplorerOutput.Init()
	if errorInit != nil {
		t.Errorf("Error in Init: %s", errorInit)
	}
	errorConnect := azureDataExplorerOutput.Connect()
	if errorConnect != nil {
		t.Errorf("Error in Connect: %s", errorConnect)
	}

	errorWrite := azureDataExplorerOutput.Write(testutil.MockMetrics())
	if errorWrite != nil {
		t.Errorf("Error in Write: %s", errorWrite)
	}

	expectedNameOfMetric := "test1"
	if actualOutputMetric["name"] != expectedNameOfMetric {
		t.Errorf("Error in Write: expected %s, but actual %s", expectedNameOfMetric, actualOutputMetric["name"])
	}

	createTableString := fmt.Sprintf(createTableCommandExpected, expectedNameOfMetric)
	if queriesSentToAzureDataExplorer[0] != createTableString {
		t.Errorf("Error in Write: expected create table query is %s, but actual is %s", createTableString, queriesSentToAzureDataExplorer[0])
	}

	createTableMappingString := fmt.Sprintf(createTableMappingCommandExpected, expectedNameOfMetric, expectedNameOfMetric)
	if queriesSentToAzureDataExplorer[0] != createTableString {
		t.Errorf("Error in Write: expected create table mapping query is %s, but actual is %s", queriesSentToAzureDataExplorer[1], createTableMappingString)
	}
	fields := actualOutputMetric["fields"]
	logger.Debug((fields.(map[string]interface{}))["value"])
}

func createFakeIngestor(client localClient, database string, namespace string) (localIngestor, error) {
	return &fakeIngestor{}, nil
}

func createFakeClient(endpoint string, clientID string, clientSecret string, tenantID string) (localClient, error) {
	return &fakeClient{}, nil
}

type fakeClient struct {
}

func (f *fakeClient) Mgmt(ctx context.Context, db string, query kusto.Stmt, options ...kusto.MgmtOption) (*kusto.RowIterator, error) {
	queriesSentToAzureDataExplorer = append(queriesSentToAzureDataExplorer, query.String())
	return &kusto.RowIterator{}, nil
}

type fakeIngestor struct {
}

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
