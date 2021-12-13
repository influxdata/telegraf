package defaults

import (
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
  ## Ensures a set of fields always exists on your metric(s) with their 
  ## respective default value.
  ## For any given field pair (key = default), if it's not set, a field 
  ## is set on the metric with the specified default.
  ## 
  ## A field is considered not set if it is nil on the incoming metric;
  ## or it is not nil but its value is an empty string or is a string 
  ## of one or more spaces.
  ##   <target-field> = <value>
  # [processors.defaults.fields]
  #   field_1 = "bar"
  #   time_idle = 0
  #   is_error = true
`

// Defaults is a processor for ensuring certain fields always exist
// on your Metrics with at least a default value.
type Defaults struct {
	DefaultFieldsSets map[string]interface{} `toml:"fields"`
}

// SampleConfig represents a sample toml config for this plugin.
func (def *Defaults) SampleConfig() string {
	return sampleConfig
}

// Description is a brief description of this processor plugin's behaviour.
func (def *Defaults) Description() string {
	return "Defaults sets default value(s) for specified fields that are not set on incoming metrics."
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
