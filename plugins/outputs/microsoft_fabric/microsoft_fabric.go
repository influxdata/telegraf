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
	"github.com/influxdata/telegraf/plugins/common/adx"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type fabric interface {
	Connect() error
	Write(metrics []telegraf.Metric) error
	Close() error
}

type MicrosoftFabric struct {
	ConnectionString string          `toml:"connection_string"`
	Timeout          config.Duration `toml:"timeout"`
	Log              telegraf.Logger `toml:"-"`

	output fabric
}

func (*MicrosoftFabric) SampleConfig() string {
	return sampleConfig
}

func (m *MicrosoftFabric) Init() error {
	// Check input parameters
	if m.ConnectionString == "" {
		return errors.New("endpoint must not be empty")
	}

	// Initialize the output fabric dependent on the type
	switch {
	case isEventstreamEndpoint(m.ConnectionString):
		m.Log.Debug("Detected EventStream endpoint...")
		eventstream := &eventstream{
			connectionString: m.ConnectionString,
			timeout:          m.Timeout,
			log:              m.Log,
		}
		if err := eventstream.init(); err != nil {
			return fmt.Errorf("initializing EventStream output failed: %w", err)
		}
		m.output = eventstream
	case isEventhouseEndpoint(strings.ToLower(m.ConnectionString)):
		m.Log.Debug("Detected EventHouse endpoint...")
		eventhouse := &eventhouse{
			connectionString: m.ConnectionString,
			Config: adx.Config{
				Timeout: m.Timeout,
			},
			log: m.Log,
		}
		if err := eventhouse.init(); err != nil {
			return fmt.Errorf("initializing EventHouse output failed: %w", err)
		}
		m.output = eventhouse
	default:
		return errors.New("invalid connection string: unable to detect endpoint type (EventStream or EventHouse)")
	}
	return nil
}

func (m *MicrosoftFabric) Close() error {
	return m.output.Close()
}

func (m *MicrosoftFabric) Connect() error {
	return m.output.Connect()
}

func (m *MicrosoftFabric) Write(metrics []telegraf.Metric) error {
	return m.output.Write(metrics)
}

func init() {
	outputs.Add("microsoft_fabric", func() telegraf.Output {
		return &MicrosoftFabric{
			Timeout: config.Duration(30 * time.Second),
		}
	})
}
