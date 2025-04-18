package microsoft_fabric

import (
	"fmt"
	"time"

	"github.com/Azure/azure-kusto-go/kusto/ingest"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/adx"
	"github.com/influxdata/telegraf/plugins/serializers/json"
)

type EventHouse struct {
	Config     *adx.Config `toml:"cluster_config"`
	client     *adx.Client
	log        telegraf.Logger
	serializer telegraf.Serializer
}

func (e *EventHouse) Init() error {
	serializer := &json.Serializer{
		TimestampUnits:  config.Duration(time.Nanosecond),
		TimestampFormat: time.RFC3339Nano,
	}
	if err := serializer.Init(); err != nil {
		return err
	}
	e.serializer = serializer
	return nil
}

func (e *EventHouse) Connect() error {
	var err error
	if e.client, err = e.Config.NewClient("Kusto.Telegraf", e.log); err != nil {
		return fmt.Errorf("creating new client failed: %w", err)
	}
	return nil
}

func (e *EventHouse) Write(metrics []telegraf.Metric) error {
	if e.Config.MetricsGrouping == adx.TablePerMetric {
		return e.writeTablePerMetric(metrics)
	}
	return e.writeSingleTable(metrics)
}

func (e *EventHouse) Close() error {
	return e.client.Close()
}

func (e *EventHouse) writeTablePerMetric(metrics []telegraf.Metric) error {
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

func (e *EventHouse) writeSingleTable(metrics []telegraf.Metric) error {
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
	err := e.client.PushMetrics(format, e.Config.TableName, metricsArray)
	return err
}
