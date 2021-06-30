package azure_data_explorer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
	"github.com/Azure/azure-kusto-go/kusto/unsafe"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/plugins/serializers/json"
)

type AzureDataExplorer struct {
	Endpoint       string          `toml:"endpoint_url"`
	Database       string          `toml:"database"`
	Log            telegraf.Logger `toml:"-"`
	Timeout        config.Duration `toml:"timeout"`
	client         localClient
	ingesters      map[string]localIngestor
	serializer     serializers.Serializer
	createIngestor ingestorFactory
}

type localIngestor interface {
	FromReader(ctx context.Context, reader io.Reader, options ...ingest.FileOption) (*ingest.Result, error)
}

type localClient interface {
	Mgmt(ctx context.Context, db string, query kusto.Stmt, options ...kusto.MgmtOption) (*kusto.RowIterator, error)
}

type ingestorFactory func(localClient, string, string) (localIngestor, error)

const createTableCommand = `.create-merge table ['%s']  (['fields']:dynamic, ['name']:string, ['tags']:dynamic, ['timestamp']:datetime);`
const createTableMappingCommand = `.create-or-alter table ['%s'] ingestion json mapping '%s_mapping' '[{"column":"fields", "Properties":{"Path":"$[\'fields\']"}},{"column":"name", "Properties":{"Path":"$[\'name\']"}},{"column":"tags", "Properties":{"Path":"$[\'tags\']"}},{"column":"timestamp", "Properties":{"Path":"$[\'timestamp\']"}}]'`

func (adx *AzureDataExplorer) Description() string {
	return "Sends metrics to Azure Data Explorer"
}

func (adx *AzureDataExplorer) SampleConfig() string {
	return `
  ## Azure Data Exlorer cluster endpoint
  ## ex: endpoint_url = "https://clustername.australiasoutheast.kusto.windows.net"
  endpoint_url = ""
  
  ## The Azure Data Explorer database that the metrics will be ingested into.
  ## The plugin will NOT generate this database automatically, it's expected that this database already exists before ingestion.
  ## ex: "exampledatabase"
  database = ""

  ## Timeout for Azure Data Explorer operations
  # timeout = "15s"
`
}

func (adx *AzureDataExplorer) Connect() error {
	authorizer, err := auth.NewAuthorizerFromEnvironmentWithResource(adx.Endpoint)
	if err != nil {
		return err
	}
	authorization := kusto.Authorization{
		Authorizer: authorizer,
	}
	client, err := kusto.New(adx.Endpoint, authorization)

	if err != nil {
		return err
	}
	adx.client = client
	adx.ingesters = make(map[string]localIngestor)
	adx.createIngestor = createRealIngestor

	return nil
}

func (adx *AzureDataExplorer) Close() error {
	adx.client = nil
	adx.ingesters = nil

	return nil
}

func (adx *AzureDataExplorer) Write(metrics []telegraf.Metric) error {
	metricsPerNamespace := make(map[string][]byte)
	// Group metrics by name and serialize them
	for _, m := range metrics {
		namespace := m.Name()
		metricInBytes, err := adx.serializer.Serialize(m)
		if err != nil {
			return err
		}
		if existingBytes, ok := metricsPerNamespace[namespace]; ok {
			metricsPerNamespace[namespace] = append(existingBytes, metricInBytes...)
		} else {
			metricsPerNamespace[namespace] = metricInBytes
		}
	}
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(adx.Timeout))
	defer cancel()

	// Push the metrics namespace-wise
	format := ingest.FileFormat(ingest.JSON)
	for namespace, mPerNamespace := range metricsPerNamespace {
		reader := bytes.NewReader(mPerNamespace)
		//create a table for the namespace if it doesn't exist
		if adx.ingesters[namespace] == nil {
			if err := adx.createAzureDataExplorerTableForNamespace(ctx, namespace); err != nil {
				return fmt.Errorf("creating table for %q failed: %v", namespace, err)
			}
			//create a new ingestor client for the namespace
			ingestor, err := adx.createIngestor(adx.client, adx.Database, namespace)
			if err != nil {
				return fmt.Errorf("creating ingestor for %q failed: %v", namespace, err)
			}
			adx.ingesters[namespace] = ingestor
		}

		mapping := ingest.IngestionMappingRef(fmt.Sprintf("%s_mapping", namespace), ingest.JSON)
		if _, err := adx.ingesters[namespace].FromReader(ctx, reader, format, mapping); err != nil {
			adx.Log.Errorf("sending ingestion request to Azure Data Explorer for metric %q failed: %v", namespace, err)
		}
	}
	return nil
}

func (adx *AzureDataExplorer) createAzureDataExplorerTableForNamespace(ctx context.Context, tableName string) error {
	createStmt := kusto.NewStmt("", kusto.UnsafeStmt(unsafe.Stmt{Add: true, SuppressWarning: true})).UnsafeAdd(fmt.Sprintf(createTableCommand, tableName))
	if _, err := adx.client.Mgmt(ctx, adx.Database, createStmt); err != nil {
		return err
	}

	createTableMappingstmt := kusto.NewStmt("", kusto.UnsafeStmt(unsafe.Stmt{Add: true, SuppressWarning: true})).UnsafeAdd(fmt.Sprintf(createTableMappingCommand, tableName, tableName))
	if _, err := adx.client.Mgmt(ctx, adx.Database, createTableMappingstmt); err != nil {
		return err
	}

	return nil
}

func (adx *AzureDataExplorer) Init() error {
	if adx.Endpoint == "" {
		return errors.New("Endpoint configuration cannot be empty")
	}
	if adx.Database == "" {
		return errors.New("Database configuration cannot be empty")
	}
	serializer, err := json.NewSerializer(time.Second)
	if err != nil {
		return err
	}
	adx.serializer = serializer
	return nil
}

func init() {
	outputs.Add("azure_data_explorer", func() telegraf.Output {
		return &AzureDataExplorer{
			Timeout: config.Duration(15 * time.Second),
		}
	})
}

func createRealIngestor(client localClient, database string, namespace string) (localIngestor, error) {
	ingestor, err := ingest.New(client.(*kusto.Client), database, namespace)
	if ingestor != nil {
		return ingestor, nil
	}
	return nil, err
}
