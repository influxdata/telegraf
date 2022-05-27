package json_template

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func TestSerializeMetric(t *testing.T) {
	tmpl := `
	{
		"sdkVersion": "{{.Tags.sdkver}}",
		"time": {{.Time.Unix}},
		"platform": "Java",
		"key": "{{.Tags.key}}",
		"events": [
		{
			"time": {{.Time.Unix}},
			"type": "{{.Name | uppercase}}",
			"flag": "{{.Tags.flagname}}",
			"experimentVersion": "0",
			"value": "{{.Tags.value}}",
			"count": {{.Fields.count_sum}}
		}
		],
		"origin": "Telegraf"
	}
    `
	m := metric.New(
		"impression",
		map[string]string{
			"key":      "12345",
			"flagname": "F5",
			"host":     "1cbbb3796fc2",
			"platform": "Java",
			"sdkver":   "4.9.1",
			"value":    "false",
		},
		map[string]interface{}{"count_sum": 5},
		time.Unix(1653643420, 0),
	)

	s, err := NewSerializer(tmpl, "raw")
	require.NoError(t, err)

	buf, err := s.Serialize(m)
	require.NoError(t, err)

	expected := map[string]interface{}{
		"sdkVersion": "4.9.1",
		"time":       float64(1653643420),
		"platform":   "Java",
		"key":        "12345",
		"events": []interface{}{
			map[string]interface{}{
				"time":              float64(1653643420),
				"type":              "IMPRESSION",
				"flag":              "F5",
				"experimentVersion": "0",
				"value":             "false",
				"count":             float64(5),
			},
		},
		"origin": "Telegraf",
	}

	var actual map[string]interface{}
	require.NoError(t, json.Unmarshal(buf, &actual))
	require.EqualValues(t, expected, actual)
}

func TestSerializeMetricCompact(t *testing.T) {
	tmpl := `
	{
		"sdkVersion": "{{.Tags.sdkver}}",
		"time": {{.Time.Unix}},
		"platform": "Java",
		"key": "{{.Tags.key}}",
		"events": [
		{
			"time": {{.Time.Unix}},
			"type": "{{.Name | uppercase}}",
			"flag": "{{.Tags.flagname}}",
			"experimentVersion": "0",
			"value": "{{.Tags.value}}",
			"count": {{.Fields.count_sum}}
		}
		],
		"origin": "Telegraf"
	}
    `
	m := metric.New(
		"impression",
		map[string]string{
			"key":      "12345",
			"flagname": "F5",
			"host":     "1cbbb3796fc2",
			"platform": "Java",
			"sdkver":   "4.9.1",
			"value":    "false",
		},
		map[string]interface{}{"count_sum": 5},
		time.Unix(1653643420, 0),
	)

	s, err := NewSerializer(tmpl, "compact")
	require.NoError(t, err)

	buf, err := s.Serialize(m)
	require.NoError(t, err)
	require.Len(t, buf, 205)
}

func TestSerializeMetricPretty(t *testing.T) {
	tmpl := `
	{
		"sdkVersion": "{{.Tags.sdkver}}",
		"time": {{.Time.Unix}},
		"platform": "Java",
		"key": "{{.Tags.key}}",
		"events": [
		{
			"time": {{.Time.Unix}},
			"type": "{{.Name | uppercase}}",
			"flag": "{{.Tags.flagname}}",
			"experimentVersion": "0",
			"value": "{{.Tags.value}}",
			"count": {{.Fields.count_sum}}
		}
		],
		"origin": "Telegraf"
	}
    `
	m := metric.New(
		"impression",
		map[string]string{
			"key":      "12345",
			"flagname": "F5",
			"host":     "1cbbb3796fc2",
			"platform": "Java",
			"sdkver":   "4.9.1",
			"value":    "false",
		},
		map[string]interface{}{"count_sum": 5},
		time.Unix(1653643420, 0),
	)

	s, err := NewSerializer(tmpl, "pretty")
	require.NoError(t, err)

	buf, err := s.Serialize(m)
	require.NoError(t, err)
	require.Len(t, buf, 296)
}

func TestSerializeMetricInvalidStyle(t *testing.T) {
	_, err := NewSerializer("", "undefined")
	require.Error(t, err, "unknown style \"undefined\"")
}

