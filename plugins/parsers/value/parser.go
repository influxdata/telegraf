package value

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
)

type ValueParser struct {
	MetricName  string
	DataType    string
	DefaultTags map[string]string
}

func (v *ValueParser) Parse(buf []byte) ([]telegraf.Metric, error) {
	// separate out any fields in the buffer, ignore anything but the last.
	values := bytes.Fields(buf)
	if len(values) < 1 {
		return []telegraf.Metric{}, nil
	}
	valueStr := string(values[len(values)-1])

	var value interface{}
	var err error
	switch v.DataType {
	case "", "auto":
		value, err = autoParse(valueStr)
	case "int", "integer":
		value, err = strconv.Atoi(valueStr)
	case "float", "long":
		value, err = strconv.ParseFloat(valueStr, 64)
	case "str", "string":
		value = valueStr
	case "bool", "boolean":
		value, err = strconv.ParseBool(valueStr)
	}
	if err != nil {
		return nil, err
	}

	fields := map[string]interface{}{"value": value}
	metric, err := telegraf.NewMetric(v.MetricName, v.DefaultTags,
		fields, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	return []telegraf.Metric{metric}, nil
}

// autoParse tries to parse the given string into (in order):
//   1. integer
//   2. float
//   3. boolean
//   4. string
func autoParse(valueStr string) (interface{}, error) {
	var value interface{}
	var err error
	// 1st try integer:
	if value, err = strconv.Atoi(valueStr); err == nil {
		return value, err
	}
	// 2nd try float:
	if value, err = strconv.ParseFloat(valueStr, 64); err == nil {
		return value, err
	}
	// 3rd try boolean:
	if value, err = strconv.ParseBool(valueStr); err == nil {
		return value, err
	}
	// 4th, none worked, so string
	return valueStr, nil
}

func (v *ValueParser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := v.Parse([]byte(line))

	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, fmt.Errorf("Can not parse the line: %s, for data format: value", line)
	}

	return metrics[0], nil
}

func (v *ValueParser) SetDefaultTags(tags map[string]string) {
	v.DefaultTags = tags
}
