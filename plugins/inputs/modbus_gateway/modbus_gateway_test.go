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

	f1 := FieldDef{
		Name:         "f1",
		Scale:        1.0,
		Offset:       0.0,
		OutputFormat: "UINT32",
		Omit:         false,
	}
	f2 := FieldDef{
		Name:         "f2",
		Scale:        1.0,
		Offset:       0.0,
		OutputFormat: "UINT32",
		Omit:         false,
	}

	e := g.Add("m", nil, tm, f1.Name, scale(&f1, uint32(1)))
	if e != nil {
		t.Errorf("Could not add field %s", f1.Name)
	}

	e = g.Add("m", nil, tm, f2.Name, scale(&f2, uint32(1)))
	if e != nil {
		t.Errorf("Could not add field %s", f2.Name)
	}

	if len(g.Metrics()) != 1 {
		t.Errorf("Grouping failed - should have generated 1 metric, but generated %d\n", len(g.Metrics()))
	}

	firstMetric := g.Metrics()[0]

	rf1, ok := firstMetric.GetField("f1")
	if !ok {
		t.Errorf("Field f1 was not in output metric")
		return
	}

	_, ok = rf1.(uint64)
	if !ok {
		t.Errorf("Metric did not match field type specificier, type was %s", reflect.TypeOf(rf1))
	}

	rf2, ok := firstMetric.GetField("f2")
	if !ok {
		t.Errorf("Field f2 was not in output metric")
		return
	}

	_, ok = rf2.(uint64)
	if !ok {
		t.Errorf("Metric did not match field type specificier, type was %s", reflect.TypeOf(rf1))
	}

}
