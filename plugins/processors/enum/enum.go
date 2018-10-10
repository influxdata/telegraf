package enum

import (
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
  [[processors.enum.mapping]]
    ## Name of the field to map
    field = "status"

    ## Destination field to be used for the mapped value.  By default the source
    ## field is used, overwriting the original value.
    # dest = "status_code"

    ## Default value to be used for all values not contained in the mapping
    ## table.  When unset, the unmodified value for the field will be used if no
    ## match is found.
    # default = 0

    ## Table of mappings
    [processors.enum.mapping.value_mappings]
      green = 1
      yellow = 2
      red = 3
`

type EnumMapper struct {
	Mappings []Mapping `toml:"mapping"`
}

type Mapping struct {
	Field         string
	Dest          string
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
	for _, mapping := range mapper.Mappings {
		if originalValue, isPresent := metric.GetField(mapping.Field); isPresent == true {
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
	if mapping.Dest != "" {
		return mapping.Dest
	}
	return mapping.Field
}

func writeField(metric telegraf.Metric, name string, value interface{}) {
	metric.RemoveField(name)
	metric.AddField(name, value)
}

func init() {
	processors.Add("enum", func() telegraf.Processor {
		return &EnumMapper{}
	})
}
