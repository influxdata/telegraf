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
	telegrafJson "github.com/influxdata/telegraf/plugins/serializers/json"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const createTableCommandExpected = `.create-merge table ['%s']  (['fields']:dynamic, ['name']:string, ['tags']:dynamic, ['timestamp']:datetime);`
const createTableMappingCommandExpected = `.create-or-alter table ['%s'] ingestion json mapping '%s_mapping' '[{"column":"fields", "Properties":{"Path":"$[\'fields\']"}},{"column":"name", "Properties":{"Path":"$[\'name\']"}},{"column":"tags", "Properties":{"Path":"$[\'tags\']"}},{"column":"timestamp", "Properties":{"Path":"$[\'timestamp\']"}}]'`

func TestWrite(t *testing.T) {
	fakeClientInstance := &fakeClient{
		queries: make([]string, 0),
	}
	plugin := AzureDataExplorer{
		Endpoint:       "someendpoint",
		Database:       "databasename",
		ClientID:       "longclientid",
		ClientSecret:   "longclientsecret",
		TenantID:       "longtenantid",
		Log:            testutil.Logger{},
		client:         fakeClientInstance,
		ingesters:      map[string]localIngestor{},
		createIngestor: createFakeIngestor,
	}

	serializer, _ := telegrafJson.NewSerializer(time.Second)
	plugin.serializer = serializer

	require.NoError(t, plugin.Write(testutil.MockMetrics()))

	expectedNameOfMetric := "test1"
	createdIngestor := plugin.ingesters["test1"]
	require.NotNil(t, createdIngestor)
	createdFakeIngestor := createdIngestor.(*fakeIngestor)
	require.Equal(t, expectedNameOfMetric, createdFakeIngestor.actualOutputMetric["name"])

	createTableString := fmt.Sprintf(createTableCommandExpected, expectedNameOfMetric)
	require.Equal(t, createTableString, fakeClientInstance.queries[0])

	createTableMappingString := fmt.Sprintf(createTableMappingCommandExpected, expectedNameOfMetric, expectedNameOfMetric)
	require.Equal(t, createTableMappingString, fakeClientInstance.queries[1])
}

func TestWriteBlankEndpoint(t *testing.T) {
	plugin := AzureDataExplorer{
		Endpoint:       "",
		Database:       "",
		ClientID:       "",
		ClientSecret:   "",
		TenantID:       "",
		Log:            testutil.Logger{},
		client:         &fakeClient{},
		ingesters:      map[string]localIngestor{},
		createIngestor: createFakeIngestor,
	}

	errorInit := plugin.Init()
	require.Error(t, errorInit)
	require.Equal(t, "Endpoint configuration cannot be empty", errorInit.Error())
}

func TestWriteErrorInMgmt(t *testing.T) {
	plugin := AzureDataExplorer{
		Endpoint:       "s",
		Database:       "s",
		ClientID:       "s",
		ClientSecret:   "s",
		TenantID:       "s",
		Log:            testutil.Logger{},
		client:         &fakeClientMgmtProduceError{},
		ingesters:      map[string]localIngestor{},
		createIngestor: createFakeIngestor,
	}

	serializer, _ := telegrafJson.NewSerializer(time.Second)
	plugin.serializer = serializer

	errorWrite := plugin.Write(testutil.MockMetrics())
	require.Error(t, errorWrite)
	require.Equal(t, "Something went wrong", errorWrite.Error())
}

type fakeClient struct {
	queries []string
}

func (f *fakeClient) Mgmt(ctx context.Context, db string, query kusto.Stmt, options ...kusto.MgmtOption) (*kusto.RowIterator, error) {
	f.queries = append(f.queries, query.String())
	return &kusto.RowIterator{}, nil
}

type fakeClientMgmtProduceError struct{}

func (f *fakeClientMgmtProduceError) Mgmt(ctx context.Context, db string, query kusto.Stmt, options ...kusto.MgmtOption) (*kusto.RowIterator, error) {
	return nil, errors.New("Something went wrong")
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
