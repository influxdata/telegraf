package basicstats

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

var m1, _ = metric.New("m",
	map[string]string{"foo": "bar"},
	map[string]interface{}{
		"a": int64(1),
		"b": int64(1),
	},
	time.Now(),
)
var m2, _ = metric.New("m",
	map[string]string{"foo": "bar"},
	map[string]interface{}{
		"a": int64(2),
		"b": int64(3),
	},
	time.Now(),
)

var n1, _ = metric.New("n",
	map[string]string{"foos": "ball"},
	map[string]interface{}{
		"a": int64(1),
		"b": int64(1),
	},
	time.Now(),
)
var n2, _ = metric.New("n",
	map[string]string{"foos": "ball"},
	map[string]interface{}{
		"a": int64(2),
		"b": int64(3),
	},
	time.Now(),
)

// Test only aggregating variance
func TestCoStatsCovariance(t *testing.T) {

	acc := testutil.Accumulator{}

	aggregator := NewBasicStats()
	aggregator.Stats = []string{}
	aggregator.CoStatsConfig = []costat{
		{Metrics: []metrics{
			{Name: "n", Field: "a"}, {Name: "m", Field: "a"},
		}},
		{Metrics: []metrics{
			{Name: "m", Field: "b"}, {Name: "n", Field: "b"},
		}},
	}

	aggregator.Add(n1)
	aggregator.Add(m1)
	aggregator.Add(n2)
	aggregator.Add(m2)
	aggregator.Add(n1)
	aggregator.Add(m1)
	aggregator.Add(n2)
	aggregator.Add(m2)
	aggregator.Add(n1)
	aggregator.Add(m1)
	aggregator.Add(n2)
	aggregator.Add(m2)

	aggregator.Push(&acc)

	expectedFields := map[string]interface{}{
		"correlation[m/a/foo:bar][n/a/foos:ball]": float64(2.202962962962964),
		"correlation[m/b/foo:bar][n/b/foos:ball]": float64(2.2029629629629635),
		"covariance[m/a/foo:bar][n/a/foos:ball]":  float64(0.660888888888889),
		"covariance[m/b/foo:bar][n/b/foos:ball]":  float64(2.6435555555555554),
	}
	expectedTags := map[string]string{
		"foo": "bar",
	}
	acc.AssertContainsTaggedFields(t, "m", expectedFields, expectedTags)
}

// Test only aggregating variance
func TestCoStatsAccuracyCostats(t *testing.T) {

	acc := testutil.Accumulator{}

	aggregator := NewBasicStats()
	aggregator.Stats = []string{}
	aggregator.CoStatsConfig = []costat{
		{Metrics: []metrics{
			{Name: "n", Field: "a"}, {Name: "m", Field: "a"},
		}},
		{Metrics: []metrics{
			{Name: "m", Field: "b"}, {Name: "n", Field: "b"},
		}},
	}

	for i := 1; i < 11; i++ {
		m, _ := metric.New("m",
			map[string]string{"foo": "bar"},
			map[string]interface{}{
				"a": int64(i),
				"b": int64(i * 2),
			},
			time.Now(),
		)
		aggregator.Add(m)
		n, _ := metric.New("n",
			map[string]string{"foos": "ball"},
			map[string]interface{}{
				"a": int64(i),
				"b": int64(i * 2),
			},
			time.Now(),
		)
		aggregator.Add(n)
	}

	aggregator.Push(&acc)

	expectedFields := map[string]interface{}{
		"correlation[m/a/foo:bar][n/a/foos:ball]": float64(0.9939393939393938),
		"correlation[m/b/foo:bar][n/b/foos:ball]": float64(0.9939393939393938),
		"covariance[m/a/foo:bar][n/a/foos:ball]":  float64(9.11111111111111),
		"covariance[m/b/foo:bar][n/b/foos:ball]":  float64(36.44444444444444),
	}
	expectedTags := map[string]string{
		"foo": "bar",
	}
	acc.AssertContainsTaggedFields(t, "m", expectedFields, expectedTags)
}
