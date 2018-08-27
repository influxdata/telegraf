package replace

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func newMetric(name string) telegraf.Metric {
	tags := map[string]string{}
	fields := map[string]interface{}{}
	m, _ := metric.New(name, tags, fields, time.Now())
	return m
}

func TestMeasurementReplace(t *testing.T) {
	r := Replace{}
	r.Old = "_"
	r.New = "-"
	metrics := []telegraf.Metric{
		newMetric("foo:some_value:bar"),
		newMetric("average:cpu:usage"),
		newMetric("average_cpu_usage"),
	}

	results := r.Apply(metrics...)
	assert.Equal(t, "foo:some-value:bar", results[0].Name(), "`_` was not changed to `-`")
	assert.Equal(t, "average:cpu:usage", results[1].Name(), "Input name should have been unchanged")
	assert.Equal(t, "average-cpu-usage", results[2].Name(), "All instances of `_` should have been changed to `-`")
}

func TestMeasurementCharDeletion(t *testing.T) {
	r := Replace{}
	r.Old = "foo"
	r.New = ""

	metrics := []telegraf.Metric{
		newMetric("foo:bar:baz"),
		newMetric("foofoofoo"),
		newMetric("barbarbar"),
	}

	results := r.Apply(metrics...)
	assert.Equal(t, ":bar:baz", results[0].Name(), "Should have deleted the initial `foo`")
	assert.Equal(t, "foofoofoo", results[1].Name(), "Should have refused to delete the whole string")
	assert.Equal(t, "barbarbar", results[2].Name(), "Should not have changed the input")
}
