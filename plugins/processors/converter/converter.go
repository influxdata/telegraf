package converter

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Conversion struct {
	Measurement []string `toml:"measurement"`
	Tag         []string `toml:"tag"`
	String      []string `toml:"string"`
	Integer     []string `toml:"integer"`
	Unsigned    []string `toml:"unsigned"`
	Boolean     []string `toml:"boolean"`
	Float       []string `toml:"float"`
}

type Converter struct {
	Tags   *Conversion     `toml:"tags"`
	Fields *Conversion     `toml:"fields"`
	Log    telegraf.Logger `toml:"-"`

	tagConversions   *ConversionFilter
	fieldConversions *ConversionFilter
}

type ConversionFilter struct {
	Measurement filter.Filter
	Tag         filter.Filter
	String      filter.Filter
	Integer     filter.Filter
	Unsigned    filter.Filter
	Boolean     filter.Filter
	Float       filter.Filter
}

func (p *Converter) Init() error {
	return p.compile()
}

func (p *Converter) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
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
	return nil
}

func compileFilter(conv *Conversion) (*ConversionFilter, error) {
	if conv == nil {
		return nil, nil
	}

	var err error
	cf := &ConversionFilter{}
	cf.Measurement, err = filter.Compile(conv.Measurement)
	if err != nil {
		return nil, err
	}

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

// convertTags converts tags into measurements or fields.
func (p *Converter) convertTags(metric telegraf.Metric) {
	if p.tagConversions == nil {
		return
	}

	for key, value := range metric.Tags() {
		if p.tagConversions.Measurement != nil && p.tagConversions.Measurement.Match(key) {
			metric.RemoveTag(key)
			metric.SetName(value)
			continue
		}

		if p.tagConversions.String != nil && p.tagConversions.String.Match(key) {
			metric.RemoveTag(key)
			metric.AddField(key, value)
			continue
		}

		if p.tagConversions.Integer != nil && p.tagConversions.Integer.Match(key) {
			v, ok := toInteger(value)
			if !ok {
				metric.RemoveTag(key)
				p.Log.Errorf("error converting to integer [%T]: %v", value, value)
				continue
			}

			metric.RemoveTag(key)
			metric.AddField(key, v)
		}

		if p.tagConversions.Unsigned != nil && p.tagConversions.Unsigned.Match(key) {
			v, ok := toUnsigned(value)
			if !ok {
				metric.RemoveTag(key)
				p.Log.Errorf("error converting to unsigned [%T]: %v", value, value)
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
				p.Log.Errorf("error converting to boolean [%T]: %v", value, value)
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
				p.Log.Errorf("error converting to float [%T]: %v", value, value)
				continue
			}

			metric.RemoveTag(key)
			metric.AddField(key, v)
			continue
		}
	}
}

// convertFields converts fields into measurements, tags, or other field types.
func (p *Converter) convertFields(metric telegraf.Metric) {
	if p.fieldConversions == nil {
		return
	}

	for key, value := range metric.Fields() {
		if p.fieldConversions.Measurement != nil && p.fieldConversions.Measurement.Match(key) {
			v, ok := toString(value)
			if !ok {
				metric.RemoveField(key)
				p.Log.Errorf("error converting to measurement [%T]: %v", value, value)
				continue
			}

			metric.RemoveField(key)
			metric.SetName(v)
			continue
		}

		if p.fieldConversions.Tag != nil && p.fieldConversions.Tag.Match(key) {
			v, ok := toString(value)
			if !ok {
				metric.RemoveField(key)
				p.Log.Errorf("error converting to tag [%T]: %v", value, value)
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
				p.Log.Errorf("error converting to float [%T]: %v", value, value)
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
				p.Log.Errorf("error converting to integer [%T]: %v", value, value)
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
				p.Log.Errorf("error converting to unsigned [%T]: %v", value, value)
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
				p.Log.Errorf("error converting to bool [%T]: %v", value, value)
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
				p.Log.Errorf("Error converting to string [%T]: %v", value, value)
				continue
			}

			metric.RemoveField(key)
			metric.AddField(key, v)
			continue
		}
	}
}

func toBool(v interface{}) (val bool, ok bool) {
	switch value := v.(type) {
	case int64:
		return value != 0, true
	case uint64:
		return value != 0, true
	case float64:
		return value != 0, true
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
		}
		return math.MaxInt64, true
	case float64:
		if value < float64(math.MinInt64) {
			return math.MinInt64, true
		} else if value > float64(math.MaxInt64) {
			return math.MaxInt64, true
		} else {
			return int64(math.Round(value)), true
		}
	case bool:
		if value {
			return 1, true
		}
		return 0, true
	case string:
		result, err := strconv.ParseInt(value, 0, 64)

		if err != nil {
			var result float64
			var err error

			if isHexadecimal(value) {
				result, err = parseHexadecimal(value)
			} else {
				result, err = strconv.ParseFloat(value, 64)
			}

			if err != nil {
				return 0, false
			}

			return toInteger(result)
		}
		return result, true
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
		}
		return uint64(value), true
	case float64:
		if value < 0.0 {
			return 0, true
		} else if value > float64(math.MaxUint64) {
			return math.MaxUint64, true
		} else {
			return uint64(math.Round(value)), true
		}
	case bool:
		if value {
			return 1, true
		}
		return 0, true
	case string:
		result, err := strconv.ParseUint(value, 0, 64)

		if err != nil {
			var result float64
			var err error

			if isHexadecimal(value) {
				result, err = parseHexadecimal(value)
			} else {
				result, err = strconv.ParseFloat(value, 64)
			}

			if err != nil {
				return 0, false
			}

			return toUnsigned(result)
		}
		return result, true
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
		}
		return 0.0, true
	case string:
		if isHexadecimal(value) {
			result, err := parseHexadecimal(value)
			return result, err == nil
		}

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

func parseHexadecimal(value string) (float64, error) {
	i := new(big.Int)

	_, success := i.SetString(value, 0)
	if !success {
		return 0, errors.New("unable to parse string to big int")
	}

	f := new(big.Float).SetInt(i)
	result, _ := f.Float64()

	return result, nil
}

func isHexadecimal(value string) bool {
	return len(value) >= 3 && strings.ToLower(value)[1] == 'x'
}

func init() {
	processors.Add("converter", func() telegraf.Processor {
		return &Converter{}
	})
}
