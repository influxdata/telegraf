package modbus_gateway

import (
	"github.com/influxdata/telegraf/metric"
	"reflect"
	"testing"
	"time"
)

func TestModbusGroupings(t *testing.T) {

	/*
	 * For this plugin to work, it has to be able to put together multiple
	 * fields into a single measurement
	 */
	g := metric.NewSeriesGrouper()

	tm := time.Now()
	r := Request{}
	f1 := FieldDef{
		Name:         "f1",
		Scale:        1.0,
		Offset:       0.0,
		OutputFormat: "INT64",
		Omit:         false,
	}
	f2 := FieldDef{
		Name:         "f2",
		Scale:        1.0,
		Offset:       0.0,
		OutputFormat: "FLOAT",
		Omit:         false,
	}

	outputToGroup(g, &r, &f1, 1, tm)
	outputToGroup(g, &r, &f2, 1, tm)

	if len(g.Metrics()) != 1 {
		t.Errorf("Grouping failed - should have generated 1 metric, but generated %d\n", len(g.Metrics()))
	}

	t.Logf("Metrics: %++v", g.Metrics())

	firstMetric := g.Metrics()[0]

	rf1, ok := firstMetric.GetField("f1")
	if !ok {
		t.Errorf("Field f1 was not in output metric")
		return
	}

	_, ok = rf1.(int64)
	if !ok {
		t.Errorf("Metric did not match field type specificier, type was %s", reflect.TypeOf(rf1))
	}

	rf2, ok := firstMetric.GetField("f2")
	if !ok {
		t.Errorf("Field f2 was not in output metric")
		return
	}

	_, ok = rf2.(float64)
	if !ok {
		t.Errorf("Metric did not match field type specificier, type was %s", reflect.TypeOf(rf1))
	}

}
