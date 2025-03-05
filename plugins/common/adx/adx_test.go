package adx

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestInitBlankEndpointData(t *testing.T) {
	plugin := Config{
		Endpoint: "",
		Database: "mydb",
	}

	_, err := plugin.NewClient("TestKusto.Telegraf", nil)
	require.Error(t, err)
	require.Equal(t, "endpoint configuration cannot be empty", err.Error())
}

func TestQueryConstruction(t *testing.T) {
	const tableName = "mytable"
	const expectedCreate = `.create-merge table ['mytable'] (['fields']:dynamic, ['name']:string, ['tags']:dynamic, ['timestamp']:datetime);`
	const expectedMapping = `` +
		`.create-or-alter table ['mytable'] ingestion json mapping 'mytable_mapping' '[{"column":"fields", ` +
		`"Properties":{"Path":"$[\'fields\']"}},{"column":"name", "Properties":{"Path":"$[\'name\']"}},{"column":"tags", ` +
		`"Properties":{"Path":"$[\'tags\']"}},{"column":"timestamp", "Properties":{"Path":"$[\'timestamp\']"}}]'`
	require.Equal(t, expectedCreate, createTableCommand(tableName).String())
	require.Equal(t, expectedMapping, createTableMappingCommand(tableName).String())
}

type fakeIngestor struct {
	actualOutputMetric []map[string]interface{}
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

func (*fakeIngestor) FromFile(_ context.Context, _ string, _ ...ingest.FileOption) (*ingest.Result, error) {
	return &ingest.Result{}, nil
}

func (*fakeIngestor) Close() error {
	return nil
}

func TestGetMetricIngestor(t *testing.T) {
	plugin := Client{
		logger: testutil.Logger{},
		client: kusto.NewMockClient(),
		cfg: &Config{
			Database:      "mydb",
			IngestionType: QueuedIngestion,
		},
	}

	plugin.ingestors = map[string]ingest.Ingestor{
		"test1": &fakeIngestor{},
	}

	ingestor, err := plugin.GetMetricIngestor(context.Background(), "test1")
	if err != nil {
		t.Errorf("Error getting ingestor: %v", err)
	}
	require.NotNil(t, ingestor)
}

func TestGetMetricIngestorNoIngester(t *testing.T) {
	plugin := Client{
		logger: testutil.Logger{},
		client: kusto.NewMockClient(),
		cfg: &Config{
			IngestionType: QueuedIngestion,
		},
	}

	plugin.ingestors = map[string]ingest.Ingestor{}

	ingestor, err := plugin.GetMetricIngestor(context.Background(), "test1")
	if err != nil {
		t.Errorf("Error getting ingestor: %v", err)
	}
	require.NotNil(t, ingestor)
}

func TestPushMetrics(t *testing.T) {
	plugin := Client{
		logger: testutil.Logger{},
		client: kusto.NewMockClient(),
		cfg: &Config{
			Database:      "mydb",
			Endpoint:      "https://ingest-test.westus.kusto.windows.net",
			IngestionType: QueuedIngestion,
		},
	}

	plugin.ingestors = map[string]ingest.Ingestor{
		"test1": &fakeIngestor{},
	}

	metrics := []byte(`[{"fields": {"value": 1}, "name": "test1", "tags": {"tag1": "value1"}, "timestamp": "2021-01-01T00:00:00Z"}]`)
	err := plugin.PushMetrics(ingest.FileFormat(ingest.JSON), "test1", metrics)
	if err != nil {
		t.Errorf("Error pushing metrics: %v", err)
	}
}

func TestClose(t *testing.T) {
	plugin := Client{
		logger: testutil.Logger{},
		cfg: &Config{
			IngestionType: QueuedIngestion,
		},
		client: kusto.NewMockClient(),
	}

	err := plugin.Close()
	require.NoError(t, err)
}
