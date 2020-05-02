package defaulter

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
	## Ensure's a set of fields always exist on your metric(s) with at the specified default value.
	## For any given field pair (key = default) under *[processors.defaulter.fields]*, if it's not set, or its value
	## is 'empty' on the incoming metric, a new or updated metric is set on the metric with the specified default.
	## 
	#	[processors.defaulter.fields]
	#	 field_1 = "bar"
	#    time_idle = 0
	#    is_error = true
`

// Defaulter is a processor for ensuring certain fields always exist
// on your Metrics with at least a default value.
type Defaulter struct {
	DefaultFieldsSets map[string]interface{} `toml:"fields"`
	Log               telegraf.Logger        `toml:"-"`
}

// SampleConfig represents a sample toml config for this plugin.
func (def *Defaulter) SampleConfig() string {
	return sampleConfig
}

// Description is a brief description of this processor plugin's behaviour.
func (def *Defaulter) Description() string {
	// return "Sets specified fields to a default value if they are nil or empty string or a single space character."
	return "Defaulter sets default value(s) for specified fields that are nil, empty or a single space character on incoming metrics."
}

// Apply contains the main implementation of this processor.
// For each metric in 'inputMetrics', it goes over each default pair.
// If the field in the pair does not exist on the metric, the associated default is added.
// If the field was found, then, if its value is the empty string or a single space, it is replaced
// by the associated default.
func (def *Defaulter) Apply(inputMetrics ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range inputMetrics {
		for defField, defValue := range def.DefaultFieldsSets {
			if maybeCurrent, isSet := metric.GetField(defField); !isSet {
				def.Log.Debugf("Field with name: %v, was not set on metric: %v.", defField, metric.Name())
				metric.AddField(defField, defValue)
			} else if maybeCurrent == "" || maybeCurrent == ' ' {
				def.Log.Debugf("Field with name: %v was set, but the value (%v) is considered empty. Setting new value to %v", defField, maybeCurrent, defValue)
				metric.RemoveField(defField)
				metric.AddField(defField, defValue)
			}
		}
	}
	return inputMetrics
}

func init() {
	processors.Add("defaulter", func() telegraf.Processor {
		return &Defaulter{}
	})
}
