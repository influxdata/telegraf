package json

import (
	"fmt"
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
	mixedValidityJSON      = "[{\"a\": 5, \"time\": \"2006-01-02T15:04:05\"}, {\"a\": 2}]"
)

const validJSONTags = `
{
    "a": 5,
    "b": {
      "c": 6
    },
    "mytag": "foobar",
    "othertag": "baz",
    "tags_object": {
        "mytag": "foobar",
        "othertag": "baz"
    }
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
    "othertag": "baz",
    "tags_array": [
        {
        "mytag": "foo"
        },
        {
        "othertag": "baz"
        }
    ],
    "anothert": "foo"
},
{
    "a": 7,
    "b": {
        "c": 8
    },
    "mytag": "bar",
    "othertag": "baz",
    "tags_array": [
    {
    "mytag": "bar"
    },
    {
    "othertag": "baz"
    }
    ],
    "anothert": "bar"
    }
]
`

func TestParseValidJSON(t *testing.T) {
	parser, err := New(&Config{
		MetricName: "json_test",
	})
	require.NoError(t, err)

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
	parser, err := New(&Config{
		MetricName: "json_test",
	})
	require.NoError(t, err)

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
	parser, err := New(&Config{
		MetricName: "json_test",
	})
	require.NoError(t, err)

	_, err = parser.Parse([]byte(invalidJSON))
	require.Error(t, err)
	_, err = parser.Parse([]byte(invalidJSON2))
	require.Error(t, err)
	_, err = parser.ParseLine(invalidJSON)
	require.Error(t, err)
}

func TestParseJSONImplicitStrictness(t *testing.T) {
	parserImplicitNoStrict, err := New(&Config{
		MetricName: "json_test",
		TimeKey:    "time",
	})
	require.NoError(t, err)

	_, err = parserImplicitNoStrict.Parse([]byte(mixedValidityJSON))
	require.NoError(t, err)
}

func TestParseJSONExplicitStrictnessFalse(t *testing.T) {
	parserNoStrict, err := New(&Config{
		MetricName: "json_test",
		TimeKey:    "time",
		Strict:     false,
	})
	require.NoError(t, err)

	_, err = parserNoStrict.Parse([]byte(mixedValidityJSON))
	require.NoError(t, err)
}

func TestParseJSONExplicitStrictnessTrue(t *testing.T) {
	parserStrict, err := New(&Config{
		MetricName: "json_test",
		TimeKey:    "time",
		Strict:     true,
	})
	require.NoError(t, err)

	_, err = parserStrict.Parse([]byte(mixedValidityJSON))
	require.Error(t, err)
}

