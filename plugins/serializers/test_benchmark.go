package serializers

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func BenchmarkMetrics(b *testing.B) [4]telegraf.Metric {
	b.Helper()
	now := time.Now()
	tags := map[string]string{
		"cpu":  "cpu0",
		"host": "realHost",
	}
	newMetric := func(v interface{}) telegraf.Metric {
		fields := map[string]interface{}{
			"usage_idle": v,
		}
		m := metric.New("cpu", tags, fields, now)
		return m
	}
	return [4]telegraf.Metric{
		newMetric(91.5),
		newMetric(91),
		newMetric(true),
		newMetric(false),
	}
}
