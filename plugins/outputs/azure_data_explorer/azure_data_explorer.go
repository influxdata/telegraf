//go:generate ../../../tools/readme_config_includer/generator
package azure_data_explorer

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/Azure/azure-kusto-go/kusto/ingest"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	common_adx "github.com/influxdata/telegraf/plugins/common/adx"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers/json"
)

//go:embed sample.conf
var sampleConfig string

type AzureDataExplorer struct {
	Log telegraf.Logger `toml:"-"`
	common_adx.Config

	serializer telegraf.Serializer
	client     *common_adx.Client
}

func (*AzureDataExplorer) SampleConfig() string {
	return sampleConfig
}

func (adx *AzureDataExplorer) Init() error {
	serializer := &json.Serializer{
		TimestampUnits:  config.Duration(time.Nanosecond),
		TimestampFormat: time.RFC3339Nano,
	}
	if err := serializer.Init(); err != nil {
		return err
	}
	adx.serializer = serializer
	return nil
}

func (adx *AzureDataExplorer) Connect() error {
	var err error
	if adx.client, err = adx.Config.NewClient("Kusto.Telegraf", adx.Log); err != nil {
		return fmt.Errorf("creating new client failed: %w", err)
	}
	return nil
}

func (adx *AzureDataExplorer) Close() error {
	return adx.client.Close()
}

func (adx *AzureDataExplorer) Write(metrics []telegraf.Metric) error {
	if adx.MetricsGrouping == common_adx.TablePerMetric {
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

	// Push the metrics for each table
	format := ingest.FileFormat(ingest.JSON)
	for tableName, tableMetrics := range tableMetricGroups {
		if err := adx.client.PushMetrics(format, tableName, tableMetrics); err != nil {
			return err
		}
	}

	return nil
}

func (adx *AzureDataExplorer) writeSingleTable(metrics []telegraf.Metric) error {
	// serialise each metric in metrics - store in byte[]
	metricsArray := make([]byte, 0)
	for _, m := range metrics {
		metricsInBytes, err := adx.serializer.Serialize(m)
		if err != nil {
			return err
		}
		metricsArray = append(metricsArray, metricsInBytes...)
	}

	// push metrics to a single table
	format := ingest.FileFormat(ingest.JSON)
	err := adx.client.PushMetrics(format, adx.TableName, metricsArray)
	return err
}

func init() {
	outputs.Add("azure_data_explorer", func() telegraf.Output {
		return &AzureDataExplorer{
			Config: common_adx.Config{
				CreateTables: true,
				Timeout:      config.Duration(20 * time.Second)},
		}
	})
}
