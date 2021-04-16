package enum

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestMetric() telegraf.Metric {
	m := metric.New("m1",
		map[string]string{
			"tag":           "tag_value",
			"duplicate_tag": "tag_value",
		},
		map[string]interface{}{
			"string_value":           "test",
			"duplicate_string_value": "test",
			"int_value":              int(200),
			"uint_value":             uint(500),
			"float_value":            float64(3.14),
			"true_value":             true,
		},
		time.Now(),
	)
	return m
}

func calculateProcessedValues(mapper EnumMapper, metric telegraf.Metric) map[string]interface{} {
	processed := mapper.Apply(metric)
	return processed[0].Fields()
}

func calculateProcessedTags(mapper EnumMapper, metric telegraf.Metric) map[string]string {
	processed := mapper.Apply(metric)
	return processed[0].Tags()
}

func assertFieldValue(t *testing.T, expected interface{}, field string, fields map[string]interface{}) {
	value, present := fields[field]
	assert.True(t, present, "value of field '"+field+"' was not present")
	assert.EqualValues(t, expected, value)
}

func assertTagValue(t *testing.T, expected interface{}, tag string, tags map[string]string) {
	value, present := tags[tag]
	assert.True(t, present, "value of tag '"+tag+"' was not present")
	assert.EqualValues(t, expected, value)
}

func TestRetainsMetric(t *testing.T) {
	mapper := EnumMapper{}
	err := mapper.Init()
	require.Nil(t, err)
	source := createTestMetric()

	target := mapper.Apply(source)[0]
	fields := target.Fields()

	assertFieldValue(t, "test", "string_value", fields)
	assertFieldValue(t, 200, "int_value", fields)
	assertFieldValue(t, 500, "uint_value", fields)
	assertFieldValue(t, float64(3.14), "float_value", fields)
	assertFieldValue(t, true, "true_value", fields)
	assert.Equal(t, "m1", target.Name())
	assert.Equal(t, source.Tags(), target.Tags())
	assert.Equal(t, source.Time(), target.Time())
}

func TestMapsSingleStringValueTag(t *testing.T) {
	mapper := EnumMapper{Mappings: []Mapping{{Tag: "tag", ValueMappings: map[string]interface{}{"tag_value": "valuable"}}}}
	err := mapper.Init()
	require.Nil(t, err)
	tags := calculateProcessedTags(mapper, createTestMetric())

	assertTagValue(t, "valuable", "tag", tags)
}

func TestMappings(t *testing.T) {
	mappings := []map[string][]interface{}{
		{
			"field_name":      []interface{}{"string_value"},
			"target_values":   []interface{}{"test", "test", "test", "not_test", "50", "true"},
			"mapped_values":   []interface{}{"test_1", 5, true, "test_1", 10, false},
			"expected_values": []interface{}{"test_1", 5, true, "test", "test", "test"},
		},
		{
			"field_name":     []interface{}{"true_value"},
			"target_value":   []interface{}{"true", "true", "true", "false", "test", "5"},
			"mapped_value":   []interface{}{false, 1, "false", false, false, false},
			"expected_value": []interface{}{false, 1, "false", true, true, true},
		},
		{
			"field_name":     []interface{}{"int_value"},
			"target_value":   []interface{}{"200", "200", "200", "200", "test", "5"},
			"mapped_value":   []interface{}{"http_ok", true, 1, float64(200.001), false, false},
			"expected_value": []interface{}{"http_ok", true, 1, float64(200.001), 200, 200},
		},
		{
			"field_name":     []interface{}{"uint_value"},
			"target_value":   []interface{}{"500", "500", "500", "test", "false", "5"},
			"mapped_value":   []interface{}{"internal_error", 1, false, false, false, false},
			"expected_value": []interface{}{"internal_error", 1, false, 500, 500, 500},
		},
		{
			"field_name":     []interface{}{"float_value"},
			"target_value":   []interface{}{"3.14", "3.14", "3.14", "3.14", "not_float", "5"},
			"mapped_value":   []interface{}{"pi", 1, false, float64(100.2), float64(3.14), "pi"},
			"expected_value": []interface{}{"pi", 1, false, float64(100.2), float64(3.14), float64(3.14)},
		},
	}

	for _, mapping := range mappings {
		fieldName := mapping["field_name"][0].(string)
		for index := range mapping["target_value"] {
			mapper := EnumMapper{Mappings: []Mapping{{Field: fieldName, ValueMappings: map[string]interface{}{mapping["target_value"][index].(string): mapping["mapped_value"][index]}}}}
			err := mapper.Init()
			assert.Nil(t, err)
			fields := calculateProcessedValues(mapper, createTestMetric())
			assertFieldValue(t, mapping["expected_value"][index], fieldName, fields)
		}
	}
}

