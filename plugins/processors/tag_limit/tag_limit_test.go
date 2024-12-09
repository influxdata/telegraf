package tag_limit

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func MustMetric(name string, tags map[string]string, fields map[string]interface{}, metricTime time.Time) telegraf.Metric {
	if tags == nil {
		tags = map[string]string{}
	}
	if fields == nil {
		fields = map[string]interface{}{}
	}
	m := metric.New(name, tags, fields, metricTime)
	return m
}

func TestUnderLimit(t *testing.T) {
	currentTime := time.Now()

	oneTags := make(map[string]string)
	oneTags["foo"] = "bar"

	tenTags := make(map[string]string)
	tenTags["a"] = "bar"
	tenTags["b"] = "bar"
	tenTags["c"] = "bar"
	tenTags["d"] = "bar"
	tenTags["e"] = "bar"
	tenTags["f"] = "bar"
	tenTags["g"] = "bar"
	tenTags["h"] = "bar"
	tenTags["i"] = "bar"
	tenTags["j"] = "bar"

	tagLimitConfig := TagLimit{
		Limit: 10,
		Keep:  []string{"foo", "bar"},
	}

	m1 := MustMetric("foo", oneTags, nil, currentTime)
	m2 := MustMetric("bar", tenTags, nil, currentTime)
	limitApply := tagLimitConfig.Apply(m1, m2)
	require.Equal(t, oneTags, limitApply[0].Tags(), "one tag")
	require.Equal(t, tenTags, limitApply[1].Tags(), "ten tags")
}

func TestTrim(t *testing.T) {
	currentTime := time.Now()

	threeTags := make(map[string]string)
	threeTags["a"] = "foo"
	threeTags["b"] = "bar"
	threeTags["z"] = "baz"

	tenTags := make(map[string]string)
	tenTags["a"] = "foo"
	tenTags["b"] = "bar"
	tenTags["c"] = "baz"
	tenTags["d"] = "abc"
	tenTags["e"] = "def"
	tenTags["f"] = "ghi"
	tenTags["g"] = "jkl"
	tenTags["h"] = "mno"
	tenTags["i"] = "pqr"
	tenTags["j"] = "stu"

	tagLimitConfig := TagLimit{
		Limit: 3,
		Keep:  []string{"a", "b"},
	}

	m1 := MustMetric("foo", threeTags, nil, currentTime)
	m2 := MustMetric("bar", tenTags, nil, currentTime)
	limitApply := tagLimitConfig.Apply(m1, m2)
	require.Equal(t, threeTags, limitApply[0].Tags(), "three tags")
	trimmedTags := limitApply[1].Tags()
	require.Len(t, trimmedTags, 3, "ten tags")
	require.Equal(t, "foo", trimmedTags["a"], "preserved: a")
	require.Equal(t, "bar", trimmedTags["b"], "preserved: b")
}

func TestTracking(t *testing.T) {
	inputRaw := []telegraf.Metric{
		metric.New("foo", map[string]string{"tag": "testing"}, map[string]interface{}{"value": 42}, time.Unix(0, 0)),
		metric.New("bar", map[string]string{"tag": "other", "host": "localhost"}, map[string]interface{}{"value": 23}, time.Unix(0, 0)),
		metric.New("baz", map[string]string{"tag": "value", "host": "localhost", "module": "main"}, map[string]interface{}{"value": 99}, time.Unix(0, 0)),
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
			map[string]string{"tag": "testing"},
			map[string]interface{}{"value": 42},
			time.Unix(0, 0),
		),
		metric.New(
			"bar",
			map[string]string{"tag": "other", "host": "localhost"},
			map[string]interface{}{"value": 23},
			time.Unix(0, 0),
		),
		metric.New(
			"baz",
			map[string]string{"tag": "value", "host": "localhost"},
			map[string]interface{}{"value": 99},
			time.Unix(0, 0),
		),
	}

	plugin := &TagLimit{
		Limit: 2,
		Keep:  []string{"tag", "host"},
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
