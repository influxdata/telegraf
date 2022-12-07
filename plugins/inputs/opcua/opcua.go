//go:generate ../../../tools/readme_config_includer/generator
package opcua

import (
	_ "embed"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/opcua"
	"github.com/influxdata/telegraf/plugins/common/opcua/input"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type OpcUA struct {
	ReadClientConfig
	Log telegraf.Logger `toml:"-"`

	client *ReadClient
}

func (*OpcUA) SampleConfig() string {
	return sampleConfig
}

// Init Initialise all required objects
func (o *OpcUA) Init() (err error) {
	o.client, err = o.ReadClientConfig.CreateReadClient(o.Log)
	return err
}

// Gather defines what data the plugin will gather.
func (o *OpcUA) Gather(acc telegraf.Accumulator) error {
	metrics, err := o.client.CurrentValues()
	if err != nil {
		return err
	}

	// Parse the resulting data into metrics
	for _, m := range metrics {
		acc.AddMetric(m)
	}
	return nil
}

// Add this plugin to telegraf
func init() {
	inputs.Add("opcua", func() telegraf.Input {
		return &OpcUA{
			ReadClientConfig: ReadClientConfig{
				InputClientConfig: input.InputClientConfig{
					OpcUAClientConfig: opcua.OpcUAClientConfig{
						Endpoint:       "opc.tcp://localhost:4840",
						SecurityPolicy: "auto",
						SecurityMode:   "auto",
						Certificate:    "/etc/telegraf/cert.pem",
						PrivateKey:     "/etc/telegraf/key.pem",
						AuthMethod:     "Anonymous",
						ConnectTimeout: config.Duration(5 * time.Second),
						RequestTimeout: config.Duration(10 * time.Second),
					},
					MetricName: "opcua",
					Timestamp:  input.TimestampSourceTelegraf,
				},
			},
		}
	})
}
