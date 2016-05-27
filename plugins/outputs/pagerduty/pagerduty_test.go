package pagerduty

import (
	"github.com/influxdata/telegraf/testutil"
	"testing"
)

func TestMetricMatch(t *testing.T) {
	metric := testutil.TestMetric(1.0, "foo")
	p := PD{
		Metric:     "foo",
		Field:      "value",
		Expression: "> 0",
	}
	if !p.Match(metric) {
		t.Error("Metric did not match for greater than expression")
	}
	p.Expression = "== 1"
	if !p.Match(metric) {
		t.Error("Metric did not match for equality expression")
	}
	p.Expression = "< 0"
	if p.Match(metric) {
		t.Error("Metric did not match for less than expression")
	}
}
