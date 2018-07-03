package enum

import (
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
## NOTE This processor will map metric values to different values. It is aimed
## to map enum values to numeric values.

## Fields to be considered
# [[processors.enum.fields]]
#
# Name of the field source field to map
#   source = "name"
#
# Optional destination field to be used for the mapped value. Source field is
# used, when no explicit destination is configured.
#   destination = "mapped"
#
# Optional default value to be used for all values not contained in the mapping
# table. Only applied when configured.
#   default = 0
#
# Value Mapping Table
#   [processors.enum.value_mappings]
#     value1 = 1
#     value2 = 2
#
## Alternatively the mapping table can be given in inline notation
#   value_mappings = {value1 = 1, value2 = 2}
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
