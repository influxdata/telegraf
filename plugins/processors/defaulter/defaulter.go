package defaulter

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
	[[processors.defaulter.values]]
		fields = ["field_1", "field_2", "field_3"]
		value = "NONE"
		metric_name = "CPU"


	[[processors.defaulter.values]]
		field = ["field_4", "field_5"]
		value = "TEST"
		metric_name = "Disk"

	# If the same field shows up in multiple of these value objects,
	#then the last one will win out.
`

type DefaultFieldsSet struct {
	Fields []string `toml:"fields"`
	Metric string   `toml:"metric_name"`
	Value  string   `toml:"value"`
}

type Defaulter struct {
	DefaultFieldsSets []DefaultFieldsSet `toml:"values"`
	Log    telegraf.Logger `toml:"-"`
}

func (def *Defaulter) SampleConfig() string {
	return sampleConfig
}

func (def *Defaulter) Description() string {
	return "Set the selected fields to a specified default value if they are nil or empty or zero"
}

func (def *Defaulter) Init() error {
	return nil
}

func (def *Defaulter) Apply(inputMetrics ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range inputMetrics {
		for _, defSet := range def.DefaultFieldsSets {
			def.Log.Debugf("Going over the fields of a metric with name: %s", metric.Name())
			if defSet.Metric != "" && metric.Name() != defSet.Metric {
				continue
			}
			for _, field := range defSet.Fields {
				maybeCurrent, isSet := metric.GetField(field)
				if !isSet {
					def.Log.Debugf("Field with name: %v was not set.", field)
					metric.AddField(field, defSet.Value)
					continue
				}

				if maybeCurrent == "" || maybeCurrent == ' ' || maybeCurrent == 0 || maybeCurrent == int64(0) || maybeCurrent == "0" {
					def.Log.Debugf("Field with name: %v was set but value was an empty: %v. Setting new value to %v", field, maybeCurrent, defSet.Value)
					metric.RemoveField(field)
					metric.AddField(field, defSet.Value)
				}
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
