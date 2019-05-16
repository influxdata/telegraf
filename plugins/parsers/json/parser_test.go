package json

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const (
	validJSON              = "{\"a\": 5, \"b\": {\"c\": 6}}"
	validJSONNewline       = "\n{\"d\": 7, \"b\": {\"d\": 8}}\n"
	validJSONArray         = "[{\"a\": 5, \"b\": {\"c\": 6}}]"
	validJSONArrayMultiple = "[{\"a\": 5, \"b\": {\"c\": 6}}, {\"a\": 7, \"b\": {\"c\": 8}}]"
	invalidJSON            = "I don't think this is JSON"
	invalidJSON2           = "{\"a\": 5, \"b\": \"c\": 6}}"
)

const validJSONTags = `
{
    "a": 5,
    "b": {
        "c": 6
    },
    "mytag": "foobar",
    "othertag": "baz"
}
`

const validJSONArrayTags = `
[
{
    "a": 5,
    "b": {
        "c": 6
    },
    "mytag": "foo",
    "othertag": "baz"
},
{
    "a": 7,
    "b": {
        "c": 8
    },
    "mytag": "bar",
    "othertag": "baz"
}
]
`

func TestParseValidJSON(t *testing.T) {
	parser := JSONParser{
		MetricName: "json_test",
	}

	// Most basic vanilla test
	metrics, err := parser.Parse([]byte(validJSON))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "json_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{}, metrics[0].Tags())

	// Test that newlines are fine
	metrics, err = parser.Parse([]byte(validJSONNewline))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "json_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"d":   float64(7),
		"b_d": float64(8),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{}, metrics[0].Tags())

	// Test that strings without TagKeys defined are ignored
	metrics, err = parser.Parse([]byte(validJSONTags))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "json_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{}, metrics[0].Tags())

	// Test that whitespace only will parse as an empty list of metrics
	metrics, err = parser.Parse([]byte("\n\t"))
	require.NoError(t, err)
	require.Len(t, metrics, 0)

	// Test that an empty string will parse as an empty list of metrics
	metrics, err = parser.Parse([]byte(""))
	require.NoError(t, err)
	require.Len(t, metrics, 0)
}

func TestParseLineValidJSON(t *testing.T) {
	parser := JSONParser{
		MetricName: "json_test",
	}

	// Most basic vanilla test
	metric, err := parser.ParseLine(validJSON)
	require.NoError(t, err)
	require.Equal(t, "json_test", metric.Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metric.Fields())
	require.Equal(t, map[string]string{}, metric.Tags())

	// Test that newlines are fine
	metric, err = parser.ParseLine(validJSONNewline)
	require.NoError(t, err)
	require.Equal(t, "json_test", metric.Name())
	require.Equal(t, map[string]interface{}{
		"d":   float64(7),
		"b_d": float64(8),
	}, metric.Fields())
	require.Equal(t, map[string]string{}, metric.Tags())

	// Test that strings without TagKeys defined are ignored
	metric, err = parser.ParseLine(validJSONTags)
	require.NoError(t, err)
	require.Equal(t, "json_test", metric.Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metric.Fields())
	require.Equal(t, map[string]string{}, metric.Tags())
}

func TestParseInvalidJSON(t *testing.T) {
	parser := JSONParser{
		MetricName: "json_test",
	}

	_, err := parser.Parse([]byte(invalidJSON))
	require.Error(t, err)
	_, err = parser.Parse([]byte(invalidJSON2))
	require.Error(t, err)
	_, err = parser.ParseLine(invalidJSON)
	require.Error(t, err)
}

func TestParseWithTagKeys(t *testing.T) {
	// Test that strings not matching tag keys are ignored
	parser := JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{"wrongtagkey"},
	}
	metrics, err := parser.Parse([]byte(validJSONTags))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "json_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{}, metrics[0].Tags())

	// Test that single tag key is found and applied
	parser = JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{"mytag"},
	}
	metrics, err = parser.Parse([]byte(validJSONTags))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "json_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{
		"mytag": "foobar",
	}, metrics[0].Tags())

	// Test that both tag keys are found and applied
	parser = JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{"mytag", "othertag"},
	}
	metrics, err = parser.Parse([]byte(validJSONTags))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "json_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{
		"mytag":    "foobar",
		"othertag": "baz",
	}, metrics[0].Tags())
}

