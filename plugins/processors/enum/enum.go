package enum

import (
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
## NOTE This processor will map metric values to different values. It is aimed
## to map enum values to numeric values.

## Fields to be considered
# [[processors.enum.fields]]
#
# Name of the field
#   key = "name"
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
	Key           string
	ValueMappings map[string]interface{}
}

func (mapper *EnumMapper) SampleConfig() string {
	return sampleConfig
}

func (mapper *EnumMapper) Description() string {
	return "Map enum values according to given table."
}

func (mapper *EnumMapper) Apply(in ...telegraf.Metric) []telegraf.Metric {
	out := make([]telegraf.Metric, 0, len(in))
	for _, source := range in {
		target, error := mapper.applyMappings(source)
		if error == nil {
			out = append(out, target)
		} else {
			out = append(out, source)
		}
	}
	return out
}

func (mapper *EnumMapper) applyMappings(source telegraf.Metric) (telegraf.Metric, error) {
	if fields, changed := mapper.applyFieldMappings(source.Fields()); changed == true {
		return metric.New(
			source.Name(),
			source.Tags(),
			fields,
			source.Time(),
			source.Type())
	} else {
		return source, nil
	}
}

func (mapper *EnumMapper) applyFieldMappings(in map[string]interface{}) (map[string]interface{}, bool) {
	out := make(map[string]interface{}, len(in))
	changed := false
	for key, value := range in {
		var isMapped bool
		out[key], isMapped = mapper.determineMappedValue(key, value)
		changed = changed || isMapped
	}
	return out, changed
}

func (mapper *EnumMapper) determineMappedValue(key string, value interface{}) (interface{}, bool) {
	adjustedValue := adjustBoolValue(value)
	if _, isString := adjustedValue.(string); isString == false {
		return value, false
	}
	for _, mapping := range mapper.Fields {
		if mapping.Key == key {
			if mappedValue, isMappedValuePresent := mapping.ValueMappings[adjustedValue.(string)]; isMappedValuePresent == true {
				return mappedValue, true
			}
		}
	}
	return value, false
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