func TestParseWithTagKeys(t *testing.T) {
	// Test that strings not matching tag keys are ignored
	parser, err := New(&Config{
		MetricName: "json_test",
		TagKeys:    []string{"wrongtagkey"},
	})
	require.NoError(t, err)

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
	parser, err = New(&Config{
		MetricName: "json_test",
		TagKeys:    []string{"mytag"},
	})
	require.NoError(t, err)

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
	parser, err = New(&Config{
		MetricName: "json_test",
		TagKeys:    []string{"mytag", "othertag"},
	})
	require.NoError(t, err)
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
	parser, err := New(&Config{
		MetricName: "json_test",
		TagKeys:    []string{"wrongtagkey"},
	})
	require.NoError(t, err)
	metric, err := parser.ParseLine(validJSONTags)
	require.NoError(t, err)
	require.Equal(t, "json_test", metric.Name())
	require.Equal(t, map[string]interface{}{
		"a":   float64(5),
		"b_c": float64(6),
	}, metric.Fields())
	require.Equal(t, map[string]string{}, metric.Tags())

	// Test that single tag key is found and applied
	parser, err = New(&Config{
		MetricName: "json_test",
		TagKeys:    []string{"mytag"},
	})
	require.NoError(t, err)

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
	parser, err = New(&Config{
		MetricName: "json_test",
		TagKeys:    []string{"mytag", "othertag"},
	})
	require.NoError(t, err)

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
	parser, err := New(&Config{
		MetricName: "json_test",
		TagKeys:    []string{"mytag"},
		DefaultTags: map[string]string{
			"t4g": "default",
		},
	})
	require.NoError(t, err)

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
	parser, err := New(&Config{
		MetricName: "json_test",
		TagKeys:    []string{"mytag"},
		DefaultTags: map[string]string{
			"mytag": "default",
		},
	})
	require.NoError(t, err)

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
	parser, err := New(&Config{
		MetricName: "json_array_test",
	})
	require.NoError(t, err)

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
	parser, err := New(&Config{
		MetricName: "json_array_test",
		TagKeys:    []string{"wrongtagkey"},
	})
	require.NoError(t, err)

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
	parser, err = New(&Config{
		MetricName: "json_array_test",
		TagKeys:    []string{"mytag"},
	})
	require.NoError(t, err)

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
	parser, err = New(&Config{
		MetricName: "json_array_test",
		TagKeys:    []string{"mytag", "othertag"},
	})
	require.NoError(t, err)

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
	parser, err := New(&Config{
		MetricName: "json_test",
	})
	require.NoError(t, err)

	// Most basic vanilla test
	_, err = parser.Parse(jsonBOM)
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

	parser, err := New(&Config{
		MetricName: "json_test",
		TagKeys:    []string{"total_devices", "total_threads", "shares_tester3_fun"},
	})
	require.NoError(t, err)

	metrics, err := parser.Parse([]byte(testString))
	require.Len(t, metrics, 1)
	require.NoError(t, err)
	require.Equal(t, 3, len(metrics[0].Tags()))
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

	parser, err := New(&Config{
		MetricName: "json_test",
		TagKeys:    []string{},
		Query:      "shares.myArr",
	})
	require.NoError(t, err)

	_, err = parser.Parse([]byte(testString))
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

	parser, err := New(&Config{
		MetricName: "json_test",
		TagKeys:    []string{"ice"},
		Query:      "meta.shares",
	})
	require.NoError(t, err)

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

	parser, err := New(&Config{
		MetricName:   "json_test",
		StringFields: []string{"last"},
		TagKeys:      []string{"first"},
		Query:        "obj.friends",
	})
	require.NoError(t, err)

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

	parser, err := New(&Config{
		MetricName: "json_test",
		TimeKey:    "b_time",
		TimeFormat: "02 Jan 06 15:04 MST",
	})
	require.NoError(t, err)
	metrics, err := parser.Parse([]byte(testString))
	require.NoError(t, err)
	require.Equal(t, 2, len(metrics))
	require.Equal(t, false, metrics[0].Time() == metrics[1].Time())
}

func TestTimeParserWithTimezone(t *testing.T) {
	testString := `{
		"time": "04 Jan 06 15:04"
	}`

	parser, err := New(&Config{
		MetricName: "json_test",
		TimeKey:    "time",
		TimeFormat: "02 Jan 06 15:04",
		Timezone:   "America/New_York",
	})
	require.NoError(t, err)
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

	parser, err := New(&Config{
		MetricName: "json_test",
		TimeKey:    "b_time",
		TimeFormat: "unix",
	})
	require.NoError(t, err)

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

	parser, err := New(&Config{
		MetricName: "json_test",
		TimeKey:    "b_time",
		TimeFormat: "unix_ms",
	})
	require.NoError(t, err)

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

	parser, err := New(&Config{
		MetricName: "json_test",
		TimeKey:    "b_time",
		TimeFormat: "02 January 06 15:04 MST",
	})
	require.NoError(t, err)

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

	parser, err = New(&Config{
		MetricName: "json_test",
		TimeKey:    "b_time",
		TimeFormat: "02 January 06 15:04 MST",
	})
	require.NoError(t, err)

	metrics, err = parser.Parse([]byte(testString2))
	require.Error(t, err)
	require.Equal(t, 0, len(metrics))
	require.Equal(t, fmt.Errorf("JSON time key could not be found"), err)
}

func TestShareTimestamp(t *testing.T) {
	parser, err := New(&Config{
		MetricName: "json_test",
	})
	require.NoError(t, err)

	metrics, err := parser.Parse([]byte(validJSONArrayMultiple))
	require.NoError(t, err)
	require.Equal(t, 2, len(metrics))
	require.Equal(t, true, metrics[0].Time() == metrics[1].Time())
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

	parser, err := New(&Config{
		NameKey: "b_c",
	})
	require.NoError(t, err)

	metrics, err := parser.Parse([]byte(testString))
	require.NoError(t, err)
	require.Equal(t, "this is my name", metrics[0].Name())
}

