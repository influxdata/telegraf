//go:generate ../../../tools/readme_config_includer/generator
package microsoft_fabric

import (
	_ "embed"
	"errors"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	adx_commons "github.com/influxdata/telegraf/plugins/common/adx"
	eh_commons "github.com/influxdata/telegraf/plugins/common/eventhub"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers/json"
)

//go:embed sample.conf
var sampleConfig string

type MicrosoftFabric struct {
	ConnectionString  string                         `toml:"connection_string"`
	Log               telegraf.Logger                `toml:"-"`
	ADXConf           *adx_commons.AzureDataExplorer `toml:"adx_conf"`
	EHConf            *eh_commons.EventHubs          `toml:"eh_conf"`
	FabricSinkService telegraf.Output
}

// Close implements telegraf.Output.
func (m *MicrosoftFabric) Close() error {
	return m.FabricSinkService.Close()
}

// Connect implements telegraf.Output.
func (m *MicrosoftFabric) Connect() error {
	return m.FabricSinkService.Connect()
}

// SampleConfig implements telegraf.Output.
func (m *MicrosoftFabric) SampleConfig() string {
	return sampleConfig
}

// Write implements telegraf.Output.
func (m *MicrosoftFabric) Write(metrics []telegraf.Metric) error {
	return m.FabricSinkService.Write(metrics)
}

func (m *MicrosoftFabric) Init() error {
	ConnectionString := m.ConnectionString

	if ConnectionString == "" {
		return errors.New("endpoint must not be empty. For Kusto refer : https://learn.microsoft.com/kusto/api/connection-strings/kusto?view=microsoft-fabric for EventHouse refer : https://learn.microsoft.com/fabric/real-time-intelligence/event-streams/add-manage-eventstream-sources?pivots=enhanced-capabilities")
	}

	if strings.HasPrefix(ConnectionString, "Endpoint=sb") {
		m.Log.Info("Detected EventHouse endpoint, using EventHouse output plugin")

		//Need discussion on it
		serializer := &json.Serializer{
			TimestampUnits:  config.Duration(time.Nanosecond),
			TimestampFormat: time.RFC3339Nano,
		}
		m.EHConf.ConnectionString = ConnectionString
		m.EHConf.Log = m.Log
		m.EHConf.SetSerializer(serializer)
		m.EHConf.Init()
		m.FabricSinkService = m.EHConf
	} else if isKustoEndpoint(strings.ToLower(ConnectionString)) {
		m.Log.Info("Detected Kusto endpoint, using Kusto output plugin")
		//Setting up the AzureDataExplorer plugin initial properties
		m.ADXConf.Endpoint = ConnectionString
		m.ADXConf.Log = m.Log
		m.ADXConf.Init()
		m.FabricSinkService = m.ADXConf
	} else {
		return errors.New("invalid connection string. For Kusto refer : https://learn.microsoft.com/kusto/api/connection-strings/kusto?view=microsoft-fabric for EventHouse refer : https://learn.microsoft.com/fabric/real-time-intelligence/event-streams/add-manage-eventstream-sources?pivots=enhanced-capabilities")
	}
	return nil
}

func isKustoEndpoint(endpoint string) bool {
	prefixes := []string{
		"data source=",
		"addr=",
		"address=",
		"network address=",
		"server=",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(endpoint, prefix) {
			return true
		}
	}
	return false
}

func init() {

	outputs.Add("microsoft_fabric", func() telegraf.Output {
		return &MicrosoftFabric{
			ADXConf: &adx_commons.AzureDataExplorer{
				Timeout:      config.Duration(20 * time.Second),
				CreateTables: true,
				AppName:      "Fabric.Telegraf",
			},
			EHConf: &eh_commons.EventHubs{
				Hub:     &eh_commons.EventHub{},
				Timeout: config.Duration(30 * time.Second),
			},
		}
	})
}
