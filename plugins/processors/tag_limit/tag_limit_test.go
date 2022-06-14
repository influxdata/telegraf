package tag_limit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
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
	require.Equal(t, 3, len(trimmedTags), "ten tags")
	require.Equal(t, "foo", trimmedTags["a"], "preserved: a")
	require.Equal(t, "bar", trimmedTags["b"], "preserved: b")
}