func TestParseArrayWithWrongType(t *testing.T) {
	data := `[{"answer": 42}, 123]`

	parser, err := New(&Config{})
	require.NoError(t, err)

	_, err = parser.Parse([]byte(data))
	require.Error(t, err)
}

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		input    []byte
		expected []telegraf.Metric
	}{
		{
			name: "tag keys with underscore issue 6705",
			config: &Config{
				MetricName: "json",
				TagKeys:    []string{"metric___name__"},
			},
			input: []byte(`{"metric": {"__name__": "howdy", "time_idle": 42}}`),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"json",
					map[string]string{
						"metric___name__": "howdy",
					},
					map[string]interface{}{
						"metric_time_idle": 42.0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name:     "parse empty array",
			config:   &Config{},
			input:    []byte(`[]`),
			expected: []telegraf.Metric{},
		},
		{
			name:     "parse null",
			config:   &Config{},
			input:    []byte(`null`),
			expected: []telegraf.Metric{},
		},
		{
			name:     "parse null with query",
			config:   &Config{Query: "result.data"},
			input:    []byte(`{"error":null,"result":{"data":null,"items_per_page":10,"total_items":0,"total_pages":0}}`),
			expected: []telegraf.Metric{},
		},
		{
			name: "parse simple array",
			config: &Config{
				MetricName: "json",
			},
			input: []byte(`[{"answer": 42}]`),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"json",
					map[string]string{},
					map[string]interface{}{
						"answer": 42.0,
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "string field glob",
			config: &Config{
				MetricName:   "json",
				StringFields: []string{"*"},
			},
			input: []byte(`
{
    "color": "red",
    "status": "error"
}
`),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"json",
					map[string]string{},
					map[string]interface{}{
						"color":  "red",
						"status": "error",
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "time key is deleted from fields",
			config: &Config{
				MetricName: "json",
				TimeKey:    "timestamp",
				TimeFormat: "unix",
			},
			input: []byte(`
{
	"value": 42,
	"timestamp":  1541183052
}
`),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"json",
					map[string]string{},
					map[string]interface{}{
						"value": 42.0,
					},
					time.Unix(1541183052, 0),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := New(tt.config)
			require.NoError(t, err)

			actual, err := parser.Parse(tt.input)
			require.NoError(t, err)

			testutil.RequireMetricsEqual(t, tt.expected, actual, testutil.IgnoreTime())
		})
	}
}

func TestParseWithWildcardTagKeys(t *testing.T) {
	var tests = []struct {
		name     string
		config   *Config
		input    []byte
		expected []telegraf.Metric
	}{
		{
			name: "wildcard matching with tags nested within object",
			config: &Config{
				MetricName: "json_test",
				TagKeys:    []string{"tags_object_*"},
			},
			input: []byte(validJSONTags),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"json_test",
					map[string]string{
						"tags_object_mytag":    "foobar",
						"tags_object_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(5),
						"b_c": float64(6),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "wildcard matching with keys containing tag",
			config: &Config{
				MetricName: "json_test",
				TagKeys:    []string{"*tag"},
			},
			input: []byte(validJSONTags),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"json_test",
					map[string]string{
						"mytag":                "foobar",
						"othertag":             "baz",
						"tags_object_mytag":    "foobar",
						"tags_object_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(5),
						"b_c": float64(6),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "strings not matching tag keys are still also ignored",
			config: &Config{
				MetricName: "json_test",
				TagKeys:    []string{"wrongtagkey", "tags_object_*"},
			},
			input: []byte(validJSONTags),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"json_test",
					map[string]string{
						"tags_object_mytag":    "foobar",
						"tags_object_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(5),
						"b_c": float64(6),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "single tag key is also found and applied",
			config: &Config{
				MetricName: "json_test",
				TagKeys:    []string{"mytag", "tags_object_*"},
			},
			input: []byte(validJSONTags),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"json_test",
					map[string]string{
						"mytag":                "foobar",
						"tags_object_mytag":    "foobar",
						"tags_object_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(5),
						"b_c": float64(6),
					},
					time.Unix(0, 0),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := New(tt.config)
			require.NoError(t, err)

			actual, err := parser.Parse(tt.input)
			require.NoError(t, err)
			testutil.RequireMetricsEqual(t, tt.expected, actual, testutil.IgnoreTime())
		})
	}
}

func TestParseLineWithWildcardTagKeys(t *testing.T) {
	var tests = []struct {
		name     string
		config   *Config
		input    string
		expected telegraf.Metric
	}{
		{
			name: "wildcard matching with tags nested within object",
			config: &Config{
				MetricName: "json_test",
				TagKeys:    []string{"tags_object_*"},
			},
			input: validJSONTags,
			expected: testutil.MustMetric(
				"json_test",
				map[string]string{
					"tags_object_mytag":    "foobar",
					"tags_object_othertag": "baz",
				},
				map[string]interface{}{
					"a":   float64(5),
					"b_c": float64(6),
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "wildcard matching with keys containing tag",
			config: &Config{
				MetricName: "json_test",
				TagKeys:    []string{"*tag"},
			},
			input: validJSONTags,
			expected: testutil.MustMetric(
				"json_test",
				map[string]string{
					"mytag":                "foobar",
					"othertag":             "baz",
					"tags_object_mytag":    "foobar",
					"tags_object_othertag": "baz",
				},
				map[string]interface{}{
					"a":   float64(5),
					"b_c": float64(6),
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "strings not matching tag keys are ignored",
			config: &Config{
				MetricName: "json_test",
				TagKeys:    []string{"wrongtagkey", "tags_object_*"},
			},
			input: validJSONTags,
			expected: testutil.MustMetric(
				"json_test",
				map[string]string{
					"tags_object_mytag":    "foobar",
					"tags_object_othertag": "baz",
				},
				map[string]interface{}{
					"a":   float64(5),
					"b_c": float64(6),
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "single tag key is also found and applied",
			config: &Config{
				MetricName: "json_test",
				TagKeys:    []string{"mytag", "tags_object_*"},
			},
			input: validJSONTags,
			expected: testutil.MustMetric(
				"json_test",
				map[string]string{
					"mytag":                "foobar",
					"tags_object_mytag":    "foobar",
					"tags_object_othertag": "baz",
				},
				map[string]interface{}{
					"a":   float64(5),
					"b_c": float64(6),
				},
				time.Unix(0, 0),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := New(tt.config)
			require.NoError(t, err)

			actual, err := parser.ParseLine(tt.input)
			require.NoError(t, err)

			testutil.RequireMetricEqual(t, tt.expected, actual, testutil.IgnoreTime())
		})
	}
}

func TestParseArrayWithWildcardTagKeys(t *testing.T) {
	var tests = []struct {
		name     string
		config   *Config
		input    []byte
		expected []telegraf.Metric
	}{
		{
			name: "wildcard matching with keys containing tag within array works",
			config: &Config{
				MetricName: "json_array_test",
				TagKeys:    []string{"*tag"},
			},
			input: []byte(validJSONArrayTags),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"json_array_test",
					map[string]string{
						"mytag":                 "foo",
						"othertag":              "baz",
						"tags_array_0_mytag":    "foo",
						"tags_array_1_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(5),
						"b_c": float64(6),
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"json_array_test",
					map[string]string{
						"mytag":                 "bar",
						"othertag":              "baz",
						"tags_array_0_mytag":    "bar",
						"tags_array_1_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(7),
						"b_c": float64(8),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: " wildcard matching with tags nested array within object works",
			config: &Config{
				MetricName: "json_array_test",
				TagKeys:    []string{"tags_array_*"},
			},
			input: []byte(validJSONArrayTags),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"json_array_test",
					map[string]string{
						"tags_array_0_mytag":    "foo",
						"tags_array_1_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(5),
						"b_c": float64(6),
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"json_array_test",
					map[string]string{
						"tags_array_0_mytag":    "bar",
						"tags_array_1_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(7),
						"b_c": float64(8),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "strings not matching tag keys are still also ignored",
			config: &Config{
				MetricName: "json_array_test",
				TagKeys:    []string{"mytag", "*tag"},
			},
			input: []byte(validJSONArrayTags),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"json_array_test",
					map[string]string{
						"mytag":                 "foo",
						"othertag":              "baz",
						"tags_array_0_mytag":    "foo",
						"tags_array_1_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(5),
						"b_c": float64(6),
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"json_array_test",
					map[string]string{
						"mytag":                 "bar",
						"othertag":              "baz",
						"tags_array_0_mytag":    "bar",
						"tags_array_1_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(7),
						"b_c": float64(8),
					},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "single tag key is also found and applied",
			config: &Config{
				MetricName: "json_array_test",
				TagKeys:    []string{"anothert", "*tag"},
			},
			input: []byte(validJSONArrayTags),
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"json_array_test",
					map[string]string{
						"anothert":              "foo",
						"mytag":                 "foo",
						"othertag":              "baz",
						"tags_array_0_mytag":    "foo",
						"tags_array_1_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(5),
						"b_c": float64(6),
					},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"json_array_test",
					map[string]string{
						"anothert":              "bar",
						"mytag":                 "bar",
						"othertag":              "baz",
						"tags_array_0_mytag":    "bar",
						"tags_array_1_othertag": "baz",
					},
					map[string]interface{}{
						"a":   float64(7),
						"b_c": float64(8),
					},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := New(tt.config)
			require.NoError(t, err)

			actual, err := parser.Parse(tt.input)
			require.NoError(t, err)

			testutil.RequireMetricsEqual(t, tt.expected, actual, testutil.IgnoreTime())
		})
	}
}
