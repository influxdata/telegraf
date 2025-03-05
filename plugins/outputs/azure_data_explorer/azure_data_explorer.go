//go:generate ../../../tools/readme_config_includer/generator
package azure_data_explorer

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/Azure/azure-kusto-go/kusto/ingest"
	"github.com/influxdata/telegraf/config"
	adx_common "github.com/influxdata/telegraf/plugins/common/adx"
	"github.com/influxdata/telegraf/plugins/serializers/json"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type AzureDataExplorer struct {
	adx_common.Config
	Log        telegraf.Logger `toml:"-"`
	serializer telegraf.Serializer
	Client     *adx_common.Client
}

func (*AzureDataExplorer) SampleConfig() string {
	return sampleConfig
}

func (adx *AzureDataExplorer) SetSerializer(serializer telegraf.Serializer) {
	adx.serializer = serializer
}

// Initialize the client and the ingestor
func (adx *AzureDataExplorer) Connect() error {
	var err error
	if adx.Client, err = adx.Config.NewClient("Kusto.Telegraf", adx.Log); err != nil {
		return fmt.Errorf("Error creating new client. Error: %w", err)
	}
	adx.Client.SetLogger(adx.Log)
	return nil
}

// Clean up and close the ingestor
func (adx *AzureDataExplorer) Close() error {
	return adx.Client.Close()
}

func (adx *AzureDataExplorer) Write(metrics []telegraf.Metric) error {
	fmt.Println("Writing metrics to Azure Data Explorer", adx)
	if adx.MetricsGrouping == adx_common.TablePerMetric {
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
		if err := adx.Client.PushMetrics(format, tableName, tableMetrics); err != nil {
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
	err := adx.Client.PushMetrics(format, adx.TableName, metricsArray)
	return err
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
func init() {
	outputs.Add("azure_data_explorer", func() telegraf.Output {
		return &AzureDataExplorer{
			Config: adx_common.Config{
				CreateTables: true,
				Timeout:      config.Duration(20 * time.Second)},
		}
	})
}
