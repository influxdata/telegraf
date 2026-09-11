package enum

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func createTestMetric() telegraf.Metric {
	m := metric.New("m1",
		map[string]string{
			"tag":           "tag_value",
			"duplicate_tag": "tag_value",
		},
		map[string]any{
			"string_value":           "test",
			"duplicate_string_value": "test",
			"int_value":              200,
			"uint_value":             uint(500),
			"float_value":            float64(3.14),
			"true_value":             true,
		},
		time.Now(),
	)
	return m
}

func calculateProcessedValues(mapper Enum, m telegraf.Metric) map[string]any {
	processed := mapper.Apply(m)
	return processed[0].Fields()
}

func calculateProcessedTags(mapper Enum, m telegraf.Metric) map[string]string {
	processed := mapper.Apply(m)
	return processed[0].Tags()
}

func assertFieldValue(t *testing.T, expected any, field string, fields map[string]any) {
	value, present := fields[field]
	require.True(t, present, "value of field '"+field+"' was not present")
	require.EqualValues(t, expected, value)
}

func assertTagValue(t *testing.T, expected any, tag string, tags map[string]string) {
	value, present := tags[tag]
	require.True(t, present, "value of tag '"+tag+"' was not present")
	require.EqualValues(t, expected, value)
}

func TestRetainsMetric(t *testing.T) {
	mapper := Enum{}
	err := mapper.Init()
	require.NoError(t, err)
	source := createTestMetric()

	target := mapper.Apply(source)[0]
	fields := target.Fields()

	assertFieldValue(t, "test", "string_value", fields)
	assertFieldValue(t, 200, "int_value", fields)
	assertFieldValue(t, 500, "uint_value", fields)
	assertFieldValue(t, float64(3.14), "float_value", fields)
	assertFieldValue(t, true, "true_value", fields)
	require.Equal(t, "m1", target.Name())
	require.Equal(t, source.Tags(), target.Tags())
	require.Equal(t, source.Time(), target.Time())
}

func TestMapsSingleStringValueTag(t *testing.T) {
	mapper := Enum{Mappings: []*mapping{{
		Tags:          []string{"tag"},
		ValueMappings: map[string]any{"tag_value": "valuable"},
	}}}
	err := mapper.Init()
	require.NoError(t, err)
	tags := calculateProcessedTags(mapper, createTestMetric())

	assertTagValue(t, "valuable", "tag", tags)
}

func TestMappings(t *testing.T) {
	mappings := []map[string][]any{
		{
			"field_name":      []any{"string_value"},
			"target_values":   []any{"test", "test", "test", "not_test", "50", "true"},
			"mapped_values":   []any{"test_1", 5, true, "test_1", 10, false},
			"expected_values": []any{"test_1", 5, true, "test", "test", "test"},
		},
		{
			"field_name":     []any{"true_value"},
			"target_value":   []any{"true", "true", "true", "false", "test", "5"},
			"mapped_value":   []any{false, 1, "false", false, false, false},
			"expected_value": []any{false, 1, "false", true, true, true},
		},
		{
			"field_name":     []any{"int_value"},
			"target_value":   []any{"200", "200", "200", "200", "test", "5"},
			"mapped_value":   []any{"http_ok", true, 1, float64(200.001), false, false},
			"expected_value": []any{"http_ok", true, 1, float64(200.001), 200, 200},
		},
		{
			"field_name":     []any{"uint_value"},
			"target_value":   []any{"500", "500", "500", "test", "false", "5"},
			"mapped_value":   []any{"internal_error", 1, false, false, false, false},
			"expected_value": []any{"internal_error", 1, false, 500, 500, 500},
		},
		{
			"field_name":     []any{"float_value"},
			"target_value":   []any{"3.14", "3.14", "3.14", "3.14", "not_float", "5"},
			"mapped_value":   []any{"pi", 1, false, float64(100.2), float64(3.14), "pi"},
			"expected_value": []any{"pi", 1, false, float64(100.2), float64(3.14), float64(3.14)},
		},
	}

	for _, mappingItem := range mappings {
		fieldName := mappingItem["field_name"][0].(string)
		for index := range mappingItem["target_value"] {
			mapper := Enum{
				Mappings: []*mapping{
					{
						Fields: []string{fieldName},
						ValueMappings: map[string]any{
							mappingItem["target_value"][index].(string): mappingItem["mapped_value"][index],
						},
					},
				},
			}
			err := mapper.Init()
			require.NoError(t, err)
			fields := calculateProcessedValues(mapper, createTestMetric())
			assertFieldValue(t, mappingItem["expected_value"][index], fieldName, fields)
		}
	}
}

func TestMapsToDefaultValueOnUnknownSourceValue(t *testing.T) {
	mapper := Enum{Mappings: []*mapping{{
		Fields:        []string{"string_value"},
		Default:       int64(42),
		ValueMappings: map[string]any{"other": int64(1)},
	}}}
	err := mapper.Init()
	require.NoError(t, err)
	fields := calculateProcessedValues(mapper, createTestMetric())

	assertFieldValue(t, 42, "string_value", fields)
}