func TestMapsToDefaultValueOnUnknownSourceValue(t *testing.T) {
	mapper := EnumMapper{Mappings: []Mapping{{Field: "string_value", Default: int64(42), ValueMappings: map[string]interface{}{"other": int64(1)}}}}
	err := mapper.Init()
	require.Nil(t, err)
	fields := calculateProcessedValues(mapper, createTestMetric())

	assertFieldValue(t, 42, "string_value", fields)
}

func TestDoNotMapToDefaultValueKnownSourceValue(t *testing.T) {
	mapper := EnumMapper{Mappings: []Mapping{{Field: "string_value", Default: int64(42), ValueMappings: map[string]interface{}{"test": int64(1)}}}}
	err := mapper.Init()
	require.Nil(t, err)
	fields := calculateProcessedValues(mapper, createTestMetric())

	assertFieldValue(t, 1, "string_value", fields)
}

func TestNoMappingWithoutDefaultOrDefinedMappingValue(t *testing.T) {
	mapper := EnumMapper{Mappings: []Mapping{{Field: "string_value", ValueMappings: map[string]interface{}{"other": int64(1)}}}}
	err := mapper.Init()
	require.Nil(t, err)
	fields := calculateProcessedValues(mapper, createTestMetric())

	assertFieldValue(t, "test", "string_value", fields)
}

func TestWritesToDestination(t *testing.T) {
	mapper := EnumMapper{Mappings: []Mapping{{Field: "string_value", Dest: "string_code", ValueMappings: map[string]interface{}{"test": int64(1)}}}}
	err := mapper.Init()
	require.Nil(t, err)
	fields := calculateProcessedValues(mapper, createTestMetric())

	assertFieldValue(t, "test", "string_value", fields)
	assertFieldValue(t, 1, "string_code", fields)
}

func TestDoNotWriteToDestinationWithoutDefaultOrDefinedMapping(t *testing.T) {
	field := "string_code"
	mapper := EnumMapper{Mappings: []Mapping{{Field: "string_value", Dest: field, ValueMappings: map[string]interface{}{"other": int64(1)}}}}
	err := mapper.Init()
	require.Nil(t, err)
	fields := calculateProcessedValues(mapper, createTestMetric())

	assertFieldValue(t, "test", "string_value", fields)
	_, present := fields[field]
	assert.False(t, present, "value of field '"+field+"' was present")
}

func TestFieldGlobMatching(t *testing.T) {
	mapper := EnumMapper{Mappings: []Mapping{{Field: "*", ValueMappings: map[string]interface{}{"test": "glob"}}}}
	err := mapper.Init()
	require.Nil(t, err)
	fields := calculateProcessedValues(mapper, createTestMetric())

	assertFieldValue(t, "glob", "string_value", fields)
	assertFieldValue(t, "glob", "duplicate_string_value", fields)
}

func TestTagGlobMatching(t *testing.T) {
	mapper := EnumMapper{Mappings: []Mapping{{Tag: "*", ValueMappings: map[string]interface{}{"tag_value": "glob"}}}}
	err := mapper.Init()
	require.Nil(t, err)
	tags := calculateProcessedTags(mapper, createTestMetric())

	assertTagValue(t, "glob", "tag", tags)
}
