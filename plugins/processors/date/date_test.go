package date

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
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

func TestTagAndField(t *testing.T) {
	dateFormatTagAndField := Date{
		TagKey:   "month",
		FieldKey: "month",
	}
	err := dateFormatTagAndField.Init()
	require.Error(t, err)
}

func TestNoOutputSpecified(t *testing.T) {
	dateFormatNoOutput := Date{}
	err := dateFormatNoOutput.Init()
	require.Error(t, err)
}

func TestMonthTag(t *testing.T) {
	dateFormatMonth := Date{
		TagKey:     "month",
		DateFormat: "Jan",
	}
	err := dateFormatMonth.Init()
	require.NoError(t, err)

	currentTime := time.Now()
	month := currentTime.Format("Jan")

	m1 := MustMetric("foo", nil, nil, currentTime)
	m2 := MustMetric("bar", nil, nil, currentTime)
	m3 := MustMetric("baz", nil, nil, currentTime)
	monthApply := dateFormatMonth.Apply(m1, m2, m3)
	require.Equal(t, map[string]string{"month": month}, monthApply[0].Tags(), "should add tag 'month'")
	require.Equal(t, map[string]string{"month": month}, monthApply[1].Tags(), "should add tag 'month'")
	require.Equal(t, map[string]string{"month": month}, monthApply[2].Tags(), "should add tag 'month'")
}

func TestMonthField(t *testing.T) {
	dateFormatMonth := Date{
		FieldKey:   "month",
		DateFormat: "Jan",
	}

	err := dateFormatMonth.Init()
	require.NoError(t, err)

	currentTime := time.Now()
	month := currentTime.Format("Jan")

	m1 := MustMetric("foo", nil, nil, currentTime)
	m2 := MustMetric("bar", nil, nil, currentTime)
	m3 := MustMetric("baz", nil, nil, currentTime)
	monthApply := dateFormatMonth.Apply(m1, m2, m3)
	require.Equal(t, map[string]interface{}{"month": month}, monthApply[0].Fields(), "should add field 'month'")
	require.Equal(t, map[string]interface{}{"month": month}, monthApply[1].Fields(), "should add field 'month'")
	require.Equal(t, map[string]interface{}{"month": month}, monthApply[2].Fields(), "should add field 'month'")
}

func TestOldDateTag(t *testing.T) {
	dateFormatYear := Date{
		TagKey:     "year",
		DateFormat: "2006",
	}

	err := dateFormatYear.Init()
	require.NoError(t, err)

	m7 := MustMetric("foo", nil, nil, time.Date(1993, 05, 27, 0, 0, 0, 0, time.UTC))
	customDateApply := dateFormatYear.Apply(m7)
	require.Equal(t, map[string]string{"year": "1993"}, customDateApply[0].Tags(), "should add tag 'year'")
}

func TestFieldUnix(t *testing.T) {
	dateFormatUnix := Date{
		FieldKey:   "unix",
		DateFormat: "unix",
	}

	err := dateFormatUnix.Init()
	require.NoError(t, err)

	currentTime := time.Now()
	unixTime := currentTime.Unix()

	m8 := MustMetric("foo", nil, nil, currentTime)
	unixApply := dateFormatUnix.Apply(m8)
	require.Equal(t, map[string]interface{}{"unix": unixTime}, unixApply[0].Fields(), "should add unix time in s as field 'unix'")
}

func TestFieldUnixNano(t *testing.T) {
	dateFormatUnixNano := Date{
		FieldKey:   "unix_ns",
		DateFormat: "unix_ns",
	}

	err := dateFormatUnixNano.Init()
	require.NoError(t, err)

	currentTime := time.Now()
	unixNanoTime := currentTime.UnixNano()

	m9 := MustMetric("foo", nil, nil, currentTime)
	unixNanoApply := dateFormatUnixNano.Apply(m9)
	require.Equal(t, map[string]interface{}{"unix_ns": unixNanoTime}, unixNanoApply[0].Fields(), "should add unix time in ns as field 'unix_ns'")
}

func TestFieldUnixMillis(t *testing.T) {
	dateFormatUnixMillis := Date{
		FieldKey:   "unix_ms",
		DateFormat: "unix_ms",
	}

	err := dateFormatUnixMillis.Init()
	require.NoError(t, err)

	currentTime := time.Now()
	unixMillisTime := currentTime.UnixNano() / 1000000

	m10 := MustMetric("foo", nil, nil, currentTime)
	unixMillisApply := dateFormatUnixMillis.Apply(m10)
	require.Equal(t, map[string]interface{}{"unix_ms": unixMillisTime}, unixMillisApply[0].Fields(), "should add unix time in ms as field 'unix_ms'")
}

func TestFieldUnixMicros(t *testing.T) {
	dateFormatUnixMicros := Date{
		FieldKey:   "unix_us",
		DateFormat: "unix_us",
	}

	err := dateFormatUnixMicros.Init()
	require.NoError(t, err)

	currentTime := time.Now()
	unixMicrosTime := currentTime.UnixNano() / 1000

	m11 := MustMetric("foo", nil, nil, currentTime)
	unixMicrosApply := dateFormatUnixMicros.Apply(m11)
	require.Equal(t, map[string]interface{}{"unix_us": unixMicrosTime}, unixMicrosApply[0].Fields(), "should add unix time in us as field 'unix_us'")
}

func TestDateOffset(t *testing.T) {
	plugin := &Date{
		TagKey:     "hour",
		DateFormat: "15",
		DateOffset: config.Duration(2 * time.Hour),
	}

	err := plugin.Init()
	require.NoError(t, err)

	m := testutil.MustMetric(
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
				"hour": "23",
			},
			map[string]interface{}{
				"time_idle": 42.0,
			},
			time.Unix(1578603600, 0),
		),
	}

	actual := plugin.Apply(m)
	testutil.RequireMetricsEqual(t, expected, actual)
}