func TestDoNotMapToDefaultValueKnownSourceValue(t *testing.T) {
	mapper := Enum{Mappings: []*mapping{{
		Fields:        []string{"string_value"},
		Default:       int64(42),
		ValueMappings: map[string]any{"test": int64(1)},
	}}}
	err := mapper.Init()
	require.NoError(t, err)
	fields := calculateProcessedValues(mapper, createTestMetric())

	assertFieldValue(t, 1, "string_value", fields)
}

func TestNoMappingWithoutDefaultOrDefinedMappingValue(t *testing.T) {
	mapper := Enum{Mappings: []*mapping{{
		Fields:        []string{"string_value"},
		ValueMappings: map[string]any{"other": int64(1)},
	}}}
	err := mapper.Init()
	require.NoError(t, err)
	fields := calculateProcessedValues(mapper, createTestMetric())

	assertFieldValue(t, "test", "string_value", fields)
}

func TestWritesToDestination(t *testing.T) {
	mapper := Enum{Mappings: []*mapping{{
		Fields:        []string{"string_value"},
		Dest:          "string_code",
		ValueMappings: map[string]any{"test": int64(1)},
	}}}
	err := mapper.Init()
	require.NoError(t, err)
	fields := calculateProcessedValues(mapper, createTestMetric())

	assertFieldValue(t, "test", "string_value", fields)
	assertFieldValue(t, 1, "string_code", fields)
}

func TestDoNotWriteToDestinationWithoutDefaultOrDefinedMapping(t *testing.T) {
	field := "string_code"
	mapper := Enum{Mappings: []*mapping{{
		Fields:        []string{"string_value"},
		Dest:          field,
		ValueMappings: map[string]any{"other": int64(1)},
	}}}
	err := mapper.Init()
	require.NoError(t, err)
	fields := calculateProcessedValues(mapper, createTestMetric())

	assertFieldValue(t, "test", "string_value", fields)
	_, present := fields[field]
	require.False(t, present, "value of field '"+field+"' was present")
}

func TestMultipleFields(t *testing.T) {
	mapper := Enum{Mappings: []*mapping{{
		Fields:        []string{"string_value", "duplicate_string_value"},
		ValueMappings: map[string]any{"test": "multiple"},
	}}}
	require.NoError(t, mapper.Init())
	fields := calculateProcessedValues(mapper, createTestMetric())

	assertFieldValue(t, "multiple", "string_value", fields)
	assertFieldValue(t, "multiple", "duplicate_string_value", fields)
}

func TestFieldGlobMatching(t *testing.T) {
	mapper := Enum{Mappings: []*mapping{{
		Fields:        []string{"*"},
		ValueMappings: map[string]any{"test": "glob"},
	}}}
	err := mapper.Init()
	require.NoError(t, err)
	fields := calculateProcessedValues(mapper, createTestMetric())

	assertFieldValue(t, "glob", "string_value", fields)
	assertFieldValue(t, "glob", "duplicate_string_value", fields)
}

func TestTagGlobMatching(t *testing.T) {
	mapper := Enum{Mappings: []*mapping{{
		Tags:          []string{"*"},
		ValueMappings: map[string]any{"tag_value": "glob"},
	}}}
	err := mapper.Init()
	require.NoError(t, err)
	tags := calculateProcessedTags(mapper, createTestMetric())

	assertTagValue(t, "glob", "tag", tags)
}

func TestCollidingValueMappings(t *testing.T) {
	mapper := Enum{Mappings: []*mapping{
		{
			Fields:        []string{"status"},
			ValueMappings: map[string]any{"green": 1, "amber": 2, "red": 3},
		},
		{
			Fields:        []string{"status_reverse"},
			ValueMappings: map[string]any{"green": 3, "amber": 2, "red": 1},
		},
	}}
	require.NoError(t, mapper.Init())

	input := metric.New("m1",
		map[string]string{
			"tag": "tag_value",
		},
		map[string]any{
			"status":         "green",
			"status_reverse": "green",
		},
		time.Now(),
	)

	output := mapper.Apply(input)[0]
	fields := output.Fields()
	assertFieldValue(t, int64(1), "status", fields)
	assertFieldValue(t, int64(3), "status_reverse", fields)
}

func TestTracking(t *testing.T) {
	m := createTestMetric()
	var delivered bool
	notify := func(telegraf.DeliveryInfo) {
		delivered = true
	}
	m, _ = metric.WithTracking(m, notify)

	mapper := Enum{Mappings: []*mapping{{
		Tags:          []string{"*"},
		ValueMappings: map[string]any{"tag_value": "glob"},
	}}}
	err := mapper.Init()
	require.NoError(t, err)

	actual := mapper.Apply(m)[0]
	assertTagValue(t, "glob", "tag", actual.Tags())

	actual.Accept()
	require.Eventually(t, func() bool {
		return delivered
	}, time.Second, 100*time.Millisecond, "no metrics delivered")
}
