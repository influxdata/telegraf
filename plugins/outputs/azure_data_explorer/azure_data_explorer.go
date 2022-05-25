package azure_data_explorer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
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
	CreateTables    bool            `toml:"create_tables"`
	client          localClient
	ingesters       map[string]localIngestor
	serializer      serializers.Serializer
	createIngestor  ingestorFactory
}

const (
	tablePerMetric = "tablepermetric"
	singleTable    = "singletable"
	// These control the amount of memory we use when ingesting blobs
	bufferSize = 1 << 20 // 1 MiB
	maxBuffers = 5
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
	if adx.MetricsGrouping == tablePerMetric {
		return adx.writeTablePerMetric(metrics)
	}
	return adx.writeSingleTable(metrics)
}

func (adx *AzureDataExplorer) writeTablePerMetric(metrics []telegraf.Metric) error {
	tableMetricGroups := make(map[string][]byte)
	// Group metrics by name and serialize them
	for _, m := range metrics {
		tableName := m.Name()
		metricInBytes, err := adx.serializer.Serialize(m)
		if err != nil {
			return err
		}
		if existingBytes, ok := tableMetricGroups[tableName]; ok {
			tableMetricGroups[tableName] = append(existingBytes, metricInBytes...)
		} else {
			tableMetricGroups[tableName] = metricInBytes
		}
	}
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(adx.Timeout))
	defer cancel()

	// Push the metrics for each table
	format := ingest.FileFormat(ingest.JSON)
	for tableName, tableMetrics := range tableMetricGroups {
		if err := adx.pushMetrics(ctx, format, tableName, tableMetrics); err != nil {
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
	err := adx.pushMetrics(ctx, format, adx.TableName, metricsArray)
	return err
}

func (adx *AzureDataExplorer) pushMetrics(ctx context.Context, format ingest.FileOption, tableName string, metricsArray []byte) error {
	ingestor, err := adx.getIngestor(ctx, tableName)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(metricsArray)
	mapping := ingest.IngestionMappingRef(fmt.Sprintf("%s_mapping", tableName), ingest.JSON)
	if _, err := ingestor.FromReader(ctx, reader, format, mapping); err != nil {
		adx.Log.Errorf("sending ingestion request to Azure Data Explorer for table %q failed: %v", tableName, err)
	}
	return nil
}

func (adx *AzureDataExplorer) getIngestor(ctx context.Context, tableName string) (localIngestor, error) {
	ingestor := adx.ingesters[tableName]

	if ingestor == nil {
		if err := adx.createAzureDataExplorerTable(ctx, tableName); err != nil {
			return nil, fmt.Errorf("creating table for %q failed: %v", tableName, err)
		}
		//create a new ingestor client for the table
		tempIngestor, err := adx.createIngestor(adx.client, adx.Database, tableName)
		if err != nil {
			return nil, fmt.Errorf("creating ingestor for %q failed: %v", tableName, err)
		}
		adx.ingesters[tableName] = tempIngestor
		ingestor = tempIngestor
	}
	return ingestor, nil
}

func (adx *AzureDataExplorer) createAzureDataExplorerTable(ctx context.Context, tableName string) error {
	if !adx.CreateTables {
		adx.Log.Info("skipped table creation")
		return nil
	}
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

	adx.MetricsGrouping = strings.ToLower(adx.MetricsGrouping)
	if adx.MetricsGrouping == singleTable && adx.TableName == "" {
		return errors.New("Table name cannot be empty for SingleTable metrics grouping type")
	}
	if adx.MetricsGrouping == "" {
		adx.MetricsGrouping = tablePerMetric
	}
	if !(adx.MetricsGrouping == singleTable || adx.MetricsGrouping == tablePerMetric) {
		return errors.New("Metrics grouping type is not valid")
	}

	serializer, err := json.NewSerializer(time.Nanosecond, time.RFC3339Nano)
	if err != nil {
		return err
	}
	adx.serializer = serializer
	return nil
}

func init() {
	outputs.Add("azure_data_explorer", func() telegraf.Output {
		return &AzureDataExplorer{
			Timeout:      config.Duration(20 * time.Second),
			CreateTables: true,
		}
	})
}

func createRealIngestor(client localClient, database string, tableName string) (localIngestor, error) {
	ingestor, err := ingest.New(client.(*kusto.Client), database, tableName, ingest.WithStaticBuffer(bufferSize, maxBuffers))
	if ingestor != nil {
		return ingestor, nil
	}
	return nil, err
}
