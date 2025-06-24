package microsoft_fabric

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Azure/azure-kusto-go/kusto/ingest"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/adx"
	"github.com/influxdata/telegraf/plugins/serializers/json"
)

type eventhouse struct {
	connectionString string
	adx.Config

	client     *adx.Client
	log        telegraf.Logger
	serializer telegraf.Serializer
}

func (e *eventhouse) init() error {
	// Initialize defaults
	e.CreateTables = true

	// Parse the connection string by splitting it into key-value pairs
	// and extract the extra keys used for plugin configuration
	pairs := strings.Split(e.connectionString, ";")
	for _, pair := range pairs {
		// Skip empty pairs
		if strings.TrimSpace(pair) == "" {
			continue
		}
		// Split each pair into key and value
		k, v, found := strings.Cut(pair, "=")
		if !found {
			return fmt.Errorf("invalid connection string format: %s", pair)
		}

		// Only lowercase the keys as the values might be case sensitive
		k = strings.ToLower(strings.TrimSpace(k))
		v = strings.TrimSpace(v)

		key := strings.ReplaceAll(k, " ", "")
		switch key {
		case "datasource", "addr", "address", "networkaddress", "server":
			e.Endpoint = v
		case "initialcatalog", "database":
			e.Database = v
		case "ingestiontype":
			e.IngestionType = v
		case "tablename":
			e.TableName = v
		case "createtables":
			switch v {
			case "true":
				e.CreateTables = true
			case "false":
				e.CreateTables = false
			default:
				return fmt.Errorf("invalid setting %q for %q", v, k)
			}
		case "metricsgroupingtype":
			if v != adx.TablePerMetric && v != adx.SingleTable {
				return errors.New("metrics grouping type is not valid:" + v)
			}
			e.MetricsGrouping = v
		}
	}

	// Setup the JSON serializer
	serializer := &json.Serializer{
		TimestampUnits:  config.Duration(time.Nanosecond),
		TimestampFormat: time.RFC3339Nano,
	}
	if err := serializer.Init(); err != nil {
		return fmt.Errorf("initializing JSON serializer failed: %w", err)
	}
	e.serializer = serializer

	return nil
}

func (e *eventhouse) Connect() error {
	client, err := e.NewClient("MSFabric.Telegraf", e.log)
	if err != nil {
		return fmt.Errorf("creating new client failed: %w", err)
	}
	e.client = client

	return nil
}

func (e *eventhouse) Write(metrics []telegraf.Metric) error {
	if e.MetricsGrouping == adx.TablePerMetric {
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
	err := e.client.PushMetrics(format, e.TableName, metricsArray)
	return err
}

func isEventhouseEndpoint(endpoint string) bool {
	prefixes := []string{
		"data source=",
		"addr=",
		"address=",
		"network address=",
		"server=",
	}

	ep := strings.ToLower(endpoint)
	return slices.ContainsFunc(prefixes, func(prefix string) bool {
		return strings.HasPrefix(ep, prefix)
	})
}
