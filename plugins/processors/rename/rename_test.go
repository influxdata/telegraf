package rename

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func newMetric(name string, tags map[string]string, fields map[string]interface{}) telegraf.Metric {
	if tags == nil {
		tags = map[string]string{}
	}
	if fields == nil {
		fields = map[string]interface{}{}
	}
	m := metric.New(name, tags, fields, time.Now())
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
	require.Equal(t, "bar", results[0].Name(), "Should change name from 'foo' to 'bar'")
	require.Equal(t, "bar", results[1].Name(), "Should not name from 'bar'")
	require.Equal(t, "quux", results[2].Name(), "Should change name from 'baz' to 'quux'")
}

func TestTagRename(t *testing.T) {
	r := Rename{
		Replaces: []Replace{
			{Tag: "hostname", Dest: "host"},
		},
	}
	m := newMetric("foo", map[string]string{"hostname": "localhost", "region": "east-1"}, nil)
	results := r.Apply(m)

	require.Equal(t, map[string]string{"host": "localhost", "region": "east-1"}, results[0].Tags(), "should change tag 'hostname' to 'host'")
}

func TestFieldRename(t *testing.T) {
	r := Rename{
		Replaces: []Replace{
			{Field: "time_msec", Dest: "time"},
		},
	}
	m := newMetric("foo", nil, map[string]interface{}{"time_msec": int64(1250), "snakes": true})
	results := r.Apply(m)

	require.Equal(t, map[string]interface{}{"time": int64(1250), "snakes": true}, results[0].Fields(), "should change field 'time_msec' to 'time'")
}

func TestTracking(t *testing.T) {
	inputRaw := []telegraf.Metric{
		metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42}, time.Unix(0, 0)),
		metric.New("bar", map[string]string{}, map[string]interface{}{"value": 99}, time.Unix(0, 0)),
		metric.New("baz", map[string]string{}, map[string]interface{}{"value": 11}, time.Unix(0, 0)),
	}

	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, len(inputRaw))
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}

	input := make([]telegraf.Metric, 0, len(inputRaw))
	for _, m := range inputRaw {
		tm, _ := metric.WithTracking(m, notify)
		input = append(input, tm)
	}

	expected := []telegraf.Metric{
		metric.New(
			"foo",
			map[string]string{},
			map[string]interface{}{"new_value": 42},
			time.Unix(0, 0),
		),
		metric.New(
			"bar",
			map[string]string{},
			map[string]interface{}{"new_value": 99},
			time.Unix(0, 0),
		),
		metric.New(
			"baz",
			map[string]string{},
			map[string]interface{}{"new_value": 11},
			time.Unix(0, 0),
		),
	}

	plugin := &Rename{
		Replaces: []Replace{
			{Field: "value", Dest: "new_value"},
		},
	}

	// Process expected metrics and compare with resulting metrics
	actual := plugin.Apply(input...)
	testutil.RequireMetricsEqual(t, expected, actual)

	// Simulate output acknowledging delivery
	for _, m := range actual {
		m.Accept()
	}

	// Check delivery
	require.Eventuallyf(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(input) == len(delivered)
	}, time.Second, 100*time.Millisecond, "%d delivered but %d expected", len(delivered), len(expected))
}
