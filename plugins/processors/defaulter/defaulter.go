package defaulter

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
)

const sampleConfig = `
	[[processors.defaulter.values]]
		fields = ["field_1", "field_2", "field_3"]
		value = "NONE"


	[[processors.defaulter.values]]
		field = ["field_4", "field_5"]
		value = "TEST"

	## If the same field shows up in multiple of these value objects,
	then the last one will win out.
`

type DefaultFieldsSet struct {
	Fields []string `toml:"fields"`
	Value string    `toml:"value"`

	filter filter.Filter
}

type Defaulter struct {
	DefaultFieldsSets []DefaultFieldsSet `toml:"values"`

	defaultValueCache map[string]string
	// May be worth it later to add a non match cache. That is, fields for which no filter matches
	// Should be placed in a cache so we don't look over the list for them
}

func (def *Defaulter) SampleConfig() string {
	return sampleConfig
}

func (def *Defaulter) Description() string {
	return "Set the selected fields to a specified default value if they are nil or empty or zero"
}

func (def *Defaulter) Init() error {
	for _, fieldsSet := range def.DefaultFieldsSets {
		f, err := filter.Compile(fieldsSet.Fields)
		if err != nil {
			return err
		}
		fieldsSet.filter = f
	}

	return nil
}

func (def *Defaulter) Apply(inputMetrics ...telegraf.Metric) []telegraf.Metric {
	if def.defaultValueCache == nil {
		def.defaultValueCache = make(map[string]string)
	}

	for _, metric := range inputMetrics {
		for fieldName, fieldValue := range metric.Fields() {
			if fieldValue == nil || fieldValue == "" || fieldValue == 0 || fieldValue == ' ' {
				if cachedVal := def.cached(fieldName); cachedVal != nil {
					metric.RemoveField(fieldName)
					metric.AddField(fieldName, &cachedVal)
					continue
				}

				var foundValue *string
				for _, set := range def.DefaultFieldsSets {
					if set.filter.Match(fieldName) {
						foundValue = &set.Value
						break
					}
				}

				if foundValue != nil {
					def.defaultValueCache[fieldName] = *foundValue
					metric.RemoveField(fieldName)
					metric.AddField(fieldName, *foundValue)
				}
			}
		}
	}

	return inputMetrics
}

func (def *Defaulter) cached(key string) *string {
	if val, ok := def.defaultValueCache[key]; ok {
		return &val
	}
	return nil
}
