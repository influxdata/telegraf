package modbusgw

import (
	"github.com/influxdata/telegraf/metric"
	"time"
)

func writeInt(grouper *metric.SeriesGrouper, req *Request, f *FieldDef, value int64, timestamp time.Time) {

	if !f.Omit {
		if f.Scale == 1.0 && f.Offset == 0.0 {
			grouper.Add(req.Measurement, nil, timestamp, f.Name, value)
		} else {
			grouper.Add(req.Measurement, nil, timestamp, f.Name, float64(value)*float64(f.Scale+f.Offset))
		}
	}
}
