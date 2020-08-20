package modbus_gateway

import (
	"github.com/influxdata/telegraf/metric"
	"testing"
	"time"
)

func TestGrouping(t *testing.T) {

	/*
	 * For this plugin to work, it has to be able to put together multiple
	 * fields into a single measurement
	 */
	g := metric.NewSeriesGrouper()

	tm := time.Now()
	r := Request{}
	f1 := FieldDef{Name: "f1", Scale: 1.0, Offset: 0.0}
	f2 := FieldDef{Name: "f2", Scale: 1.0, Offset: 0.0}

	writeInt(g, &r, &f1, 1, tm)
	writeInt(g, &r, &f2, 1, tm)

	if len(g.Metrics()) != 1 {
		t.Errorf("Grouping failed - should have generated 1 metric, but generated %d\n", len(g.Metrics()))
	}

}
