package t128_filter

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

func TestFilters(t *testing.T) {
	testCases := []struct {
		Name          string
		Conditions    []Condition
		InputMetrics  []telegraf.Metric
		OutputMetrics []telegraf.Metric
	}{
		{
			Name:          "drops with no conditions",
			Conditions:    []Condition{},
			InputMetrics:  []telegraf.Metric{newMetric("some-measurement", nil, nil)},
			OutputMetrics: []telegraf.Metric{},
		},
		{
			Name:       "drops",
			Conditions: []Condition{{Tags: tags{"tag1": {"value1"}}}},
			InputMetrics: []telegraf.Metric{
				newMetric("some-measurement", map[string]string{"tag1": "value1"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "value2"}, nil),
			},
			OutputMetrics: []telegraf.Metric{newMetric("some-measurement", map[string]string{"tag1": "value1"}, nil)},
		},
		{
			Name:       "drops if no tag",
			Conditions: []Condition{{Tags: tags{"tag1": {"value1"}}}},
			InputMetrics: []telegraf.Metric{
				newMetric("some-measurement", nil, nil),
			},
			OutputMetrics: []telegraf.Metric{},
		},
		{
			Name:       "ors conditions together",
			Conditions: []Condition{{Tags: tags{"tag1": {"value1"}}}, {Tags: tags{"tag1": {"value2"}}}},
			InputMetrics: []telegraf.Metric{
				newMetric("some-measurement", map[string]string{"tag1": "value1"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "value2"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "value3"}, nil),
			},
			OutputMetrics: []telegraf.Metric{
				newMetric("some-measurement", map[string]string{"tag1": "value1"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "value2"}, nil),
			},
		},
		{
			Name:       "ands tags together by default",
			Conditions: []Condition{{Tags: tags{"tag1": {"value1"}, "tag2": {"value2"}}}},
			InputMetrics: []telegraf.Metric{
				newMetric("some-measurement", map[string]string{"tag1": "value1", "tag2": "value2"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "value1", "tag2": "value1"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "value2", "tag2": "value2"}, nil),
			},
			OutputMetrics: []telegraf.Metric{newMetric("some-measurement", map[string]string{"tag1": "value1", "tag2": "value2"}, nil)},
		},
		{
			Name:       "or operation ors tags together",
			Conditions: []Condition{{Operation: orOperation, Tags: tags{"tag1": {"value1"}, "tag2": {"value2"}}}},
			InputMetrics: []telegraf.Metric{
				newMetric("some-measurement", map[string]string{"tag1": "value1", "tag2": "value2"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "value1", "tag2": "value1"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "value2", "tag2": "value2"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "value2", "tag2": "value1"}, nil),
			},
			OutputMetrics: []telegraf.Metric{
				newMetric("some-measurement", map[string]string{"tag1": "value1", "tag2": "value2"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "value1", "tag2": "value1"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "value2", "tag2": "value2"}, nil),
			},
		},
		{
			Name:       "ors multiple values",
			Conditions: []Condition{{Tags: tags{"tag1": {"value1", "value2"}}}},
			InputMetrics: []telegraf.Metric{
				newMetric("some-measurement", map[string]string{"tag1": "value1"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "value2"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "value3"}, nil),
			},
			OutputMetrics: []telegraf.Metric{
				newMetric("some-measurement", map[string]string{"tag1": "value1"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "value2"}, nil),
			},
		},
		{
			Name:       "regex matches whole tag values",
			Conditions: []Condition{{Mode: regexMode, Tags: tags{"tag1": {"234.*"}}}},
			InputMetrics: []telegraf.Metric{
				newMetric("some-measurement", map[string]string{"tag1": "12345"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "something-else"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "23456"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "234"}, nil),
			},
			OutputMetrics: []telegraf.Metric{
				newMetric("some-measurement", map[string]string{"tag1": "23456"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "234"}, nil),
			},
		},
		{
			Name:       "glob matches whole tag values",
			Conditions: []Condition{{Mode: globMode, Tags: tags{"tag1": {"234*"}}}},
			InputMetrics: []telegraf.Metric{
				newMetric("some-measurement", map[string]string{"tag1": "12345"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "something-else"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "23456"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "234"}, nil),
			},
			OutputMetrics: []telegraf.Metric{
				newMetric("some-measurement", map[string]string{"tag1": "23456"}, nil),
				newMetric("some-measurement", map[string]string{"tag1": "234"}, nil),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			assert.True(t, len(testCase.InputMetrics) > 0, "need at least one metric to process")

			r := newFilter()
			r.Conditions = testCase.Conditions
			r.log = testutil.Logger{}
			assert.Nil(t, r.Init())

			result := r.Apply(testCase.InputMetrics...)

			assert.Equal(t, testCase.OutputMetrics, result)
		})
	}
}

func TestValidation(t *testing.T) {
	testCases := []struct {
		Name       string
		Conditions []Condition
	}{
		{
			Name:       "needs valid regex",
			Conditions: []Condition{{Mode: regexMode, Tags: tags{"tag1": {"invalid(regex"}}}},
		},
		{
			Name:       "needs valid glob",
			Conditions: []Condition{{Mode: globMode, Tags: tags{"tag1": {"invalid[glob"}}}},
		},
		{
			Name:       "invalid mode",
			Conditions: []Condition{{Mode: "some-invalid-mode", Tags: tags{"tag1": {"just needed a tag"}}}},
		},
		{
			Name:       "invalid operation",
			Conditions: []Condition{{Operation: "some-invalid-operation", Tags: tags{"tag1": {"just needed a tag"}}}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r := newFilter()
			r.Conditions = testCase.Conditions
			r.log = testutil.Logger{}
			assert.NotNil(t, r.Init())
		})
	}
}

func TestLoadsFromToml(t *testing.T) {

	plugin := &T128Filter{}
	exampleConfig := []byte(`
		[[condition]]
		  mode = "glob"
		  operation = "or"

		[condition.tags]
		  tag1 = ["value1", "value2"]

		[[condition]]

		[condition.tags]
		  tag1 = ["value3"]
	`)

	assert.NoError(t, toml.Unmarshal(exampleConfig, plugin))
	assert.Equal(t, []Condition{{Mode: globMode, Operation: orOperation, Tags: tags{"tag1": {"value1", "value2"}}}, {Tags: tags{"tag1": {"value3"}}}}, plugin.Conditions)
}

func TestLoadsFromTomlComplainsAboutDuplicateTags(t *testing.T) {

	plugin := &T128Filter{}
	exampleConfig := []byte(`
		[[condition]]

		[condition.tags]
		  tag1 = ["value1", "value2"]
		  tag1 = ["value3"]
	`)

	assert.Error(t, toml.Unmarshal(exampleConfig, plugin))
}
