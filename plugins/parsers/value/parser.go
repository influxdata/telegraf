package value

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

type ValueParser struct {
	MetricName  string
	DataType    string
	DefaultTags map[string]string
	FieldName   string
}

func (v *ValueParser) Parse(buf []byte) ([]telegraf.Metric, error) {
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
	case "", "int", "integer":
		value, err = strconv.Atoi(vStr)
	case "float", "long":
		value, err = strconv.ParseFloat(vStr, 64)
	case "str", "string":
		value = vStr
	case "bool", "boolean":
		value, err = strconv.ParseBool(vStr)
	}
	if err != nil {
		return nil, err
	}

	fields := map[string]interface{}{v.FieldName: value}
	m := metric.New(v.MetricName, v.DefaultTags,
		fields, time.Now().UTC())

	return []telegraf.Metric{m}, nil
}

func (v *ValueParser) ParseLine(line string) (telegraf.Metric, error) {
	metrics, err := v.Parse([]byte(line))

	if err != nil {
		return nil, err
	}

	if len(metrics) < 1 {
		return nil, fmt.Errorf("can not parse the line: %s, for data format: value", line)
	}

	return metrics[0], nil
}

func (v *ValueParser) SetDefaultTags(tags map[string]string) {
	v.DefaultTags = tags
}

func NewValueParser(metricName, dataType, fieldName string, defaultTags map[string]string) *ValueParser {
	if fieldName == "" {
		fieldName = "value"
	}

	return &ValueParser{
		MetricName:  metricName,
		DataType:    dataType,
		DefaultTags: defaultTags,
		FieldName:   fieldName,
	}
}
