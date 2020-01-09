package date

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func MustMetric(name string, tags map[string]string, fields map[string]interface{}, metricTime time.Time) telegraf.Metric {
	if tags == nil {
		tags = map[string]string{}
	}
	if fields == nil {
		fields = map[string]interface{}{}
	}
	m, _ := metric.New(name, tags, fields, metricTime)
	return m
}

func TestMonthTag(t *testing.T) {
	dateFormatMonth := Date{
		TagKey:     "month",
		DateFormat: "Jan",
	}

	currentTime := time.Now()
	month := currentTime.Format("Jan")

	m1 := MustMetric("foo", nil, nil, currentTime)
	m2 := MustMetric("bar", nil, nil, currentTime)
	m3 := MustMetric("baz", nil, nil, currentTime)
	monthApply := dateFormatMonth.Apply(m1, m2, m3)
	assert.Equal(t, map[string]string{"month": month}, monthApply[0].Tags(), "should add tag 'month'")
	assert.Equal(t, map[string]string{"month": month}, monthApply[1].Tags(), "should add tag 'month'")
	assert.Equal(t, map[string]string{"month": month}, monthApply[2].Tags(), "should add tag 'month'")
}

func TestYearTag(t *testing.T) {
	dateFormatYear := Date{
		TagKey:     "year",
		DateFormat: "2006",
	}
	currentTime := time.Now()
	year := currentTime.Format("2006")

	m4 := MustMetric("foo", nil, nil, currentTime)
	m5 := MustMetric("bar", nil, nil, currentTime)
	m6 := MustMetric("baz", nil, nil, currentTime)
	yearApply := dateFormatYear.Apply(m4, m5, m6)
	assert.Equal(t, map[string]string{"year": year}, yearApply[0].Tags(), "should add tag 'year'")
	assert.Equal(t, map[string]string{"year": year}, yearApply[1].Tags(), "should add tag 'year'")
	assert.Equal(t, map[string]string{"year": year}, yearApply[2].Tags(), "should add tag 'year'")
}

func TestOldDateTag(t *testing.T) {
	dateFormatYear := Date{
		TagKey:     "year",
		DateFormat: "2006",
	}

	m7 := MustMetric("foo", nil, nil, time.Date(1993, 05, 27, 0, 0, 0, 0, time.UTC))
	customDateApply := dateFormatYear.Apply(m7)
	assert.Equal(t, map[string]string{"year": "1993"}, customDateApply[0].Tags(), "should add tag 'year'")
}

func TestDateOffset(t *testing.T) {
	plugin := &Date{
		TagKey:     "hour",
		DateFormat: "15",
		DateOffset: internal.Duration{Duration: 2 * time.Hour},
	}

	metric := testutil.MustMetric(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"time_idle": 42.0,
		},
		time.Unix(1578603600, 0),
	)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"hour": "15",
			},
			map[string]interface{}{
				"time_idle": 42.0,
			},
			time.Unix(1578603600, 0),
		),
	}

	actual := plugin.Apply(metric)
	testutil.RequireMetricsEqual(t, expected, actual)
}
