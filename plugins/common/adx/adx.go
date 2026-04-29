package adx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-kusto-go/azkustodata"
	kustoerrors "github.com/Azure/azure-kusto-go/azkustodata/errors"
	"github.com/Azure/azure-kusto-go/azkustodata/kql"
	"github.com/Azure/azure-kusto-go/azkustoingest"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
)

const (
	TablePerMetric = "tablepermetric"
	SingleTable    = "singletable"
	// These control the amount of memory we use when ingesting blobs
	bufferSize       = 1 << 20 // 1 MiB
	maxBuffers       = 5
	ManagedIngestion = "managed"
	QueuedIngestion  = "queued"
)

type Config struct {
	Endpoint        string          `toml:"endpoint_url"`
	Database        string          `toml:"database"`
	Timeout         config.Duration `toml:"timeout"`
	MetricsGrouping string          `toml:"metrics_grouping_type"`
	TableName       string          `toml:"table_name"`
	CreateTables    bool            `toml:"create_tables"`
	IngestionType   string          `toml:"ingestion_type"`
}

type Client struct {
	cfg       *Config
	conn      *azkustodata.ConnectionStringBuilder
	client    *azkustodata.Client
	ingestors map[string]azkustoingest.Ingestor
	logger    telegraf.Logger
}

func (cfg *Config) NewClient(app string, log telegraf.Logger) (*Client, error) {
	if cfg.Endpoint == "" {
		return nil, errors.New("endpoint configuration cannot be empty")
	}
	if cfg.Database == "" {
		return nil, errors.New("database configuration cannot be empty")
	}

	cfg.MetricsGrouping = strings.ToLower(cfg.MetricsGrouping)
	if cfg.MetricsGrouping == SingleTable && cfg.TableName == "" {
		return nil, errors.New("table name cannot be empty for SingleTable metrics grouping type")
	}

	if cfg.MetricsGrouping == "" {
		cfg.MetricsGrouping = TablePerMetric
	}

	if cfg.MetricsGrouping != SingleTable && cfg.MetricsGrouping != TablePerMetric {
		return nil, errors.New("metrics grouping type is not valid")
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = config.Duration(20 * time.Second)
	}

	switch cfg.IngestionType {
	case "":
		cfg.IngestionType = QueuedIngestion
	case ManagedIngestion, QueuedIngestion:
		// Do nothing as those are valid
	default:
		return nil, fmt.Errorf("unknown ingestion type %q", cfg.IngestionType)
	}

	conn := azkustodata.NewConnectionStringBuilder(cfg.Endpoint).WithDefaultAzureCredential()
	conn.SetConnectorDetails("Telegraf", internal.ProductToken(), app, "", false, "")
	client, err := azkustodata.New(conn)
	if err != nil {
		return nil, err
	}
	return &Client{
		cfg:       cfg,
		client:    client,
		conn:      conn,
		ingestors: make(map[string]azkustoingest.Ingestor),
		logger:    log,
	}, nil
}

func (adx *Client) Close() error {
	errs := kustoerrors.NewCombinedError()
	for _, v := range adx.ingestors {
		if err := v.Close(); err != nil {
			errs.AddError(err)
		}
	}
	if err := adx.client.Close(); err != nil {
		errs.AddError(err)
	}

	adx.client = nil
	adx.ingestors = nil

	// Combine errors into a single object and return the combined error
	return errs.GetError()
}

func (adx *Client) PushMetrics(format azkustoingest.FileOption, tableName string, metrics []byte) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(adx.cfg.Timeout))
	defer cancel()
	metricIngestor, err := adx.getMetricIngestor(ctx, tableName)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(metrics)
	mapping := azkustoingest.IngestionMappingRef(tableName+"_mapping", azkustoingest.JSON)
	if metricIngestor != nil {
		if _, err := metricIngestor.FromReader(ctx, reader, format, mapping); err != nil {
			return fmt.Errorf("sending ingestion request to Azure Data Explorer for table %q failed: %w", tableName, err)
		}
	}
	return nil
}

func (adx *Client) getMetricIngestor(ctx context.Context, tableName string) (azkustoingest.Ingestor, error) {
	if ingestor := adx.ingestors[tableName]; ingestor != nil {
		return ingestor, nil
	}

	if adx.cfg.CreateTables {
		if _, err := adx.client.Mgmt(ctx, adx.cfg.Database, createTableCommand(tableName)); err != nil {
			return nil, fmt.Errorf("creating table for %q failed: %w", tableName, err)
		}

		if _, err := adx.client.Mgmt(ctx, adx.cfg.Database, createTableMappingCommand(tableName)); err != nil {
			return nil, err
		}
	}

	options := []azkustoingest.Option{
		azkustoingest.WithDefaultDatabase(adx.cfg.Database),
		azkustoingest.WithDefaultTable(tableName),
	}

	// Create a new ingestor client for the table
	var ingestor azkustoingest.Ingestor
	var err error
	switch strings.ToLower(adx.cfg.IngestionType) {
	case ManagedIngestion:
		ingestor, err = azkustoingest.NewManaged(adx.conn, options...)
	case QueuedIngestion:
		options = append(options, azkustoingest.WithStaticBuffer(bufferSize, maxBuffers))
		ingestor, err = azkustoingest.New(adx.conn, options...)
	default:
		return nil, fmt.Errorf(`ingestion_type has to be one of %q or %q`, ManagedIngestion, QueuedIngestion)
	}
	if err != nil {
		return nil, fmt.Errorf("creating ingestor for %q failed: %w", tableName, err)
	}
	adx.ingestors[tableName] = ingestor

	return ingestor, nil
}

func createTableCommand(table string) azkustodata.Statement {
	builder := kql.New(`.create-merge table ['`).AddTable(table).AddLiteral(`'] `)
	builder.AddLiteral(`(['fields']:dynamic, ['name']:string, ['tags']:dynamic, ['timestamp']:datetime);`)

	return builder
}

func createTableMappingCommand(table string) azkustodata.Statement {
	builder := kql.New(`.create-or-alter table ['`).AddTable(table).AddLiteral(`'] `)
	builder.AddLiteral(`ingestion json mapping '`).AddTable(table + "_mapping").AddLiteral(`' `)
	builder.AddLiteral(`'[{"column":"fields", `)
	builder.AddLiteral(`"Properties":{"Path":"$[\'fields\']"}},{"column":"name", `)
	builder.AddLiteral(`"Properties":{"Path":"$[\'name\']"}},{"column":"tags", `)
	builder.AddLiteral(`"Properties":{"Path":"$[\'tags\']"}},{"column":"timestamp", `)
	builder.AddLiteral(`"Properties":{"Path":"$[\'timestamp\']"}}]'`)

	return builder
}
