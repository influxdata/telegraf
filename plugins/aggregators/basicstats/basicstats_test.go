package basicstats

import (
	"math"
	"testing"
	"time"

	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

var m1, _ = metric.New("m1",
	map[string]string{"foo": "bar"},
	map[string]interface{}{
		"a": int64(1),
		"b": int64(1),
		"c": float64(2),
		"d": float64(2),
	},
	time.Now(),
)
var m2, _ = metric.New("m1",
	map[string]string{"foo": "bar"},
	map[string]interface{}{
		"a":        int64(1),
		"b":        int64(3),
		"c":        float64(4),
		"d":        float64(6),
		"e":        float64(200),
		"ignoreme": "string",
		"andme":    true,
	},
	time.Now(),
)

func BenchmarkApply(b *testing.B) {
	minmax := NewBasicStats()

	for n := 0; n < b.N; n++ {
		minmax.Add(m1)
		minmax.Add(m2)
	}
}

// Test two metrics getting added.
func TestBasicStatsWithPeriod(t *testing.T) {
	acc := testutil.Accumulator{}
	minmax := NewBasicStats()

	minmax.Add(m1)
	minmax.Add(m2)
	minmax.Push(&acc)

	expectedFields := map[string]interface{}{
		"a_count": float64(2), //a
		"a_max":   float64(1),
		"a_min":   float64(1),
		"a_mean":  float64(1),
		"a_stdev": float64(0),
		"a_s2":    float64(0),
		"b_count": float64(2), //b
		"b_max":   float64(3),
		"b_min":   float64(1),
		"b_mean":  float64(2),
		"b_s2":    float64(2),
		"b_stdev": math.Sqrt(2),
		"c_count": float64(2), //c
		"c_max":   float64(4),
		"c_min":   float64(2),
		"c_mean":  float64(3),
		"c_s2":    float64(2),
		"c_stdev": math.Sqrt(2),
		"d_count": float64(2), //d
		"d_max":   float64(6),
		"d_min":   float64(2),
		"d_mean":  float64(4),
		"d_s2":    float64(8),
		"d_stdev": math.Sqrt(8),
		"e_count": float64(1), //e
		"e_max":   float64(200),
		"e_min":   float64(200),
		"e_mean":  float64(200),
	}
	expectedTags := map[string]string{
		"foo": "bar",
	}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)
}

// Test two metrics getting added with a push/reset in between (simulates
// getting added in different periods.)
func TestBasicStatsDifferentPeriods(t *testing.T) {
	acc := testutil.Accumulator{}
	minmax := NewBasicStats()

	minmax.Add(m1)
	minmax.Push(&acc)
	expectedFields := map[string]interface{}{
		"a_count": float64(1), //a
		"a_max":   float64(1),
		"a_min":   float64(1),
		"a_mean":  float64(1),
		"b_count": float64(1), //b
		"b_max":   float64(1),
		"b_min":   float64(1),
		"b_mean":  float64(1),
		"c_count": float64(1), //c
		"c_max":   float64(2),
		"c_min":   float64(2),
		"c_mean":  float64(2),
		"d_count": float64(1), //d
		"d_max":   float64(2),
		"d_min":   float64(2),
		"d_mean":  float64(2),
	}
	expectedTags := map[string]string{
		"foo": "bar",
	}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)

	acc.ClearMetrics()
	minmax.Reset()
	minmax.Add(m2)
	minmax.Push(&acc)
	expectedFields = map[string]interface{}{
		"a_count": float64(1), //a
		"a_max":   float64(1),
		"a_min":   float64(1),
		"a_mean":  float64(1),
		"b_count": float64(1), //b
		"b_max":   float64(3),
		"b_min":   float64(3),
		"b_mean":  float64(3),
		"c_count": float64(1), //c
		"c_max":   float64(4),
		"c_min":   float64(4),
		"c_mean":  float64(4),
		"d_count": float64(1), //d
		"d_max":   float64(6),
		"d_min":   float64(6),
		"d_mean":  float64(6),
		"e_count": float64(1), //e
		"e_max":   float64(200),
		"e_min":   float64(200),
		"e_mean":  float64(200),
	}
	expectedTags = map[string]string{
		"foo": "bar",
	}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)
}
