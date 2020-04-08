package enum

import (
	"fmt"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
  [[processors.enum.mapping]]
    ## Name of the field(s) to map
    fields = ["status"]

    ## Name of the tag(s) to map
    # tags = ["status"]

    ## Destination tag or field to be used for the mapped value.  By default the
    ## source tag or field is used, overwriting the original value.
    dest = "status_code"

    ## Default value to be used for all values not contained in the mapping
    ## table.  When unset, the unmodified value for the field will be used if no
    ## match is found.
    # default = 0

    ## Table of mappings
    [processors.enum.mapping.value_mappings]
      green = 1
      amber = 2
      red = 3
`

type EnumMapper struct {
	Mappings []Mapping `toml:"mapping"`
}

type Mapping struct {
	Tags          []string
	Fields        []string
	Dest          []string
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
		if len(mapping.Fields) != 0 {
			for i, m_field := range mapping.Fields {
				if originalValue, isPresent := metric.GetField(m_field); isPresent {
					if adjustedValue, isString := adjustBoolValue(originalValue).(string); isString {
						if mappedValue, isMappedValuePresent := mapping.mapValue(adjustedValue); isMappedValuePresent {
							writeField(metric, mapping.getDestination(i), mappedValue)
						}
					}
				}
			}
		}
		if len(mapping.Tags) != 0 {
			for i, m_tag := range mapping.Tags {
				if originalValue, isPresent := metric.GetTag(m_tag); isPresent {
					if mappedValue, isMappedValuePresent := mapping.mapValue(originalValue); isMappedValuePresent {
						switch val := mappedValue.(type) {
						case string:
							writeTag(metric, mapping.getDestinationTag(i), val)
						default:
							writeTag(metric, mapping.getDestinationTag(i), fmt.Sprintf("%v", val))
						}
					}
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

func (mapping *Mapping) getDestination(index int) string {
	if mappedDest, found := mapping.Dest[index]; found == true {
		return mapping.Dest[index]
	}
	return mapping.Fields[index]
}

func (mapping *Mapping) getDestinationTag(index int) string {
	if mappedDest, found := mapping.Dest[index]; found == true {
		return mapping.Dest[index]
	}
	return mapping.Tags[index]
}

func writeField(metric telegraf.Metric, name string, value interface{}) {
	metric.RemoveField(name)
	metric.AddField(name, value)
}

func writeTag(metric telegraf.Metric, name string, value string) {
	metric.RemoveTag(name)
	metric.AddTag(name, value)
}

func init() {
	processors.Add("enum", func() telegraf.Processor {
		return &EnumMapper{}
	})
}
