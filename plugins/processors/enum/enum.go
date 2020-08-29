package enum

import (
	"fmt"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
  [[processors.enum.mapping]]
    ## Name of the field to map
    field = "status"

    ## Name of the tag to map
    # tag = "status"

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
	Tag           string
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
		if mapping.Field != "" {
			if originalValue, isPresent := metric.GetField(mapping.Field); isPresent {
				if adjustedValue, isString := adjustValue(originalValue).(string); isString {
					if mappedValue, isMappedValuePresent := mapping.mapValue(adjustedValue); isMappedValuePresent {
						writeField(metric, mapping.getDestination(), mappedValue)
					}
				}
			}
		}
		if mapping.Tag != "" {
			if originalValue, isPresent := metric.GetTag(mapping.Tag); isPresent {
				if mappedValue, isMappedValuePresent := mapping.mapValue(originalValue); isMappedValuePresent {
					switch val := mappedValue.(type) {
					case string:
						writeTag(metric, mapping.getDestinationTag(), val)
					default:
						writeTag(metric, mapping.getDestinationTag(), fmt.Sprintf("%v", val))
					}
				}
			}
		}
	}
	return metric
}

func adjustValue(in interface{}) interface{} {
	switch val := in.(type) {
	case bool:
		return strconv.FormatBool(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	default:
		return in
	}
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

func (mapping *Mapping) getDestinationTag() string {
	if mapping.Dest != "" {
		return mapping.Dest
	}
	return mapping.Tag
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
