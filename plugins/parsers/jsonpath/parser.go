package jsonpath

import (
	"fmt"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
)

type TimeFunc func() time.Time

type Parser struct {
	Configs     []Config
	DefaultTags map[string]string
	Log         telegraf.Logger
	TimeFunc    func() time.Time
}

type Config struct {
	MetricSelection string `toml:"metric_selection"`
	MetricName      string `toml:"metric_name"`
	Fields          []FieldKeys
}

type FieldKeys struct {
	Name  string `toml:"name"`
	Query string `toml:"query"`
	Type  string `toml:"type"`
}

func (p *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	err := oj.Validate(buf)
	if err != nil {
		return nil, fmt.Errorf("The provided JSON is invalid: %v", err)
	}

	obj, err := oj.Parse(buf)
	if err != nil {
		return nil, err
	}

	var t []telegraf.Metric
	for _, config := range p.Configs {
		if len(config.MetricSelection) == 0 {
			config.MetricSelection = "*"
		}
		m, err := p.query(obj, config)
		if err != nil {
			return nil, err
		}
		t = append(t, m)
	}

	return t, nil
}

func (p *Parser) ParseLine(line string) (telegraf.Metric, error) {
	err := oj.ValidateString(line)
	if err != nil {
		return nil, fmt.Errorf("The provided JSON is invalid: %v", err)
	}

	obj, err := oj.ParseString(line)
	if err != nil {
		return nil, err
	}

	switch len(p.Configs) {
	case 0:
		return nil, nil
	case 1:
		return p.query(obj, p.Configs[0])
	}

	return nil, fmt.Errorf("cannot parse line with multiple (%d) configurations", len(p.Configs))
}

func (p *Parser) query(obj interface{}, config Config) (telegraf.Metric, error) {
	tags := map[string]string{}
	fields := map[string]interface{}{}
	for _, field := range config.Fields {
		x, err := jp.ParseString(field.Query)
		if err != nil {
			return nil, err
		}

		result := x.Get(obj)
		for i, input := range result {
			// If a field type is defined, check if needs to be converted
			if field.Type != "" {
				result[i], err = p.convertType(input, field)
				if err != nil {
					return nil, err
				}
			}

			fields[field.Name] = result[i]
		}

		if p.TimeFunc == nil {
			p.TimeFunc = time.Now
		}
	}

	return metric.New(config.MetricName, tags, fields, p.TimeFunc()), nil
}

func (p *Parser) SetDefaultTags(tags map[string]string) {

}

func (p *Parser) SetTimeFunc(fn TimeFunc) {
	p.TimeFunc = fn
}

// convertType will convert the value parsed from the input JSON to the specified type in the config
func (p *Parser) convertType(input interface{}, configField FieldKeys) (interface{}, error) {
	switch inputType := input.(type) {
	case string:
		if configField.Type != "string" {
			switch configField.Type {
			case "int":
				r, err := strconv.Atoi(inputType)
				if err != nil {
					return nil, fmt.Errorf("Unable to convert field '%s' to type int: %v", configField.Name, err)
				}
				return r, nil
			case "float":
				r, err := strconv.ParseFloat(inputType, 64)
				if err != nil {
					return nil, fmt.Errorf("Unable to convert field '%s' to type float: %v", configField.Name, err)
				}
				return r, nil
			case "bool":
				r, err := strconv.ParseBool(inputType)
				if err != nil {
					return nil, fmt.Errorf("Unable to convert field '%s' to type bool: %v", configField.Name, err)
				}
				return r, nil
			}
		}
	case bool:
		if configField.Type != "bool" {
			switch configField.Type {
			case "string":
				return strconv.FormatBool(inputType), nil
			case "int":
				if inputType {
					return int64(1), nil
				}

				return int64(0), nil
			}
		}
	case int64:
		if configField.Type != "int" {
			switch configField.Type {
			case "string":
				return fmt.Sprint(inputType), nil
			case "float":
				return float64(inputType), nil
			case "bool":
				if inputType == 0 {
					return false, nil
				} else if inputType == 1 {
					return true, nil
				} else {
					return nil, fmt.Errorf("Unable to convert field '%s' to type bool", configField.Name)
				}
			}
		}
	case float64:
		if configField.Type != "float" {
			switch configField.Type {
			case "string":
				return fmt.Sprint(inputType), nil
			case "int":
				return int64(inputType), nil
			}
		}
	default:
		return nil, fmt.Errorf("unknown format '%T' for field  '%s'", inputType, configField.Name)
	}

	return input, nil
}