func TestParseLineWithTagKeys(t *testing.T) {
	// Test that strings not matching tag keys are ignored
	parser := JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{"wrongtagkey"},
	}
	metric, err := parser.ParseLine(validJSONTags)
	require.NoError(t, err)
	require.Equal(t, "json_test", metric.Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metric.Fields())
	require.Equal(t, map[string]string{}, metric.Tags())

	// Test that single tag key is found and applied
	parser = JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{"mytag"},
	}
	metric, err = parser.ParseLine(validJSONTags)
	require.NoError(t, err)
	require.Equal(t, "json_test", metric.Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metric.Fields())
	require.Equal(t, map[string]string{
		"mytag": "foobar",
	}, metric.Tags())

	// Test that both tag keys are found and applied
	parser = JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{"mytag", "othertag"},
	}
	metric, err = parser.ParseLine(validJSONTags)
	require.NoError(t, err)
	require.Equal(t, "json_test", metric.Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metric.Fields())
	require.Equal(t, map[string]string{
		"mytag":    "foobar",
		"othertag": "baz",
	}, metric.Tags())
}

func TestParseValidJSONDefaultTags(t *testing.T) {
	parser := JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{"mytag"},
		DefaultTags: map[string]string{
			"t4g": "default",
		},
	}

	// Most basic vanilla test
	metrics, err := parser.Parse([]byte(validJSON))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "json_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{"t4g": "default"}, metrics[0].Tags())

	// Test that tagkeys and default tags are applied
	metrics, err = parser.Parse([]byte(validJSONTags))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "json_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{
		"t4g":   "default",
		"mytag": "foobar",
	}, metrics[0].Tags())
}

// Test that default tags are overridden by tag keys
func TestParseValidJSONDefaultTagsOverride(t *testing.T) {
	parser := JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{"mytag"},
		DefaultTags: map[string]string{
			"mytag": "default",
		},
	}

	// Most basic vanilla test
	metrics, err := parser.Parse([]byte(validJSON))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "json_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{"mytag": "default"}, metrics[0].Tags())

	// Test that tagkeys override default tags
	metrics, err = parser.Parse([]byte(validJSONTags))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "json_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{
		"mytag": "foobar",
	}, metrics[0].Tags())
}

// Test that json arrays can be parsed
func TestParseValidJSONArray(t *testing.T) {
	parser := JSONParser{
		MetricName: "json_array_test",
	}

	// Most basic vanilla test
	metrics, err := parser.Parse([]byte(validJSONArray))
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	require.Equal(t, "json_array_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{}, metrics[0].Tags())

	// Basic multiple datapoints
	metrics, err = parser.Parse([]byte(validJSONArrayMultiple))
	require.NoError(t, err)
	require.Len(t, metrics, 2)
	require.Equal(t, "json_array_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{}, metrics[1].Tags())
	require.Equal(t, "json_array_test", metrics[1].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(7),
		"b_c": float64(8),
	}, metrics[1].Fields())
	require.Equal(t, map[string]string{}, metrics[1].Tags())
}

func TestParseArrayWithTagKeys(t *testing.T) {
	// Test that strings not matching tag keys are ignored
	parser := JSONParser{
		MetricName: "json_array_test",
		TagKeys:    []string{"wrongtagkey"},
	}
	metrics, err := parser.Parse([]byte(validJSONArrayTags))
	require.NoError(t, err)
	require.Len(t, metrics, 2)
	require.Equal(t, "json_array_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{}, metrics[0].Tags())

	require.Equal(t, "json_array_test", metrics[1].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(7),
		"b_c": float64(8),
	}, metrics[1].Fields())
	require.Equal(t, map[string]string{}, metrics[1].Tags())

	// Test that single tag key is found and applied
	parser = JSONParser{
		MetricName: "json_array_test",
		TagKeys:    []string{"mytag"},
	}
	metrics, err = parser.Parse([]byte(validJSONArrayTags))
	require.NoError(t, err)
	require.Len(t, metrics, 2)
	require.Equal(t, "json_array_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{
		"mytag": "foo",
	}, metrics[0].Tags())

	require.Equal(t, "json_array_test", metrics[1].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(7),
		"b_c": float64(8),
	}, metrics[1].Fields())
	require.Equal(t, map[string]string{
		"mytag": "bar",
	}, metrics[1].Tags())

	// Test that both tag keys are found and applied
	parser = JSONParser{
		MetricName: "json_array_test",
		TagKeys:    []string{"mytag", "othertag"},
	}
	metrics, err = parser.Parse([]byte(validJSONArrayTags))
	require.NoError(t, err)
	require.Len(t, metrics, 2)
	require.Equal(t, "json_array_test", metrics[0].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metrics[0].Fields())
	require.Equal(t, map[string]string{
		"mytag":    "foo",
		"othertag": "baz",
	}, metrics[0].Tags())

	require.Equal(t, "json_array_test", metrics[1].Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(7),
		"b_c": float64(8),
	}, metrics[1].Fields())
	require.Equal(t, map[string]string{
		"mytag":    "bar",
		"othertag": "baz",
	}, metrics[1].Tags())
}

