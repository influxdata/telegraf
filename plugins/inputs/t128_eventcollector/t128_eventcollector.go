package t128_eventcollector

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type T128EventCollector struct {
	CollectorName string `toml:"collector_name"`
	LogName       string `toml:"log_name"`
	Topic         string `toml:"topic"`
	IndexFile     string `toml:"index_file"`
	EventType     string `toml:"event_type"`
}

var sampleConfig = `
## Collect data from a 128T instance using graphQL.
[[inputs.t128_eventcollector]]
## Required. A name for the collector which will be used as the measurement name of the produced data.
# collector_name = "session-records"
## Required. The name of the log file to produce.
# [inputs.t128_eventcollector.log_name]
# 	log-name = "event_collector"
## The TANK topic to consume.
# [inputs.t128_eventcollector.topic]
# 	topic = "events"
## Required. A (unique) file to use for index tracking. This tracking allows each
## event/session record to be produced once. By default, no tracking is used and
## event/session record are produced starting from the point telegraf is launched.
# [inputs.t128_eventcollector.index_file]
# 	index-file = "/var/lib/128t-monitoring/state/events.index"
## Event filtering based on type.
# [inputs.t128_eventcollector.event_type]
# 	event_type = "alarm"
`

// SampleConfig returns the default configuration of the Input
func (*T128EventCollector) SampleConfig() string {
	return sampleConfig
}

// Description returns a one-sentence description on the Input
func (*T128EventCollector) Description() string {
	return "Read 128T Event Collector"
}

// Init sets up the input to be ready for action
func (plugin *T128EventCollector) Init() error {
	//check and load config
	err := plugin.checkConfig()
	if err != nil {
		return err
	}

	return nil
}

func (plugin *T128EventCollector) checkConfig() error {
	if plugin.CollectorName == "" {
		return fmt.Errorf("collector_name is a required configuration field")
	}

	if plugin.LogName == "" {
		return fmt.Errorf("log_name is a required configuration field")
	}

	if plugin.IndexFile == "" {
		return fmt.Errorf("index_file is a required configuration field")
	}

	return nil
}

// Gather takes in an accumulator and adds the metrics that the Input gathers
func (plugin *T128EventCollector) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("t128_eventcollector", func() telegraf.Input {
		return &T128EventCollector{}
	})
}
