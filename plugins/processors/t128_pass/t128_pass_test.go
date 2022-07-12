package t128_pass

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/influxdata/toml"
	"github.com/stretchr/testify/assert"
)

var t0 = time.Unix(0, 0)

func newMetric(name string, tags map[string]string, fields map[string]interface{}) telegraf.Metric {
	if tags == nil {
		tags = map[string]string{}
	}
	if fields == nil {
		fields = map[string]interface{}{}
	}

	return metric.New(name, tags, fields, t0)
}

func passedMetric(name string, tags map[string]string, fields map[string]interface{}) Metric {
	return Metric{Metric: newMetric(name, tags, combineFields(map[string]interface{}{"at-least": "one-field"}, fields)), Dropped: false}
}

func droppedMetric(name string, tags map[string]string, fields map[string]interface{}) Metric {
	return Metric{Metric: newMetric(name, tags, combineFields(map[string]interface{}{"at-least": "one-field"}, fields)), Dropped: true}
}

func combineFields(a, b map[string]interface{}) map[string]interface{} {
	for k, v := range b {
		a[k] = v
	}

	return a
}

type Metric struct {
	telegraf.Metric
	Dropped bool
}

func TestPass(t *testing.T) {
	testCases := []struct {
		Name         string
		Conditions   []Condition
		InputMetrics []Metric
	}{
		{
			Name:         "drops with no conditions",
			Conditions:   []Condition{},
			InputMetrics: []Metric{droppedMetric("some-measurement", nil, nil)},
		},
		{
			Name:       "drops",
			Conditions: []Condition{{Tags: leaves{"tag1": {"value1"}}}},
			InputMetrics: []Metric{
				passedMetric("some-measurement", map[string]string{"tag1": "value1"}, nil),
				droppedMetric("some-measurement", map[string]string{"tag1": "value2"}, nil),
			},
		},
		{
			Name:       "drops if no tag",
			Conditions: []Condition{{Tags: leaves{"tag1": {"value1"}}}},
			InputMetrics: []Metric{
				droppedMetric("some-measurement", nil, nil),
			},
		},
		{
			Name:       "drops if no tag overriden by config",
			Conditions: []Condition{{IgnoreMissingKeys: true, Tags: leaves{"tag1": {"value1"}}}},
			InputMetrics: []Metric{
				passedMetric("some-measurement", nil, nil),
			},
		},
		{
			Name:       "ors conditions together",
			Conditions: []Condition{{Tags: leaves{"tag1": {"value1"}}}, {Tags: leaves{"tag1": {"value2"}}}},
			InputMetrics: []Metric{
				passedMetric("some-measurement", map[string]string{"tag1": "value1"}, nil),
				passedMetric("some-measurement", map[string]string{"tag1": "value2"}, nil),
				droppedMetric("some-measurement", map[string]string{"tag1": "value3"}, nil),
			},
		},
		{
			Name:       "ands tags together by default",
			Conditions: []Condition{{Tags: leaves{"tag1": {"value1"}, "tag2": {"value2"}}}},
			InputMetrics: []Metric{
				passedMetric("some-measurement", map[string]string{"tag1": "value1", "tag2": "value2"}, nil),
				droppedMetric("some-measurement", map[string]string{"tag1": "value1", "tag2": "value1"}, nil),
				droppedMetric("some-measurement", map[string]string{"tag1": "value2", "tag2": "value2"}, nil),
			},
		},
		{
			Name:       "or operation ors tags together",
			Conditions: []Condition{{Operation: orOperation, Tags: leaves{"tag1": {"value1"}, "tag2": {"value2"}}}},
			InputMetrics: []Metric{
				passedMetric("some-measurement", map[string]string{"tag1": "value1", "tag2": "value2"}, nil),
				passedMetric("some-measurement", map[string]string{"tag1": "value1", "tag2": "value1"}, nil),
				passedMetric("some-measurement", map[string]string{"tag1": "value2", "tag2": "value2"}, nil),
				droppedMetric("some-measurement", map[string]string{"tag1": "value2", "tag2": "value1"}, nil),
			},
		},
		{
			Name:       "ors multiple values",
			Conditions: []Condition{{Tags: leaves{"tag1": {"value1", "value2"}}}},
			InputMetrics: []Metric{
				passedMetric("some-measurement", map[string]string{"tag1": "value1"}, nil),
				passedMetric("some-measurement", map[string]string{"tag1": "value2"}, nil),
				droppedMetric("some-measurement", map[string]string{"tag1": "value3"}, nil),
			},
		},
		{
			Name:       "regex matches whole tag values",
			Conditions: []Condition{{Mode: regexMode, Tags: leaves{"tag1": {"234.*"}}}},
			InputMetrics: []Metric{
				droppedMetric("some-measurement", map[string]string{"tag1": "12345"}, nil),
				droppedMetric("some-measurement", map[string]string{"tag1": "something-else"}, nil),
				passedMetric("some-measurement", map[string]string{"tag1": "23456"}, nil),
				passedMetric("some-measurement", map[string]string{"tag1": "234"}, nil),
			},
		},
		{
			Name:       "glob matches whole tag values",
			Conditions: []Condition{{Mode: globMode, Tags: leaves{"tag1": {"234*"}}}},
			InputMetrics: []Metric{
				droppedMetric("some-measurement", map[string]string{"tag1": "12345"}, nil),
				droppedMetric("some-measurement", map[string]string{"tag1": "something-else"}, nil),
				passedMetric("some-measurement", map[string]string{"tag1": "23456"}, nil),
				passedMetric("some-measurement", map[string]string{"tag1": "234"}, nil),
			},
		},
		{
			Name:       "inverts",
			Conditions: []Condition{{Invert: true, Tags: leaves{"tag1": {"value1"}}}},
			InputMetrics: []Metric{
				droppedMetric("some-measurement", map[string]string{"tag1": "value1"}, nil),
				passedMetric("some-measurement", map[string]string{"tag1": "value2"}, nil),
			},
		},
		{
			Name:       "fields by themselves",
			Conditions: []Condition{{Fields: fieldTypes{String: leaves{"field1": {"value1"}}}}},
			InputMetrics: []Metric{
				passedMetric("some-measurement", nil, map[string]interface{}{"field1": "value1"}),
				droppedMetric("some-measurement", nil, map[string]interface{}{"field1": "value2"}),
			},
		},
		{
			Name:       "field applies mode",
			Conditions: []Condition{{Mode: globMode, Fields: fieldTypes{String: leaves{"field1": {"value*"}}}}},
			InputMetrics: []Metric{
				passedMetric("some-measurement", nil, map[string]interface{}{"field1": "value1"}),
				passedMetric("some-measurement", nil, map[string]interface{}{"field1": "value2"}),
				droppedMetric("some-measurement", nil, map[string]interface{}{"field2": "value1"}),
			},
		},
		{
			Name:       "converts fields to strings for comparison",
			Conditions: []Condition{{Fields: fieldTypes{String: leaves{"field1": {"12"}}}}},
			InputMetrics: []Metric{
				passedMetric("some-measurement", nil, map[string]interface{}{"field1": 12}),
				droppedMetric("some-measurement", nil, map[string]interface{}{"field1": 23}),
			},
		},
		{
			Name:       "field ands with tags",
			Conditions: []Condition{{Tags: leaves{"tag1": {"value1"}}, Fields: fieldTypes{String: leaves{"field1": {"value2"}}}}},
			InputMetrics: []Metric{
				passedMetric("some-measurement", map[string]string{"tag1": "value1"}, map[string]interface{}{"field1": "value2"}),
				droppedMetric("some-measurement", map[string]string{"tag1": "value2"}, map[string]interface{}{"field1": "value2"}),
				droppedMetric("some-measurement", map[string]string{"tag1": "value1"}, map[string]interface{}{"field1": "value1"}),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			assert.True(t, len(testCase.InputMetrics) > 0, "need at least one metric to process")

			r := newPass()
			r.Conditions = testCase.Conditions
			r.log = testutil.Logger{}
			assert.Nil(t, r.Init())

			inputMetrics := make([]telegraf.Metric, len(testCase.InputMetrics))
			for i, metric := range testCase.InputMetrics {
				inputMetrics[i] = metric.Metric
			}

			result := r.Apply(inputMetrics...)

			assert.NotNil(t, result)

			for i, metric := range result {
				actuallyDropped := len(metric.FieldList()) == 0
				assert.Equal(t, testCase.InputMetrics[i].Dropped, actuallyDropped)
			}
		})
	}
}

