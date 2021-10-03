package receive_parse

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/common/transport"

	"github.com/influxdata/telegraf/plugins/parsers"
)

// PostProcessor allows to perform arbitrary post-processing (e.g. type conversion or formatting) on the parsed metrics.
type PostProcessor struct {
	Name    string
	Process func(m telegraf.Metric) error
}

// ReceiveAndParse is a general plugin implementation for plugins that receive some data over a transport
// (e.g. exec, or http) and parse this data into metrics.
// Individual plugins then only need to specify the transport and to configure the parser.
type ReceiveAndParse struct {
	// Receiver will also contain all configurable parameters for the plugin
	transport.Receiver
	// Parser to process the raw data received by the transport
	Parser parsers.Parser `toml:"-"`
	// PostProcessors allow to post-process the parsed metrics in an arbitrary manner
	PostProcessors []PostProcessor `toml:"-"`

	// Description for the plugin
	DescriptionText string `toml:"-"`

	// Log is the logging facility automatically filled by telegraf
	Log telegraf.Logger `toml:"-"`
}

// Description returns the description of the GraphicsSMI plugin
func (r *ReceiveAndParse) Description() string {
	return r.DescriptionText
}

// SampleConfig returns the sample configuration for the GraphicsSMI plugin
func (r *ReceiveAndParse) SampleConfig() string {
	return r.Receiver.SampleConfig()
}

// Init implements the initializer interface
func (r *ReceiveAndParse) Init() error {
	// Try to push the logger to the receiver and parser
	models.SetLoggerOnPlugin(r.Receiver, r.Log)
	models.SetLoggerOnPlugin(r.Parser, r.Log)

	// Try to initialize the transport
	if t, ok := r.Receiver.(telegraf.Initializer); ok {
		if err := t.Init(); err != nil {
			return fmt.Errorf("initializing receiver failed: %v", err)
		}
	}

	// Try to initialize the parser
	if p, ok := r.Parser.(telegraf.Initializer); ok {
		if err := p.Init(); err != nil {
			return fmt.Errorf("initializing parser failed: %v", err)
		}
	}

	fmt.Printf("got: %v\n", r)

	return nil
}

// Gather implements the telegraf interface
func (r *ReceiveAndParse) Gather(acc telegraf.Accumulator) error {
	data, err := r.Receiver.Receive()
	if err != nil {
		return fmt.Errorf("receiving data failed: %v", err)
	}

	return r.Parse(acc, data)
}

func (r *ReceiveAndParse) Parse(acc telegraf.Accumulator, data []byte) error {
	metrics, err := r.Parser.Parse(data)
	if err != nil {
		return fmt.Errorf("parsing data failed: %v", err)
	}

	for _, metric := range metrics {
		for _, proc := range r.PostProcessors {
			r.Log.Debugf("Running post-processors %q on metric %q...", proc.Name, metric.Name())
			if err := proc.Process(metric); err != nil {
				acc.AddError(fmt.Errorf("post-processor %q failed: %v", proc.Name, err))
				continue
			}
		}
		acc.AddMetric(metric)
	}
	return nil
}
