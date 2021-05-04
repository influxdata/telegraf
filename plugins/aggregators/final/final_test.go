package final

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestSimple(t *testing.T) {
	acc := testutil.Accumulator{}
	final := NewFinal()

	tags := map[string]string{"foo": "bar"}
	m1 := metric.New("m1",
		tags,
		map[string]interface{}{"a": int64(1)},
		time.Unix(1530939936, 0))
	m2 := metric.New("m1",
		tags,
		map[string]interface{}{"a": int64(2)},
		time.Unix(1530939937, 0))
	m3 := metric.New("m1",
		tags,
		map[string]interface{}{"a": int64(3)},
		time.Unix(1530939938, 0))
	final.Add(m1)
	final.Add(m2)
	final.Add(m3)
	final.Push(&acc)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"m1",
			tags,
			map[string]interface{}{
				"a_final": 3,
			},
			time.Unix(1530939938, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics())
}

func TestTwoTags(t *testing.T) {
	acc := testutil.Accumulator{}
	final := NewFinal()

	tags1 := map[string]string{"foo": "bar"}
	tags2 := map[string]string{"foo": "baz"}

	m1 := metric.New("m1",
		tags1,
		map[string]interface{}{"a": int64(1)},
		time.Unix(1530939936, 0))
	m2 := metric.New("m1",
		tags2,
		map[string]interface{}{"a": int64(2)},
		time.Unix(1530939937, 0))
	m3 := metric.New("m1",
		tags1,
		map[string]interface{}{"a": int64(3)},
		time.Unix(1530939938, 0))
	final.Add(m1)
	final.Add(m2)
	final.Add(m3)
	final.Push(&acc)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"m1",
			tags2,
			map[string]interface{}{
				"a_final": 2,
			},
			time.Unix(1530939937, 0),
		),
		testutil.MustMetric(
			"m1",
			tags1,
			map[string]interface{}{
				"a_final": 3,
			},
			time.Unix(1530939938, 0),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.SortMetrics())
}

func TestLongDifference(t *testing.T) {
	acc := testutil.Accumulator{}
	final := NewFinal()
	final.SeriesTimeout = config.Duration(30 * time.Second)
	tags := map[string]string{"foo": "bar"}

	now := time.Now()

	m1 := metric.New("m",
		tags,
		map[string]interface{}{"a": int64(1)},
		now.Add(time.Second*-290))
	m2 := metric.New("m",
		tags,
		map[string]interface{}{"a": int64(2)},
		now.Add(time.Second*-275))
	m3 := metric.New("m",
		tags,
		map[string]interface{}{"a": int64(3)},
		now.Add(time.Second*-100))
	m4 := metric.New("m",
		tags,
		map[string]interface{}{"a": int64(4)},
		now.Add(time.Second*-20))
	final.Add(m1)
	final.Add(m2)
	final.Push(&acc)
	final.Add(m3)
	final.Push(&acc)
	final.Add(m4)
	final.Push(&acc)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"m",
			tags,
			map[string]interface{}{
				"a_final": 2,
			},
			now.Add(time.Second*-275),
		),
		testutil.MustMetric(
			"m",
			tags,
			map[string]interface{}{
				"a_final": 3,
			},
			now.Add(time.Second*-100),
		),
	}
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.SortMetrics())
}