func TestSerializeMetricBatch(t *testing.T) {
	tmpl := `
	{
		"sdkVersion": "{{(index . 0).Tags.sdkver}}",
		"time": {{(index . 0).Time.Unix}},
		"platform": "{{(index . 1).Tags.platform}}",
		"key": "{{(index . 0).Tags.key}}",
		"events": [
{{- range $idx, $metric := . }}
			{
				"time": {{$metric.Time.Unix}},
				"type": "{{$metric.Name | uppercase}}",
				"flag": "{{$metric.Tags.flagname}}",
				"experimentVersion": "0",
				"value": "{{$metric.Tags.value}}",
				"count": {{$metric.Fields.count_sum}}
			}{{- if not (last $idx $metrics)}},{{- end}}
{{- end}}
		],
		"origin": "Telegraf"
	}
   `
	m := []telegraf.Metric{
		metric.New(
			"impression",
			map[string]string{
				"key":      "12345",
				"flagname": "F5",
				"host":     "1cbbb3796fc2",
				"platform": "Java",
				"sdkver":   "4.9.1",
				"value":    "false",
			},
			map[string]interface{}{"count_sum": 5},
			time.Unix(1653643420, 0),
		),
		metric.New(
			"expression",
			map[string]string{
				"key":      "67890",
				"flagname": "E27",
				"host":     "509743jcr",
				"platform": "Golang",
				"sdkver":   "1.18.1",
				"value":    "true",
			},
			map[string]interface{}{"count_sum": 42},
			time.Unix(1653653420, 0),
		),
	}
	s, err := NewSerializer(tmpl, "raw")
	require.NoError(t, err)

	buf, err := s.SerializeBatch(m)
	require.NoError(t, err)

	expected := map[string]interface{}{
		"sdkVersion": "4.9.1",
		"time":       float64(1653643420),
		"platform":   "Golang",
		"key":        "12345",
		"events": []interface{}{
			map[string]interface{}{
				"time":              float64(1653643420),
				"type":              "IMPRESSION",
				"flag":              "F5",
				"experimentVersion": "0",
				"value":             "false",
				"count":             float64(5),
			},
			map[string]interface{}{
				"time":              float64(1653653420),
				"type":              "EXPRESSION",
				"flag":              "E27",
				"experimentVersion": "0",
				"value":             "true",
				"count":             float64(42),
			},
		},
		"origin": "Telegraf",
	}

	var actual map[string]interface{}
	require.NoError(t, json.Unmarshal(buf, &actual))
	require.EqualValues(t, expected, actual)
}

func TestSerializeMetricDefault(t *testing.T) {
	tmpl := `
	[
{{- range $idx, $metric := . }}
		{
			"name":  "{{$metric.Name}}",
			"fields": {
				{{template "fields" $metric}}
			},
			"tags": {
				{{template "tags" $metric}}
			},
			"time":   {{$metric.Time.Unix}}
		}{{- if not (last $idx $metrics)}},{{- end}}
{{- end}}
	]
   `
	m := []telegraf.Metric{
		metric.New(
			"impression",
			map[string]string{
				"key":      "12345",
				"flagname": "F5",
				"host":     "1cbbb3796fc2",
				"platform": "Java",
				"sdkver":   "4.9.1",
				"value":    "false",
			},
			map[string]interface{}{
				"an_int":   5,
				"a_float":  3.1415,
				"a_bool":   true,
				"a_string": "some arbitrary string",
			},
			time.Unix(1653643420, 0),
		),
		metric.New(
			"expression",
			map[string]string{
				"key":      "67890",
				"flagname": "E27",
				"host":     "509743jcr",
				"platform": "Golang",
				"sdkver":   "1.18.1",
				"value":    "true",
			},
			map[string]interface{}{"count_sum": 42},
			time.Unix(1653653420, 0),
		),
	}
	s, err := NewSerializer(tmpl, "raw")
	require.NoError(t, err)

	buf, err := s.SerializeBatch(m)
	require.NoError(t, err, string(buf))

	expectedRaw := make([]interface{}, 0)
	for _, x := range m {
		expectedRaw = append(expectedRaw, map[string]interface{}{
			"name":   x.Name(),
			"time":   x.Time().Unix(),
			"fields": x.Fields(),
			"tags":   x.Tags(),
		})
	}
	var expected []interface{}
	tmp, err := json.Marshal(expectedRaw)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(tmp, &expected))

	var actual []interface{}
	require.NoError(t, json.Unmarshal(buf, &actual))
	require.EqualValues(t, expected, actual)
}