func TestValidation(t *testing.T) {
	testCases := []struct {
		Name       string
		Conditions []Condition
	}{
		{
			Name:       "needs valid tag regex",
			Conditions: []Condition{{Mode: regexMode, Tags: leaves{"tag1": {"invalid(regex"}}}},
		},
		{
			Name:       "needs valid tag glob",
			Conditions: []Condition{{Mode: globMode, Tags: leaves{"tag1": {"invalid[glob"}}}},
		},
		{
			Name:       "needs valid field regex",
			Conditions: []Condition{{Mode: regexMode, Fields: fieldTypes{String: leaves{"tag1": {"invalid(regex"}}}}},
		},
		{
			Name:       "needs valid field glob",
			Conditions: []Condition{{Mode: globMode, Fields: fieldTypes{String: leaves{"tag1": {"invalid[glob"}}}}},
		},
		{
			Name:       "invalid mode",
			Conditions: []Condition{{Mode: "some-invalid-mode", Tags: leaves{"tag1": {"just needed a tag"}}}},
		},
		{
			Name:       "invalid operation",
			Conditions: []Condition{{Operation: "some-invalid-operation", Tags: leaves{"tag1": {"just needed a tag"}}}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r := newPass()
			r.Conditions = testCase.Conditions
			r.log = testutil.Logger{}
			assert.NotNil(t, r.Init())
		})
	}
}

func TestLoadsFromToml(t *testing.T) {

	plugin := &T128Pass{}
	exampleConfig := []byte(`
		[[condition]]
		  mode = "glob"
		  operation = "or"
		  invert = true
		  ignore_missing_keys = true

		[condition.tags]
		  tag1 = ["value1", "value2"]

		[condition.fields.string]
		  field1 = ["value1", "value2"]

		[[condition]]

		[condition.tags]
		  tag1 = ["value3"]
	`)

	assert.NoError(t, toml.Unmarshal(exampleConfig, plugin))
	assert.Equal(t,
		[]Condition{
			{
				Mode:              globMode,
				Operation:         orOperation,
				Invert:            true,
				IgnoreMissingKeys: true,
				Tags:              leaves{"tag1": {"value1", "value2"}},
				Fields:            fieldTypes{leaves{"field1": {"value1", "value2"}}},
			},
			{
				Tags: leaves{"tag1": {"value3"}},
			},
		},
		plugin.Conditions)
}

func TestLoadsFromTomlComplainsAboutDuplicateTags(t *testing.T) {

	plugin := &T128Pass{}
	exampleConfig := []byte(`
		[[condition]]

		[condition.tags]
		  tag1 = ["value1", "value2"]
		  tag1 = ["value3"]
	`)

	assert.Error(t, toml.Unmarshal(exampleConfig, plugin))
}
