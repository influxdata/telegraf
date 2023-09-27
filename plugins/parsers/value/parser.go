package value

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type Parser struct {
	DataType    string            `toml:"data_type"`
	FieldName   string            `toml:"value_field_name"`
	MetricName  string            `toml:"-"`
	DefaultTags map[string]string `toml:"-"`
}

func (v *Parser) Init() error {
	switch v.DataType {
	case "", "int", "integer":
		v.DataType = "int"
	case "float", "long":
		v.DataType = "float"
	case "str", "string":
		v.DataType = "string"
	case "bool", "boolean":
		v.DataType = "bool"
	case "auto_integer", "auto_float":
		// Do nothing both are valid
	default:
		return fmt.Errorf("unknown datatype %q", v.DataType)
	}

	if v.FieldName == "" {
		v.FieldName = "value"
	}

	return nil
}

func (v *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	vStr := string(bytes.TrimSpace(bytes.Trim(buf, "\x00")))

	// unless it's a string, separate out any fields in the buffer,
	// ignore anything but the last.
	if v.DataType != "string" {
		values := strings.Fields(vStr)
		if len(values) < 1 {
			return []telegraf.Metric{}, nil
		}
		vStr = values[len(values)-1]
	}

	var value interface{}
	var err error
	switch v.DataType {
	case "int":
		value, err = strconv.Atoi(vStr)
	case "float":
		value, err = strconv.ParseFloat(vStr, 64)
	case "string":
		value = vStr
	case "bool":
		value, err = strconv.ParseBool(vStr)
	case "auto_integer":
		value, err = strconv.Atoi(vStr)
		if err != nil {
			value = vStr
			err = nil
		}
	case "auto_float":
		value, err = strconv.ParseFloat(vStr, 64)
		if err != nil {
			value = vStr
			err = nil
		}
	}
	if err != nil {
		return nil, err
	}

	fields := map[string]interface{}{v.FieldName: value}
	m := metric.New(v.MetricName, v.DefaultTags,
		fields, time.Now().UTC())

	return []telegraf.Metric{m}, nil
}

func (v *Parser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := v.Parse([]byte(line))

	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, fmt.Errorf("can not parse the line: %s, for data format: value", line)
	}

	return metrics[0], nil
}

func (v *Parser) SetDefaultTags(tags map[string]string) {
	v.DefaultTags = tags
}

func init() {
	parsers.Add("value",
		func(defaultMetricName string) telegraf.Parser {
			return &Parser{
				FieldName:  "value",
				MetricName: defaultMetricName,
			}
		},
	)
}
