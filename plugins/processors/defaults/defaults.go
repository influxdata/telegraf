package defaults

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
	## Ensure's a set of fields always exist on your metric(s) with the respective default value.
	## For any given field pair (key = default) under *[processors.defaults.fields]*, if it's not set, 
	## a new or updated field is set on the metric with the specified default.
	## 
	## A field is considered not set if it is nil on the incoming metric; or it is not nil but its value is an empty string or
    ## a string of one or more spaces.
	#
	#	[processors.defaults.fields]
	#	 field_1 = "bar"
	#    time_idle = 0
	#    is_error = true
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
	// return "Sets specified fields to a default value if they are nil or empty string or a single space character."
	return "Defaults sets default value(s) for specified fields that are nil, empty or a single space character on incoming metrics."
}

// Apply contains the main implementation of this processor.
// For each metric in 'inputMetrics', it goes over each default pair.
// If the field in the pair does not exist on the metric, the associated default is added.
// If the field was found, then, if its value is the empty string or a single space, it is replaced
// by the associated default.
func (def *Defaults) Apply(inputMetrics ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range inputMetrics {
		for defField, defValue := range def.DefaultFieldsSets {
			if maybeCurrent, isSet := metric.GetField(defField); !isSet {
				metric.AddField(defField, defValue)
			} else if maybeCurrent == "" || maybeCurrent == " " {
				metric.RemoveField(defField)
				metric.AddField(defField, defValue)
			}
		}
	}
	return inputMetrics
}

func init() {
	processors.Add("defaults", func() telegraf.Processor {
		return &Defaults{}
	})
}
