package date

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func newMetric(name string, tags map[string]string, fields map[string]interface{}, metricTime time.Time) telegraf.Metric {
	if tags == nil {
		tags = map[string]string{}
	}
	if fields == nil {
		fields = map[string]interface{}{}
	}
	m, _ := metric.New(name, tags, fields, metricTime)
	return m
}

func TestDateTag(t *testing.T) {
	dateFormatMonth := Date{
		TagKey:     "month",
		DateFormat: "Jan",
	}

	dateFormatYear := Date{
		TagKey:     "year",
		DateFormat: "2006",
	}

	currentTime := time.Now()
	month := currentTime.Format("Jan")
	year := currentTime.Format("2006")

	m1 := newMetric("foo", nil, nil, currentTime)
	m2 := newMetric("bar", nil, nil, currentTime)
	m3 := newMetric("baz", nil, nil, currentTime)
	monthApply := dateFormatMonth.Apply(m1, m2, m3)
	assert.Equal(t, map[string]string{"month": month}, monthApply[0].Tags(), "should add tag 'month'")
	assert.Equal(t, map[string]string{"month": month}, monthApply[1].Tags(), "should add tag 'month'")
	assert.Equal(t, map[string]string{"month": month}, monthApply[2].Tags(), "should add tag 'month'")

	m4 := newMetric("foo", nil, nil, currentTime)
	m5 := newMetric("bar", nil, nil, currentTime)
	m6 := newMetric("baz", nil, nil, currentTime)
	yearApply := dateFormatYear.Apply(m4, m5, m6)
	assert.Equal(t, map[string]string{"year": year}, yearApply[0].Tags(), "should add tag 'year'")
	assert.Equal(t, map[string]string{"year": year}, yearApply[1].Tags(), "should add tag 'year'")
	assert.Equal(t, map[string]string{"year": year}, yearApply[2].Tags(), "should add tag 'year'")

	m7 := newMetric("foo", nil, nil, time.Date(1993, 05, 27, 0, 0, 0, 0, time.UTC))
	customDateApply := dateFormatYear.Apply(m7)
	assert.Equal(t, map[string]string{"year": "1993"}, customDateApply[0].Tags(), "should add tag 'year'")
}