var jsonBOM = []byte("\xef\xbb\xbf[{\"value\":17}]")

func TestHttpJsonBOM(t *testing.T) {
	parser := JSONParser{
		MetricName: "json_test",
	}

	// Most basic vanilla test
	_, err := parser.Parse(jsonBOM)
	require.NoError(t, err)
}

//for testing issue #4260
func TestJSONParseNestedArray(t *testing.T) {
	testString := `{
	"total_devices": 5,
	"total_threads": 10,
	"shares": {
		"total": 5,
		"accepted": 5,
		"rejected": 0,
		"avg_find_time": 4,
		"tester": "work",
		"tester2": "don't want this",
		"tester3": {
			"hello":"sup",
			"fun":"money",
			"break":9
		}
	}
	}`

	parser := JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{"total_devices", "total_threads", "shares_tester3_fun"},
	}

	metrics, err := parser.Parse([]byte(testString))
	log.Printf("m[0] name: %v, tags: %v, fields: %v", metrics[0].Name(), metrics[0].Tags(), metrics[0].Fields())
	require.NoError(t, err)
	require.Equal(t, len(parser.TagKeys), len(metrics[0].Tags()))
}

func TestJSONQueryErrorOnArray(t *testing.T) {
	testString := `{
		"total_devices": 5,
		"total_threads": 10,
		"shares": {
			"total": 5,
			"accepted": 6,
			"test_string": "don't want this",
			"test_obj": {
				"hello":"sup",
				"fun":"money",
				"break":9
			},
			"myArr":[4,5,6]
		}
	}`

	parser := JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{},
		JSONQuery:  "shares.myArr",
	}

	_, err := parser.Parse([]byte(testString))
	require.Error(t, err)
}

func TestArrayOfObjects(t *testing.T) {
	testString := `{
		"meta": {
			"info":9,
			"shares": [{
				"channel": 6,
				"time": 1130,
				"ice":"man"
			},
			{
				"channel": 5,
				"time": 1030,
				"ice":"bucket"
			},
			{
				"channel": 10,
				"time": 330,
				"ice":"cream"
			}]
		},
		"more_stuff":"junk"
	}`

	parser := JSONParser{
		MetricName: "json_test",
		TagKeys:    []string{"ice"},
		JSONQuery:  "meta.shares",
	}

	metrics, err := parser.Parse([]byte(testString))
	require.NoError(t, err)
	require.Equal(t, 3, len(metrics))
}

func TestUseCaseJSONQuery(t *testing.T) {
	testString := `{
		"obj": {
			"name": {"first": "Tom", "last": "Anderson"},
			"age":37,
			"children": ["Sara","Alex","Jack"],
			"fav.movie": "Deer Hunter",
			"friends": [
				{"first": "Dale", "last": "Murphy", "age": 44},
				{"first": "Roger", "last": "Craig", "age": 68},
				{"first": "Jane", "last": "Murphy", "age": 47}
			]
		}
	}`

	parser := JSONParser{
		MetricName:   "json_test",
		StringFields: []string{"last"},
		TagKeys:      []string{"first"},
		JSONQuery:    "obj.friends",
	}

	metrics, err := parser.Parse([]byte(testString))
	require.NoError(t, err)
	require.Equal(t, 3, len(metrics))
	require.Equal(t, metrics[0].Fields()["last"], "Murphy")
}

func TestTimeParser(t *testing.T) {
	testString := `[
		{
			"a": 5,
			"b": {
				"c": 6,
				"time":"04 Jan 06 15:04 MST"
			},
			"my_tag_1": "foo",
			"my_tag_2": "baz"
		},
		{
			"a": 7,
			"b": {
				"c": 8,
				"time":"11 Jan 07 15:04 MST"
			},
			"my_tag_1": "bar",
			"my_tag_2": "baz"
		}
	]`

	parser := JSONParser{
		MetricName:     "json_test",
		JSONTimeKey:    "b_time",
		JSONTimeFormat: "02 Jan 06 15:04 MST",
	}
	metrics, err := parser.Parse([]byte(testString))
	require.NoError(t, err)
	require.Equal(t, 2, len(metrics))
	require.Equal(t, false, metrics[0].Time() == metrics[1].Time())
}

