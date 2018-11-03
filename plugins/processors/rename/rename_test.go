package rename

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func newMetric(name string, tags map[string]string, fields map[string]interface{}) telegraf.Metric {
	if tags == nil {
		tags = map[string]string{}
	}
	if fields == nil {
		fields = map[string]interface{}{}
	}
	m, _ := metric.New(name, tags, fields, time.Now())
	return m
}

func TestMeasurementRename(t *testing.T) {
	r := Rename{
		Replaces: []Replace{
			{Measurement: "foo", Dest: "bar"},
			{Measurement: "baz", Dest: "quux"},
		},
	}
	m1 := newMetric("foo", nil, nil)
	m2 := newMetric("bar", nil, nil)
	m3 := newMetric("baz", nil, nil)
	results := r.Apply(m1, m2, m3)
	assert.Equal(t, "bar", results[0].Name(), "Should change name from 'foo' to 'bar'")
	assert.Equal(t, "bar", results[1].Name(), "Should not name from 'bar'")
	assert.Equal(t, "quux", results[2].Name(), "Should change name from 'baz' to 'quux'")
}

func TestTagRename(t *testing.T) {
	r := Rename{
		Replaces: []Replace{
			{Tag: "hostname", Dest: "host"},
		},
	}
	m := newMetric("foo", map[string]string{"hostname": "localhost", "region": "east-1"}, nil)
	results := r.Apply(m)

	assert.Equal(t, map[string]string{"host": "localhost", "region": "east-1"}, results[0].Tags(), "should change tag 'hostname' to 'host'")
}

func TestFieldRename(t *testing.T) {
	r := Rename{
		Replaces: []Replace{
			{Field: "time_msec", Dest: "time"},
		},
	}
	m := newMetric("foo", nil, map[string]interface{}{"time_msec": int64(1250), "snakes": true})
	results := r.Apply(m)

	assert.Equal(t, map[string]interface{}{"time": int64(1250), "snakes": true}, results[0].Fields(), "should change field 'time_msec' to 'time'")
}
