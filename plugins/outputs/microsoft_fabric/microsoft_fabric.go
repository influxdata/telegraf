//go:generate ../../../tools/readme_config_includer/generator
package microsoft_fabric

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type fabricOutput interface {
	Init() error
	Connect() error
	Write(metrics []telegraf.Metric) error
	Close() error
}

type MicrosoftFabric struct {
	ConnectionString string          `toml:"connection_string"`
	Timeout          config.Duration `toml:"timeout"`
	Log              telegraf.Logger `toml:"-"`

	eventhouse   *eventhouse
	eventstream  *eventstream
	activePlugin fabricOutput
}

func (*MicrosoftFabric) SampleConfig() string {
	return sampleConfig
}

func (m *MicrosoftFabric) Init() error {

	if m.ConnectionString == "" {
		return errors.New("endpoint must not be empty")
	}

	switch {
	case isEventstreamEndpoint(m.ConnectionString):
		m.Log.Info("Detected EventStream endpoint, using EventStream output plugin")
		eventstream := &eventstream{}
		eventstream.connectionString = m.ConnectionString
		eventstream.log = m.Log
		eventstream.timeout = m.Timeout
		if err := eventstream.parseconnectionString(m.ConnectionString); err != nil {
			return fmt.Errorf("parsing connection string failed: %w", err)
		}
		m.eventstream = eventstream
		if err := m.eventstream.Init(); err != nil {
			return fmt.Errorf("initializing EventStream output failed: %w", err)
		}
		m.activePlugin = eventstream
	case isEventhouseEndpoint(strings.ToLower(m.ConnectionString)):
		m.Log.Info("Detected EventHouse endpoint, using EventHouse output plugin")
		// Setting up the AzureDataExplorer plugin initial properties
		eventhouse := &eventhouse{}
		m.eventhouse = eventhouse
		if err := m.eventhouse.Init(); err != nil {
			return fmt.Errorf("initializing EventHouse output failed: %w", err)
		}
		eventhouse.Endpoint = m.ConnectionString
		eventhouse.log = m.Log
		eventhouse.Timeout = m.Timeout
		if err := eventhouse.parseconnectionString(m.ConnectionString); err != nil {
			return fmt.Errorf("parsing connection string failed: %w", err)
		}
		m.activePlugin = m.eventhouse
	default:
		return errors.New("invalid connection string")
	}
	return nil
}

func (m *MicrosoftFabric) Close() error {
	return m.activePlugin.Close()
}

func (m *MicrosoftFabric) Connect() error {
	return m.activePlugin.Connect()
}

func (m *MicrosoftFabric) Write(metrics []telegraf.Metric) error {
	return m.activePlugin.Write(metrics)
}

func init() {
	outputs.Add("microsoft_fabric", func() telegraf.Output {
		return &MicrosoftFabric{
			Timeout: config.Duration(30 * time.Second),
		}
	})
}
