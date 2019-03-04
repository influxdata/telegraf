package converter

import (
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
  ## Tags to convert
  ##
  ## The table key determines the target type, and the array of key-values
  ## select the keys to convert.  The array may contain globs.
  ##   <target-type> = [<tag-key>...]
  [processors.converter.tags]
    string = []
    integer = []
    unsigned = []
    boolean = []
    float = []

  ## Fields to convert
  ##
  ## The table key determines the target type, and the array of key-values
  ## select the keys to convert.  The array may contain globs.
  ##   <target-type> = [<field-key>...]
  [processors.converter.fields]
    tag = []
    string = []
    integer = []
    unsigned = []
    boolean = []
    float = []
`

type Conversion struct {
	Tag      []string `toml:"tag"`
	String   []string `toml:"string"`
	Integer  []string `toml:"integer"`
	Unsigned []string `toml:"unsigned"`
	Boolean  []string `toml:"boolean"`
	Float    []string `toml:"float"`
}

type Converter struct {
	Tags   *Conversion `toml:"tags"`
	Fields *Conversion `toml:"fields"`

	initialized      bool
	tagConversions   *ConversionFilter
	fieldConversions *ConversionFilter
}

type ConversionFilter struct {
	Tag      filter.Filter
	String   filter.Filter
	Integer  filter.Filter
	Unsigned filter.Filter
	Boolean  filter.Filter
	Float    filter.Filter
}

func (p *Converter) SampleConfig() string {
	return sampleConfig
}

func (p *Converter) Description() string {
	return "Convert values to another metric value type"
}

func (p *Converter) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	if !p.initialized {
		err := p.compile()
		if err != nil {
			logPrintf("initialization error: %v\n", err)
			return metrics
		}
	}

	for _, metric := range metrics {
		p.convertTags(metric)
		p.convertFields(metric)
	}
	return metrics
}

func (p *Converter) compile() error {
	tf, err := compileFilter(p.Tags)
	if err != nil {
		return err
	}

	ff, err := compileFilter(p.Fields)
	if err != nil {
		return err
	}

	if tf == nil && ff == nil {
		return fmt.Errorf("no filters found")
	}

	p.tagConversions = tf
	p.fieldConversions = ff
	p.initialized = true
	return nil
}

func compileFilter(conv *Conversion) (*ConversionFilter, error) {
	if conv == nil {
		return nil, nil
	}

	var err error
	cf := &ConversionFilter{}
	cf.Tag, err = filter.Compile(conv.Tag)
	if err != nil {
		return nil, err
	}

	cf.String, err = filter.Compile(conv.String)
	if err != nil {
		return nil, err
	}

	cf.Integer, err = filter.Compile(conv.Integer)
	if err != nil {
		return nil, err
	}

	cf.Unsigned, err = filter.Compile(conv.Unsigned)
	if err != nil {
		return nil, err
	}

	cf.Boolean, err = filter.Compile(conv.Boolean)
	if err != nil {
		return nil, err
	}

	cf.Float, err = filter.Compile(conv.Float)
	if err != nil {
		return nil, err
	}

	return cf, nil
}

// convertTags converts tags into fields
func (p *Converter) convertTags(metric telegraf.Metric) {
	if p.tagConversions == nil {
		return
	}

	for key, value := range metric.Tags() {
		if p.tagConversions.String != nil && p.tagConversions.String.Match(key) {
			metric.RemoveTag(key)
			metric.AddField(key, value)
			continue
		}

		if p.tagConversions.Integer != nil && p.tagConversions.Integer.Match(key) {
			v, ok := toInteger(value)
			if !ok {
				metric.RemoveTag(key)
				logPrintf("error converting to integer [%T]: %v\n", value, value)
				continue
			}

			metric.RemoveTag(key)
			metric.AddField(key, v)
		}

		if p.tagConversions.Unsigned != nil && p.tagConversions.Unsigned.Match(key) {
			v, ok := toUnsigned(value)
			if !ok {
				metric.RemoveTag(key)
				logPrintf("error converting to unsigned [%T]: %v\n", value, value)
				continue
			}

			metric.RemoveTag(key)
			metric.AddField(key, v)
			continue
		}

		if p.tagConversions.Boolean != nil && p.tagConversions.Boolean.Match(key) {
			v, ok := toBool(value)
			if !ok {
				metric.RemoveTag(key)
				logPrintf("error converting to boolean [%T]: %v\n", value, value)
				continue
			}

			metric.RemoveTag(key)
			metric.AddField(key, v)
			continue
		}

		if p.tagConversions.Float != nil && p.tagConversions.Float.Match(key) {
			v, ok := toFloat(value)
			if !ok {
				metric.RemoveTag(key)
				logPrintf("error converting to float [%T]: %v\n", value, value)
				continue
			}

			metric.RemoveTag(key)
			metric.AddField(key, v)
			continue
		}
	}
}

// convertFields converts fields into tags or other field types
func (p *Converter) convertFields(metric telegraf.Metric) {
	if p.fieldConversions == nil {
		return
	}

	for key, value := range metric.Fields() {
		if p.fieldConversions.Tag != nil && p.fieldConversions.Tag.Match(key) {
			v, ok := toString(value)
			if !ok {
				metric.RemoveField(key)
				logPrintf("error converting to tag [%T]: %v\n", value, value)
				continue
			}

			metric.RemoveField(key)
			metric.AddTag(key, v)
			continue
		}

		if p.fieldConversions.Float != nil && p.fieldConversions.Float.Match(key) {
			v, ok := toFloat(value)
			if !ok {
				metric.RemoveField(key)
				logPrintf("error converting to integer [%T]: %v\n", value, value)
				continue
			}

			metric.RemoveField(key)
			metric.AddField(key, v)
			continue
		}

		if p.fieldConversions.Integer != nil && p.fieldConversions.Integer.Match(key) {
			v, ok := toInteger(value)
			if !ok {
				metric.RemoveField(key)
				logPrintf("error converting to integer [%T]: %v\n", value, value)
				continue
			}

			metric.RemoveField(key)
			metric.AddField(key, v)
			continue
		}

		if p.fieldConversions.Unsigned != nil && p.fieldConversions.Unsigned.Match(key) {
			v, ok := toUnsigned(value)
			if !ok {
				metric.RemoveField(key)
				logPrintf("error converting to unsigned [%T]: %v\n", value, value)
				continue
			}

			metric.RemoveField(key)
			metric.AddField(key, v)
			continue
		}

		if p.fieldConversions.Boolean != nil && p.fieldConversions.Boolean.Match(key) {
			v, ok := toBool(value)
			if !ok {
				metric.RemoveField(key)
				logPrintf("error converting to bool [%T]: %v\n", value, value)
				continue
			}

			metric.RemoveField(key)
			metric.AddField(key, v)
			continue
		}

		if p.fieldConversions.String != nil && p.fieldConversions.String.Match(key) {
			v, ok := toString(value)
			if !ok {
				metric.RemoveField(key)
				logPrintf("error converting to string [%T]: %v\n", value, value)
				continue
			}

			metric.RemoveField(key)
			metric.AddField(key, v)
			continue
		}
	}
}

func toBool(v interface{}) (bool, bool) {
	switch value := v.(type) {
	case int64, uint64, float64:
		if value != 0 {
			return true, true
		} else {
			return false, false
		}
	case bool:
		return value, true
	case string:
		result, err := strconv.ParseBool(value)
		return result, err == nil
	}
	return false, false
}

func toInteger(v interface{}) (int64, bool) {
	switch value := v.(type) {
	case int64:
		return value, true
	case uint64:
		if value <= uint64(math.MaxInt64) {
			return int64(value), true
		} else {
			return math.MaxInt64, true
		}
	case float64:
		if value < float64(math.MinInt64) {
			return math.MinInt64, true
		} else if value > float64(math.MaxInt64) {
			return math.MaxInt64, true
		} else {
			return int64(Round(value)), true
		}
	case bool:
		if value {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		result, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, false
		}
		return toInteger(result)
	}
	return 0, false
}

func toUnsigned(v interface{}) (uint64, bool) {
	switch value := v.(type) {
	case uint64:
		return value, true
	case int64:
		if value < 0 {
			return 0, true
		} else {
			return uint64(value), true
		}
	case float64:
		if value < 0.0 {
			return 0, true
		} else if value > float64(math.MaxUint64) {
			return math.MaxUint64, true
		} else {
			return uint64(Round(value)), true
		}
	case bool:
		if value {
			return 1, true
		} else {
			return 0, true
		}
	case string:
		result, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, false
		}
		return toUnsigned(result)
	}
	return 0, false
}

func toFloat(v interface{}) (float64, bool) {
	switch value := v.(type) {
	case int64:
		return float64(value), true
	case uint64:
		return float64(value), true
	case float64:
		return value, true
	case bool:
		if value {
			return 1.0, true
		} else {
			return 0.0, true
		}
	case string:
		result, err := strconv.ParseFloat(value, 64)
		return result, err == nil
	}
	return 0.0, false
}

func toString(v interface{}) (string, bool) {
	switch value := v.(type) {
	case int64:
		return strconv.FormatInt(value, 10), true
	case uint64:
		return strconv.FormatUint(value, 10), true
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64), true
	case bool:
		return strconv.FormatBool(value), true
	case string:
		return value, true
	}
	return "", false
}

// math.Round was not added until Go 1.10, can be removed when support for Go
// 1.9 is dropped.
func Round(x float64) float64 {
	t := math.Trunc(x)
	if math.Abs(x-t) >= 0.5 {
		return t + math.Copysign(1, x)
	}
	return t
}

func logPrintf(format string, v ...interface{}) {
	log.Printf("D! [processors.converter] "+format, v...)
}

func init() {
	processors.Add("converter", func() telegraf.Processor {
		return &Converter{}
	})
}
