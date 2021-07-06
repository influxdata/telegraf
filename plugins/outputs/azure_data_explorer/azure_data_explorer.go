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
	Endpoint        string          `toml:"endpoint_url"`
	Database        string          `toml:"database"`
	Log             telegraf.Logger `toml:"-"`
	Timeout         config.Duration `toml:"timeout"`
	MetricsGrouping string          `toml:"metrics_grouping_type"`
	TableName       string          `toml:"table_name"`
	client          localClient
	ingesters       map[string]localIngestor
	serializer      serializers.Serializer
	createIngestor  ingestorFactory
}

const (
	TablePerMetric = "TablePerMetric"
	SingleTable    = "SingleTable"
)

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

  ## Type of metrics grouping used when pushing to Azure Data Explorer. 
  ## Default is "TablePerMetric" for one table per different metric. 
  ## For more information, please check the plugin README.
  # metrics_grouping_type = "TablePerMetric"

  ## Name of the single table to store all the metrics (Only needed if metrics_grouping_type is "SingleTable").
  # table_name = ""

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
	if adx.MetricsGrouping == TablePerMetric {
		return adx.writeTablePerMetric(metrics)
	} else {
		return adx.writeSingleTable(metrics)
	}
}

func (adx *AzureDataExplorer) writeTablePerMetric(metrics []telegraf.Metric) error {
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
		if err := adx.pushMetrics(ctx, format, namespace, mPerNamespace); err != nil {
			return err
		}
	}

	return nil
}

func (adx *AzureDataExplorer) writeSingleTable(metrics []telegraf.Metric) error {
	//serialise each metric in metrics - store in byte[]
	metricsArray := make([]byte, 0)
	for _, m := range metrics {
		metricsInBytes, err := adx.serializer.Serialize(m)
		if err != nil {
			return err
		}
		metricsArray = append(metricsArray, metricsInBytes...)
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(adx.Timeout))
	defer cancel()

	//push metrics to a single table
	format := ingest.FileFormat(ingest.JSON)
	if err := adx.pushMetrics(ctx, format, adx.TableName, metricsArray); err != nil {
		return err
	}

	return nil
}

func (adx *AzureDataExplorer) pushMetrics(ctx context.Context, format ingest.FileOption, namespace string, metricsArray []byte) error {
	ingestor, err := adx.getIngestor(ctx, namespace)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(metricsArray)
	mapping := ingest.IngestionMappingRef(fmt.Sprintf("%s_mapping", namespace), ingest.JSON)
	if _, err := ingestor.FromReader(ctx, reader, format, mapping); err != nil {
		adx.Log.Errorf("sending ingestion request to Azure Data Explorer for metric %q failed: %v", namespace, err)
	}
	return nil
}

func (adx *AzureDataExplorer) getIngestor(ctx context.Context, namespace string) (localIngestor, error) {
	ingestor := adx.ingesters[namespace]

	if ingestor == nil {
		if err := adx.createAzureDataExplorerTableForNamespace(ctx, namespace); err != nil {
			return nil, fmt.Errorf("creating table for %q failed: %v", namespace, err)
		}
		//create a new ingestor client for the namespace
		tempIngestor, err := adx.createIngestor(adx.client, adx.Database, namespace)
		if err != nil {
			return nil, fmt.Errorf("creating ingestor for %q failed: %v", namespace, err)
		} else {
			adx.ingesters[namespace] = tempIngestor
			ingestor = tempIngestor
		}
	}
	return ingestor, nil
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

	if adx.MetricsGrouping == SingleTable && adx.TableName == "" {
		return errors.New("Table name cannot be empty for SingleTable metrics grouping type")
	}
	if adx.MetricsGrouping == "" {
		adx.MetricsGrouping = TablePerMetric
	}
	if !(adx.MetricsGrouping == SingleTable || adx.MetricsGrouping == TablePerMetric || adx.MetricsGrouping == "") {
		return errors.New("Metrics grouping type is not valid")
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
