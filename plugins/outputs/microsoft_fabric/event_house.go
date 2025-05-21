package microsoft_fabric

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-kusto-go/kusto/ingest"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/adx"
	"github.com/influxdata/telegraf/plugins/serializers/json"
)

type eventhouse struct {
	config *adx.Config
	client *adx.Client

	log        telegraf.Logger
	serializer telegraf.Serializer
}

func (e *eventhouse) Init() error {
	serializer := &json.Serializer{
		TimestampUnits:  config.Duration(time.Nanosecond),
		TimestampFormat: time.RFC3339Nano,
	}
	if err := serializer.Init(); err != nil {
		return err
	}
	e.serializer = serializer
	e.config = &adx.Config{}
	e.config.CreateTables = true
	return nil
}

func (e *eventhouse) Connect() error {
	client, err := e.config.NewClient("Kusto.Telegraf", e.log)
	if err != nil {
		return fmt.Errorf("creating new client failed: %w", err)
	}
	e.client = client

	return nil
}

func (e *eventhouse) Write(metrics []telegraf.Metric) error {
	if e.config.MetricsGrouping == adx.TablePerMetric {
		return e.writeTablePerMetric(metrics)
	}
	return e.writeSingleTable(metrics)
}

func (e *eventhouse) Close() error {
	return e.client.Close()
}

func (e *eventhouse) writeTablePerMetric(metrics []telegraf.Metric) error {
	tableMetricGroups := make(map[string][]byte)
	// Group metrics by name and serialize them
	for _, m := range metrics {
		tableName := m.Name()
		metricInBytes, err := e.serializer.Serialize(m)
		if err != nil {
			return err
		}
		if existingBytes, ok := tableMetricGroups[tableName]; ok {
			tableMetricGroups[tableName] = append(existingBytes, metricInBytes...)
		} else {
			tableMetricGroups[tableName] = metricInBytes
		}
	}

	// Push the metrics for each table
	format := ingest.FileFormat(ingest.JSON)
	for tableName, tableMetrics := range tableMetricGroups {
		if err := e.client.PushMetrics(format, tableName, tableMetrics); err != nil {
			return err
		}
	}

	return nil
}

func (e *eventhouse) writeSingleTable(metrics []telegraf.Metric) error {
	// serialise each metric in metrics - store in byte[]
	metricsArray := make([]byte, 0)
	for _, m := range metrics {
		metricsInBytes, err := e.serializer.Serialize(m)
		if err != nil {
			return err
		}
		metricsArray = append(metricsArray, metricsInBytes...)
	}

	// push metrics to a single table
	format := ingest.FileFormat(ingest.JSON)
	err := e.client.PushMetrics(format, e.config.TableName, metricsArray)
	return err
}

func (e *eventhouse) parseconnectionString(cs string) error {
	// Parse the connection string to extract the endpoint and database
	if cs == "" {
		return errors.New("connection string must not be empty")
	}
	// Split the connection string into key-value pairs
	pairs := strings.Split(cs, ";")
	for _, pair := range pairs {
		// Split each pair into key and value
		k, v, found := strings.Cut(pair, "=")
		if !found {
			return fmt.Errorf("invalid connection string format: %s", pair)
		}
		k = strings.ToLower(strings.TrimSpace(k))
		v = strings.TrimSpace(v)
		switch k {
		case "data source", "addr", "address", "network address", "server":
			e.config.Endpoint = v
		case "initial catalog", "database":
			e.config.Database = v
		case "ingestion type", "ingestiontype":
			e.config.IngestionType = v
		case "table name", "tablename":
			e.config.TableName = v
		case "create tables", "createtables":
			if v == "false" {
				e.config.CreateTables = false
			} else {
				e.config.CreateTables = true
			}
		case "metrics grouping type, metricsgroupingtype":
			if v != adx.TablePerMetric && v != adx.SingleTable {
				return errors.New("metrics grouping type is not valid:" + v)
			}
			e.config.MetricsGrouping = v
		}
	}
	return nil
}
