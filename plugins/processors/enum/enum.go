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
				if mappedValue, isMappedValuePresent := mapping.ValueMappings[adjustedValue]; isMappedValuePresent == true {
					metric.RemoveField(mapping.Source)
					metric.AddField(mapping.Source, mappedValue)
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

func init() {
	processors.Add("enum", func() telegraf.Processor {
		return &EnumMapper{}
	})
}
