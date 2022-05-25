package defaults

import (
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

// Defaults is a processor for ensuring certain fields always exist
// on your Metrics with at least a default value.
type Defaults struct {
	DefaultFieldsSets map[string]interface{} `toml:"fields"`
}

// Apply contains the main implementation of this processor.
// For each metric in 'inputMetrics', it goes over each default pair.
// If the field in the pair does not exist on the metric, the associated default is added.
// If the field was found, then, if its value is the empty string or one or more spaces, it is replaced
// by the associated default.
func (def *Defaults) Apply(inputMetrics ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range inputMetrics {
		for defField, defValue := range def.DefaultFieldsSets {
			if maybeCurrent, isSet := metric.GetField(defField); !isSet {
				metric.AddField(defField, defValue)
			} else if trimmed, isStr := maybeTrimmedString(maybeCurrent); isStr && trimmed == "" {
				metric.RemoveField(defField)
				metric.AddField(defField, defValue)
			}
		}
	}
	return inputMetrics
}

func maybeTrimmedString(v interface{}) (string, bool) {
	if value, ok := v.(string); ok {
		return strings.TrimSpace(value), true
	}

	return "", false
}

func init() {
	processors.Add("defaults", func() telegraf.Processor {
		return &Defaults{}
	})
}
