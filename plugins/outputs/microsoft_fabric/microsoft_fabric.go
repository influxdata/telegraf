//go:generate ../../../tools/readme_config_includer/generator
package microsoft_fabric

import (
	_ "embed"
	"errors"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	adx "github.com/influxdata/telegraf/plugins/common/adx"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type MicrosoftFabric struct {
	ConnectionString string          `toml:"connection_string"`
	Log              telegraf.Logger `toml:"-"`
	Eventhouse       *EventHouse     `toml:"eventhouse_conf"`
	Eventhubs        *EventHubs      `toml:"eventhubs_conf"`
	activePlugin     FabricOutput
}

// Close implements telegraf.Output.
func (m *MicrosoftFabric) Close() error {
	if m.activePlugin == nil {
		return errors.New("no active plugin to close")
	}
	return m.activePlugin.Close()
}

// Connect implements telegraf.Output.
func (m *MicrosoftFabric) Connect() error {
	if m.activePlugin == nil {
		return errors.New("no active plugin to connect")
	}
	return m.activePlugin.Connect()
}

// SampleConfig implements telegraf.Output.
func (m *MicrosoftFabric) SampleConfig() string {
	return sampleConfig
}

// Write implements telegraf.Output.
func (m *MicrosoftFabric) Write(metrics []telegraf.Metric) error {
	if m.activePlugin == nil {
		return errors.New("no active plugin to write to")
	}
	return m.activePlugin.Write(metrics)
}

func (m *MicrosoftFabric) Init() error {
	ConnectionString := m.ConnectionString

	if ConnectionString == "" {
		return errors.New("endpoint must not be empty. For Kusto refer : https://learn.microsoft.com/kusto/api/connection-strings/kusto?view=microsoft-fabric for EventHouse refer : https://learn.microsoft.com/fabric/real-time-intelligence/event-streams/add-manage-eventstream-sources?pivots=enhanced-capabilities")
	}

	if strings.HasPrefix(ConnectionString, "Endpoint=sb") {
		m.Log.Info("Detected EventHouse endpoint, using EventHouse output plugin")
		m.Eventhubs.connectionString = ConnectionString
		m.Eventhubs.log = m.Log
		m.Eventhubs.Init()
		m.activePlugin = m.Eventhubs
	} else if isKustoEndpoint(strings.ToLower(ConnectionString)) {
		m.Log.Info("Detected Kusto endpoint, using Kusto output plugin")
		//Setting up the AzureDataExplorer plugin initial properties
		m.Eventhouse.Config.Endpoint = ConnectionString
		m.Eventhouse.log = m.Log
		m.Eventhouse.Init()
		m.activePlugin = m.Eventhouse
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
			Eventhubs: &EventHubs{
				Timeout: config.Duration(30 * time.Second),
			},
			Eventhouse: &EventHouse{
				Config: &adx.Config{
					Timeout: config.Duration(30 * time.Second),
				},
			},
		}
	})
}
