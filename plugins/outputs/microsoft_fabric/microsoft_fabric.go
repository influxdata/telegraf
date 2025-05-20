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
	Log              telegraf.Logger `toml:"-"`
	Timeout          config.Duration `toml:"timeout"`

	eventhouse   *eventhouse
	eventstream  *eventstream
	activePlugin fabricOutput
}

func (*MicrosoftFabric) SampleConfig() string {
	return sampleConfig
}

func (m *MicrosoftFabric) Init() error {
	connectionString := m.ConnectionString

	if connectionString == "" {
		return errors.New("endpoint must not be empty")
	}

	if strings.HasPrefix(connectionString, "Endpoint=sb") {
		m.Log.Info("Detected EventStream endpoint, using EventStream output plugin")
		eventstream := &eventstream{}
		eventstream.connectionString = connectionString
		eventstream.log = m.Log
		eventstream.timeout = m.Timeout
		if err := eventstream.parseconnectionString(connectionString); err != nil {
			return fmt.Errorf("parsing connection string failed: %w", err)
		}
		m.eventstream = eventstream
		if err := m.eventstream.Init(); err != nil {
			return fmt.Errorf("initializing EventStream output failed: %w", err)
		}
		m.activePlugin = eventstream
	} else if isKustoEndpoint(strings.ToLower(connectionString)) {
		m.Log.Info("Detected EventHouse endpoint, using EventHouse output plugin")
		// Setting up the AzureDataExplorer plugin initial properties
		eventhouse := &eventhouse{}
		m.eventhouse = eventhouse
		if err := m.eventhouse.Init(); err != nil {
			return fmt.Errorf("initializing EventHouse output failed: %w", err)
		}
		eventhouse.config.Endpoint = connectionString
		eventhouse.log = m.Log
		eventhouse.config.Timeout = m.Timeout
		if err := eventhouse.parseconnectionString(connectionString); err != nil {
			return fmt.Errorf("parsing connection string failed: %w", err)
		}
		m.activePlugin = m.eventhouse
	} else {
		return errors.New("invalid connection string")
	}
	return nil
}

func (m *MicrosoftFabric) Close() error {
	if m.activePlugin == nil {
		return errors.New("no active plugin to close")
	}
	return m.activePlugin.Close()
}

func (m *MicrosoftFabric) Connect() error {
	if m.activePlugin == nil {
		return errors.New("no active plugin to connect")
	}
	return m.activePlugin.Connect()
}

func (m *MicrosoftFabric) Write(metrics []telegraf.Metric) error {
	if m.activePlugin == nil {
		return errors.New("no active plugin to write to")
	}
	return m.activePlugin.Write(metrics)
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
			Timeout: config.Duration(30 * time.Second),
		}
	})
}
