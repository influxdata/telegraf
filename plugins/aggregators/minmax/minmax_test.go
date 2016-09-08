package minmax

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
)

func BenchmarkApply(b *testing.B) {
	minmax := MinMax{}
	minmax.clearCache()

	m1, _ := telegraf.NewMetric("m1",
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
	m2, _ := telegraf.NewMetric("m1",
		map[string]string{"foo": "bar"},
		map[string]interface{}{
			"a": int64(3),
			"b": int64(3),
			"c": int64(3),
			"d": int64(3),
			"e": int64(3),
			"f": float64(1),
			"g": float64(1),
			"h": float64(1),
			"i": float64(1),
			"j": float64(1),
		},
		time.Now(),
	)

	for n := 0; n < b.N; n++ {
		minmax.apply(m1)
		minmax.apply(m2)
	}
}
