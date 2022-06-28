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

func (v *Parser) Parse(buf []byte) ([]telegraf.Metric, error) {
	if v.FieldName == "" {
		v.FieldName = "value"
	}

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

// InitFromConfig is a compatibility function to construct the parser the old way
func (v *Parser) InitFromConfig(config *parsers.Config) error {
	v.MetricName = config.MetricName
	v.DefaultTags = config.DefaultTags
	return v.Init()
}

func (v *Parser) Init() error {
	if v.FieldName == "" {
		v.FieldName = "value"
	}

	return nil
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