func TestTimeParserWithTimezone(t *testing.T) {
	testString := `{
		"time": "04 Jan 06 15:04"
	}`

	parser := JSONParser{
		MetricName:     "json_test",
		JSONTimeKey:    "time",
		JSONTimeFormat: "02 Jan 06 15:04",
		JSONTimezone:   "America/New_York",
	}
	metrics, err := parser.Parse([]byte(testString))
	require.NoError(t, err)
	require.Equal(t, 1, len(metrics))
	require.EqualValues(t, int64(1136405040000000000), metrics[0].Time().UnixNano())
}

func TestUnixTimeParser(t *testing.T) {
	testString := `[
		{
			"a": 5,
			"b": {
				"c": 6,
				"time": "1536001411.1234567890"
			},
			"my_tag_1": "foo",
			"my_tag_2": "baz"
		},
		{
			"a": 7,
			"b": {
				"c": 8,
				"time": 1536002769.123
			},
			"my_tag_1": "bar",
			"my_tag_2": "baz"
		}
	]`

	parser := JSONParser{
		MetricName:     "json_test",
		JSONTimeKey:    "b_time",
		JSONTimeFormat: "unix",
	}
	metrics, err := parser.Parse([]byte(testString))
	require.NoError(t, err)
	require.Equal(t, 2, len(metrics))
	require.Equal(t, false, metrics[0].Time() == metrics[1].Time())
}

func TestUnixMsTimeParser(t *testing.T) {
	testString := `[
		{
			"a": 5,
			"b": {
				"c": 6,
				"time": "1536001411100"
			},
			"my_tag_1": "foo",
			"my_tag_2": "baz"
		},
		{
			"a": 7,
			"b": {
				"c": 8,
				"time": 1536002769123
			},
			"my_tag_1": "bar",
			"my_tag_2": "baz"
		}
	]`

	parser := JSONParser{
		MetricName:     "json_test",
		JSONTimeKey:    "b_time",
		JSONTimeFormat: "unix_ms",
	}
	metrics, err := parser.Parse([]byte(testString))
	require.NoError(t, err)
	require.Equal(t, 2, len(metrics))
	require.Equal(t, false, metrics[0].Time() == metrics[1].Time())
}

func TestTimeErrors(t *testing.T) {
	testString := `{
		"a": 5,
		"b": {
			"c": 6,
			"time":"04 Jan 06 15:04 MST"
		},
		"my_tag_1": "foo",
		"my_tag_2": "baz"
	}`

	parser := JSONParser{
		MetricName:     "json_test",
		JSONTimeKey:    "b_time",
		JSONTimeFormat: "02 January 06 15:04 MST",
	}

	metrics, err := parser.Parse([]byte(testString))
	require.Error(t, err)
	require.Equal(t, 0, len(metrics))

	testString2 := `{
		"a": 5,
		"b": {
			"c": 6
		},
		"my_tag_1": "foo",
		"my_tag_2": "baz"
	}`

	parser = JSONParser{
		MetricName:     "json_test",
		JSONTimeKey:    "b_time",
		JSONTimeFormat: "02 January 06 15:04 MST",
	}

	metrics, err = parser.Parse([]byte(testString2))
	log.Printf("err: %v", err)
	require.Error(t, err)
	require.Equal(t, 0, len(metrics))
	require.Equal(t, fmt.Errorf("JSON time key could not be found"), err)
}

func TestNameKey(t *testing.T) {
	testString := `{
		"a": 5,
		"b": {
			"c": "this is my name",
			"time":"04 Jan 06 15:04 MST"
		},
		"my_tag_1": "foo",
		"my_tag_2": "baz"
	}`

	parser := JSONParser{
		JSONNameKey: "b_c",
	}

	metrics, err := parser.Parse([]byte(testString))
	require.NoError(t, err)
	require.Equal(t, "this is my name", metrics[0].Name())
}

func TestTimeKeyDelete(t *testing.T) {
	data := `{
		"timestamp": 1541183052,
		"value": 42
	}`

	parser := JSONParser{
		MetricName:     "json",
		JSONTimeKey:    "timestamp",
		JSONTimeFormat: "unix",
	}

	metrics, err := parser.Parse([]byte(data))
	require.NoError(t, err)
	expected := []telegraf.Metric{
		testutil.MustMetric("json",
			map[string]string{},
			map[string]interface{}{"value": 42.0},
			time.Unix(1541183052, 0)),
	}

	testutil.RequireMetricsEqual(t, expected, metrics)
}
