package minmax

import (
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
		"c": int64(1),
		"d": int64(1),
		"e": int64(1),
		"f": float64(2),
		"g": float64(2),
		"h": float64(2),
		"i": float64(2),
		"j": float64(3),
	},
	time.Now(),
)
var m2, _ = metric.New("m1",
	map[string]string{"foo": "bar"},
	map[string]interface{}{
		"a":        int64(1),
		"b":        int64(3),
		"c":        int64(3),
		"d":        int64(3),
		"e":        int64(3),
		"f":        float64(1),
		"g":        float64(1),
		"h":        float64(1),
		"i":        float64(1),
		"j":        float64(1),
		"k":        float64(200),
		"l":        uint64(200),
		"ignoreme": "string",
		"andme":    true,
	},
	time.Now(),
)

func BenchmarkApply(b *testing.B) {
	minmax := NewMinMax()

	for n := 0; n < b.N; n++ {
		minmax.Add(m1)
		minmax.Add(m2)
	}
}

// Test two metrics getting added.
func TestMinMaxWithPeriod(t *testing.T) {
	acc := testutil.Accumulator{}
	minmax := NewMinMax()

	minmax.Add(m1)
	minmax.Add(m2)
	minmax.Push(&acc)

	expectedFields := map[string]interface{}{
		"a_max": float64(1),
		"a_min": float64(1),
		"b_max": float64(3),
		"b_min": float64(1),
		"c_max": float64(3),
		"c_min": float64(1),
		"d_max": float64(3),
		"d_min": float64(1),
		"e_max": float64(3),
		"e_min": float64(1),
		"f_max": float64(2),
		"f_min": float64(1),
		"g_max": float64(2),
		"g_min": float64(1),
		"h_max": float64(2),
		"h_min": float64(1),
		"i_max": float64(2),
		"i_min": float64(1),
		"j_max": float64(3),
		"j_min": float64(1),
		"k_max": float64(200),
		"k_min": float64(200),
		"l_max": float64(200),
		"l_min": float64(200),
	}
	expectedTags := map[string]string{
		"foo": "bar",
	}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)
}

// Test two metrics getting added with a push/reset in between (simulates
// getting added in different periods.)
func TestMinMaxDifferentPeriods(t *testing.T) {
	acc := testutil.Accumulator{}
	minmax := NewMinMax()

	minmax.Add(m1)
	minmax.Push(&acc)
	expectedFields := map[string]interface{}{
		"a_max": float64(1),
		"a_min": float64(1),
		"b_max": float64(1),
		"b_min": float64(1),
		"c_max": float64(1),
		"c_min": float64(1),
		"d_max": float64(1),
		"d_min": float64(1),
		"e_max": float64(1),
		"e_min": float64(1),
		"f_max": float64(2),
		"f_min": float64(2),
		"g_max": float64(2),
		"g_min": float64(2),
		"h_max": float64(2),
		"h_min": float64(2),
		"i_max": float64(2),
		"i_min": float64(2),
		"j_max": float64(3),
		"j_min": float64(3),
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
		"a_max": float64(1),
		"a_min": float64(1),
		"b_max": float64(3),
		"b_min": float64(3),
		"c_max": float64(3),
		"c_min": float64(3),
		"d_max": float64(3),
		"d_min": float64(3),
		"e_max": float64(3),
		"e_min": float64(3),
		"f_max": float64(1),
		"f_min": float64(1),
		"g_max": float64(1),
		"g_min": float64(1),
		"h_max": float64(1),
		"h_min": float64(1),
		"i_max": float64(1),
		"i_min": float64(1),
		"j_max": float64(1),
		"j_min": float64(1),
		"k_max": float64(200),
		"k_min": float64(200),
		"l_max": float64(200),
		"l_min": float64(200),
	}
	expectedTags = map[string]string{
		"foo": "bar",
	}
	acc.AssertContainsTaggedFields(t, "m1", expectedFields, expectedTags)
}
