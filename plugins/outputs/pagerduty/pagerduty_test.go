package pagerduty

import (
	"github.com/influxdata/telegraf/testutil"
	"testing"
)

func TestPDAlert(t *testing.T) {
	metric := testutil.TestMetric(2.0, "foo")
	var err error
	var tripped bool
	p := PD{
		Metric:     "foo",
		Field:      "value",
		Expression: "> 5",
	}
	if !p.isMatch(metric) {
		t.Error("Metric should match when name is same")
	}
	tripped, err = p.isTripped(metric)
	if err != nil {
		t.Error(err)
	}
	if tripped {
		t.Error("Metric should not trigger alert when its expression evaluates to false")
	}
	p.Expression = "> 1"
	tripped, err = p.isTripped(metric)
	if err != nil {
		t.Error(err)
	}
	if !tripped {
		t.Error("Metric should trigger alert when expression evaluates to true")
	}
}
