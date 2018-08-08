package fieldparser

import (
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/stretchr/testify/require"
)

//compares metrics without comparing time
func compareMetrics(t *testing.T, metrics1 []telegraf.Metric, metrics2 []telegraf.Metric) {
	for i, m1 := range metrics1 {
		m2 := metrics2[i]
		require.True(t, reflect.DeepEqual(m1.Tags(), m2.Tags()))
		require.True(t, reflect.DeepEqual(m1.Fields(), m2.Fields()))
		//require.True(t, m1.Name() == m2.Name())
	}
}

func Metric(v telegraf.Metric, err error) telegraf.Metric {
	if err != nil {
		panic(err)
	}
	return v
}

func TestApply(t *testing.T) {
	tests := []struct {
		name        string
		parseFields []string
		config      parsers.Config
		input       telegraf.Metric
		expected    []telegraf.Metric
	}{
		{
			name: "parse tag and fields",
			config: parsers.Config{
				DataFormat: "logfmt",
			},
			input: Metric(
				metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"test_name": `ts=2018-07-24T19:43:40.275Z lvl=info msg="http request" method=POST`,
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"test_name": `ts=2018-07-24T19:43:40.275Z lvl=info msg="http request" method=POST`,
					},
					time.Unix(0, 0))),
				Metric(metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"ts":     "2018-07-24T19:43:40.275Z",
						"lvl":    "info",
						"msg":    "http request",
						"method": "post",
					},
					time.Unix(0, 0))),
			},
		},
		{
			name: "parse one field",
			config: parsers.Config{
				DataFormat: "logfmt",
			},
			input: Metric(
				metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"level_test": `lvl=info`,
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"level_test": `lvl=info`,
					},
					time.Unix(0, 0))),
				Metric(metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"lvl": "info",
					},
					time.Unix(0, 0))),
			},
		},
		{
			name: "parse two fields",
			config: parsers.Config{
				DataFormat: "logfmt",
			},
			input: Metric(
				metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"test": "ts=2018-07-24T19:43:40.275Z lvl=info",
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"test": "ts=2018-07-24T19:43:40.275Z lvl=info",
					},
					time.Unix(0, 0))),
				Metric(metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"ts":  "2018-07-24T19:43:40.275Z",
						"lvl": "info",
					},
					time.Unix(0, 0))),
			},
		},
		{
			name: "parse one tag",
			config: parsers.Config{
				DataFormat: "logfmt",
			},
			input: Metric(
				metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"test_name": `ts=2018-07-24T19:43:40.275Z`,
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"test_name": `ts=2018-07-24T19:43:40.275Z`,
					},
					time.Unix(0, 0))),
				Metric(metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"ts": "2018-07-24T19:43:40.275Z",
					},
					time.Unix(0, 0))),
			},
		},
		{
			name: "parse two tags",
			config: parsers.Config{
				DataFormat: "logfmt",
			},
			input: Metric(
				metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"timestamp": `ts=2018-07-24T19:43:40.275Z`,
						"message":   `msg="http request"`,
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"timestamp": `ts=2018-07-24T19:43:40.275Z`,
						"message":   `msg="http request"`,
					},
					time.Unix(0, 0))),
			},
		},
		{
			name: "Fail to parse fields but still parses tag",
			config: parsers.Config{
				DataFormat: "logfmt",
			},
			input: Metric(
				metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"test_name": `ts=2018-07-24T19:43:40.275Z lvl=info msg="http request" method=POST`,
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"test_name": `ts=2018-07-24T19:43:40.275Z lvl=info msg="http request" method=POST`,
					},
					time.Unix(0, 0))),
				Metric(metric.New(
					"bad",
					map[string]string{},
					map[string]interface{}{},
					time.Unix(0, 0))),
			},
		},
	}

	for _, tt := range tests {
		parser := FieldParser{
			config:      tt.config,
			parseFields: tt.parseFields,
		}

		output := parser.Apply(tt.input)

		compareMetrics(t, output, tt.expected)
	}
}

func TestBadApply(t *testing.T) {
	tests := []struct {
		name        string
		parseFields []string
		config      parsers.Config
		input       telegraf.Metric
		expected    []telegraf.Metric
	}{
		{
			name:        "field not found",
			parseFields: []string{"bad_field"},
			config: parsers.Config{
				DataFormat: "logfmt",
			},
			input: Metric(
				metric.New(
					"bad",
					map[string]string{},
					map[string]interface{}{
						"some_field": 5,
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"bad",
					map[string]string{},
					map[string]interface{}{
						"some_field": 5,
					},
					time.Unix(0, 0))),
			},
		},
	}

	for _, tt := range tests {
		parser := FieldParser{
			config:      tt.config,
			parseFields: tt.parseFields,
		}

		output := parser.Apply(tt.input)

		compareMetrics(t, output, tt.expected)
	}
}
