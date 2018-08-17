package parser

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
	if len(metrics1) != len(metrics2) {
		t.Errorf("Output doesn't match expected")
	}
	for i, m1 := range metrics1 {
		m2 := metrics2[i]
		if m1 == nil || m2 == nil {
			t.Errorf("Metric is nil, can't compare")
			continue
		}
		require.True(t, reflect.DeepEqual(m1.Tags(), m2.Tags()))
		require.True(t, reflect.DeepEqual(m1.Fields(), m2.Fields()))
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
		original    string
	}{
		{
			name:        "parse one field [replace]",
			parseFields: []string{"sample"},
			original:    "replace",
			config: parsers.Config{
				DataFormat: "json",
				TagKeys: []string{
					"ts",
					"lvl",
					"msg",
					"method",
				},
			},
			input: Metric(
				metric.New(
					"singleField",
					map[string]string{
						"some": "tag",
					},
					map[string]interface{}{
						"sample": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"singleField",
					map[string]string{
						"ts":     "2018-07-24T19:43:40.275Z",
						"lvl":    "info",
						"msg":    "http request",
						"method": "POST",
					},
					map[string]interface{}{},
					time.Unix(0, 0))),
			},
		},
		{
			name:        "parse one field [merge]",
			parseFields: []string{"sample"},
			original:    "merge",
			config: parsers.Config{
				DataFormat: "json",
				TagKeys: []string{
					"ts",
					"lvl",
					"msg",
					"method",
				},
			},
			input: Metric(
				metric.New(
					"singleField",
					map[string]string{
						"some": "tag",
					},
					map[string]interface{}{
						"sample": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"singleField",
					map[string]string{
						"some":   "tag",
						"ts":     "2018-07-24T19:43:40.275Z",
						"lvl":    "info",
						"msg":    "http request",
						"method": "POST",
					},
					map[string]interface{}{
						"sample": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
					},
					time.Unix(0, 0))),
			},
		},
		{
			name:        "parse one field [keep]",
			parseFields: []string{"sample"},
			original:    "keep",
			config: parsers.Config{
				DataFormat: "json",
				TagKeys: []string{
					"ts",
					"lvl",
					"msg",
					"method",
				},
			},
			input: Metric(
				metric.New(
					"singleField",
					map[string]string{
						"some": "tag",
					},
					map[string]interface{}{
						"sample": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"singleField",
					map[string]string{
						"some": "tag",
					},
					map[string]interface{}{
						"sample": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
					},
					time.Unix(0, 0))),
				Metric(metric.New(
					"singleField",
					map[string]string{
						"ts":     "2018-07-24T19:43:40.275Z",
						"lvl":    "info",
						"msg":    "http request",
						"method": "POST",
					},
					map[string]interface{}{},
					time.Unix(0, 0))),
			},
		},
		{
			name:        "parse influx field [keep]",
			parseFields: []string{"message"},
			config: parsers.Config{
				DataFormat: "influx",
			},
			input: Metric(
				metric.New(
					"influxField",
					map[string]string{},
					map[string]interface{}{
						"message": "deal,computer_name=hosta message=\"stuff\" 1530654676316265790",
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"influxField",
					map[string]string{},
					map[string]interface{}{
						"message": "deal,computer_name=hosta message=\"stuff\" 1530654676316265790",
					},
					time.Unix(0, 0))),
				Metric(metric.New(
					"deal",
					map[string]string{
						"computer_name": "hosta",
					},
					map[string]interface{}{
						"message": "stuff",
					},
					time.Unix(0, 0))),
			},
		},
		{
			name:        "parse influx field [merge]",
			parseFields: []string{"message"},
			original:    "merge",
			config: parsers.Config{
				DataFormat: "influx",
			},
			input: Metric(
				metric.New(
					"influxField",
					map[string]string{
						"some": "tag",
					},
					map[string]interface{}{
						"message": "deal,computer_name=hosta message=\"stuff\" 1530654676316265790",
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"influxField",
					map[string]string{
						"computer_name": "hosta",
						"some":          "tag",
					},
					map[string]interface{}{
						"message": "stuff",
					},
					time.Unix(0, 0))),
			},
		},
		{
			name:        "parse grok field",
			parseFields: []string{"grokSample"},
			original:    "replace",
			config: parsers.Config{
				DataFormat:   "grok",
				GrokPatterns: []string{"%{COMBINED_LOG_FORMAT}"},
			},
			input: Metric(
				metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"grokSample": "127.0.0.1 - - [11/Dec/2013:00:01:45 -0800] \"GET /xampp/status.php HTTP/1.1\" 200 3891 \"http://cadenza/xampp/navi.php\" \"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.9; rv:25.0) Gecko/20100101 Firefox/25.0\"",
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"sucess",
					map[string]string{
						"resp_code": "200",
						"verb":      "GET",
					},
					map[string]interface{}{
						"resp_bytes":   int64(3891),
						"auth":         "-",
						"request":      "/xampp/status.php",
						"referrer":     "http://cadenza/xampp/navi.php",
						"agent":        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.9; rv:25.0) Gecko/20100101 Firefox/25.0",
						"client_ip":    "127.0.0.1",
						"ident":        "-",
						"http_version": float64(1.1),
					},
					time.Unix(0, 0))),
			},
		},
		{
			name:        "parse two fields [replace]",
			parseFields: []string{"field_1", "field_2"},
			original:    "replace",
			config: parsers.Config{
				DataFormat: "json",
				TagKeys:    []string{"lvl", "err"},
			},
			input: Metric(
				metric.New(
					"bigMeasure",
					map[string]string{},
					map[string]interface{}{
						"field_1": `{"lvl":"info","msg":"http request"}`,
						"field_2": `{"err":"fatal","fatal":"security threat"}`,
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"bigMeasure",
					map[string]string{
						"lvl": "info",
						"err": "fatal",
					},
					map[string]interface{}{},
					time.Unix(0, 0))),
			},
		},
		{
			name:        "parse two fields [merge]",
			parseFields: []string{"field_1", "field_2"},
			original:    "merge",
			config: parsers.Config{
				DataFormat: "json",
				TagKeys:    []string{"lvl", "msg", "err", "fatal"},
			},
			input: Metric(
				metric.New(
					"bigMeasure",
					map[string]string{},
					map[string]interface{}{
						"field_1": `{"lvl":"info","msg":"http request"}`,
						"field_2": `{"err":"fatal","fatal":"security threat"}`,
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"bigMeasure",
					map[string]string{
						"lvl":   "info",
						"msg":   "http request",
						"err":   "fatal",
						"fatal": "security threat",
					},
					map[string]interface{}{
						"field_1": `{"lvl":"info","msg":"http request"}`,
						"field_2": `{"err":"fatal","fatal":"security threat"}`,
					},
					time.Unix(0, 0))),
			},
		},
		{
			name:        "parse two fields [keep]",
			parseFields: []string{"field_1", "field_2"},
			original:    "keep",
			config: parsers.Config{
				DataFormat: "json",
				TagKeys:    []string{"lvl", "msg", "err", "fatal"},
			},
			input: Metric(
				metric.New(
					"bigMeasure",
					map[string]string{},
					map[string]interface{}{
						"field_1": `{"lvl":"info","msg":"http request"}`,
						"field_2": `{"err":"fatal","fatal":"security threat"}`,
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"bigMeasure",
					map[string]string{},
					map[string]interface{}{
						"field_1": `{"lvl":"info","msg":"http request"}`,
						"field_2": `{"err":"fatal","fatal":"security threat"}`,
					},
					time.Unix(0, 0))),
				Metric(metric.New(
					"bigMeasure",
					map[string]string{
						"lvl":   "info",
						"msg":   "http request",
						"err":   "fatal",
						"fatal": "security threat",
					},
					map[string]interface{}{},
					time.Unix(0, 0))),
			},
		},
		{
			name:        "Fail to parse one field but parses other [keep]",
			parseFields: []string{"good", "bad"},
			original:    "keep",
			config: parsers.Config{
				DataFormat: "json",
				TagKeys:    []string{"lvl"},
			},
			input: Metric(
				metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"good": `{"lvl":"info"}`,
						"bad":  "why",
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"good": `{"lvl":"info"}`,
						"bad":  "why",
					},
					time.Unix(0, 0))),
				Metric(metric.New(
					"success",
					map[string]string{
						"lvl": "info",
					},
					map[string]interface{}{},
					time.Unix(0, 0))),
			},
		},
		{
			name:        "Fail to parse one field but parses other [keep] v2",
			parseFields: []string{"bad", "good", "ok"},
			original:    "keep",
			config: parsers.Config{
				DataFormat: "json",
				TagKeys:    []string{"lvl", "thing"},
			},
			input: Metric(
				metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"bad":  "why",
						"good": `{"lvl":"info"}`,
						"ok":   `{"thing":"thang"}`,
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"success",
					map[string]string{},
					map[string]interface{}{
						"bad":  "why",
						"good": `{"lvl":"info"}`,
						"ok":   `{"thing":"thang"}`,
					},
					time.Unix(0, 0))),
				Metric(metric.New(
					"success",
					map[string]string{
						"lvl":   "info",
						"thing": "thang",
					},
					map[string]interface{}{},
					time.Unix(0, 0))),
			},
		},
		{
			name:        "Fail to parse one field but parses other [merge]",
			parseFields: []string{"good", "bad"},
			original:    "merge",
			config: parsers.Config{
				DataFormat: "json",
				TagKeys:    []string{"lvl"},
			},
			input: Metric(
				metric.New(
					"success",
					map[string]string{
						"a": "tag",
					},
					map[string]interface{}{
						"good": `{"lvl":"info"}`,
						"bad":  "why",
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"success",
					map[string]string{
						"a":   "tag",
						"lvl": "info",
					},
					map[string]interface{}{
						"good": `{"lvl":"info"}`,
						"bad":  "why",
					},
					time.Unix(0, 0))),
			},
		},
		{
			name:        "Fail to parse one field but parses other [replace]",
			parseFields: []string{"good", "bad"},
			original:    "replace",
			config: parsers.Config{
				DataFormat: "json",
				TagKeys:    []string{"lvl"},
			},
			input: Metric(
				metric.New(
					"success",
					map[string]string{
						"thing": "tag",
					},
					map[string]interface{}{
						"good": `{"lvl":"info"}`,
						"bad":  "why",
					},
					time.Unix(0, 0))),
			expected: []telegraf.Metric{
				Metric(metric.New(
					"success",
					map[string]string{
						"lvl": "info",
					},
					map[string]interface{}{},
					time.Unix(0, 0))),
			},
		},
	}

	for _, tt := range tests {
		parser := Parser{
			Config:      tt.config,
			ParseFields: tt.parseFields,
			Original:    tt.original,
		}

		output := parser.Apply(tt.input)
		t.Logf("Testing: %s", tt.name)
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
				DataFormat: "json",
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
		{
			name:        "non string field",
			parseFields: []string{"some_field"},
			config: parsers.Config{
				DataFormat: "json",
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
		parser := Parser{
			Config:      tt.config,
			ParseFields: tt.parseFields,
		}

		output := parser.Apply(tt.input)

		compareMetrics(t, output, tt.expected)
	}
}

// Benchmarks

func getMetricFields(metric telegraf.Metric) interface{} {
	key := "field3"
	if value, ok := metric.Fields()[key]; ok {
		return value
	}
	return nil
}

func getMetricFieldList(metric telegraf.Metric) interface{} {
	key := "field3"
	fields := metric.FieldList()
	for _, field := range fields {
		if field.Key == key {
			return field.Value
		}
	}
	return nil
}

func BenchmarkFieldListing(b *testing.B) {
	metric := Metric(metric.New(
		"test",
		map[string]string{
			"some": "tag",
		},
		map[string]interface{}{
			"field0": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
			"field1": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
			"field2": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
			"field3": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
			"field4": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
			"field5": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
			"field6": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
		},
		time.Unix(0, 0)))

	for n := 0; n < b.N; n++ {
		getMetricFieldList(metric)
	}
}

func BenchmarkFields(b *testing.B) {
	metric := Metric(metric.New(
		"test",
		map[string]string{
			"some": "tag",
		},
		map[string]interface{}{
			"field0": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
			"field1": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
			"field2": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
			"field3": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
			"field4": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
			"field5": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
			"field6": `{"ts":"2018-07-24T19:43:40.275Z","lvl":"info","msg":"http request","method":"POST"}`,
		},
		time.Unix(0, 0)))

	for n := 0; n < b.N; n++ {
		getMetricFields(metric)
	}
}
