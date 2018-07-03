package enum

import (
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
  [[processors.enum.fields]]
    ## Name of the field to map
    source = "name"

    ## Destination field to be used for the mapped value.  By default the source
    ## field is used, overwriting the original value.
    # destination = "mapped"

    ## Default value to be used for all values not contained in the mapping
    ## table.  When unset, the unmodified value for the field will be used if no
    ## match is found.
    # default = 0

    ## Table of mappings
    [processors.enum.fields.value_mappings]
      value1 = 1
      value2 = 2
`

type EnumMapper struct {
	Fields []Mapping
}

type Mapping struct {
	Source        string
	Destination   string
	Default       interface{}
	ValueMappings map[string]interface{}
}

func (mapper *EnumMapper) SampleConfig() string {
	return sampleConfig
}

func (mapper *EnumMapper) Description() string {
	return "Map enum values according to given table."
}

func (mapper *EnumMapper) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for i := 0; i < len(in); i++ {
		in[i] = mapper.applyMappings(in[i])
	}
	return in
}

func (mapper *EnumMapper) applyMappings(metric telegraf.Metric) telegraf.Metric {
	for _, mapping := range mapper.Fields {
		if originalValue, isPresent := metric.GetField(mapping.Source); isPresent == true {
			if adjustedValue, isString := adjustBoolValue(originalValue).(string); isString == true {
				if mappedValue, isMappedValuePresent := mapping.mapValue(adjustedValue); isMappedValuePresent == true {
					writeField(metric, mapping.getDestination(), mappedValue)
				}
			}
		}
	}
	return metric
}

func adjustBoolValue(in interface{}) interface{} {
	if mappedBool, isBool := in.(bool); isBool == true {
		return strconv.FormatBool(mappedBool)
	}
	return in
}

func (mapping *Mapping) mapValue(original string) (interface{}, bool) {
	if mapped, found := mapping.ValueMappings[original]; found == true {
		return mapped, true
	}
	if mapping.Default != nil {
		return mapping.Default, true
	}
	return original, false
}

func (mapping *Mapping) getDestination() string {
	if mapping.Destination != "" {
		return mapping.Destination
	}
	return mapping.Source
}

func writeField(metric telegraf.Metric, name string, value interface{}) {
	if metric.HasField(name) {
		metric.RemoveField(name)
	}
	metric.AddField(name, value)
}

func init() {
	processors.Add("enum", func() telegraf.Processor {
		return &EnumMapper{}
	})
}
