package azure_blob

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
)

var (
	veryOldTime    = time.Date(1900, 1, 1, 0, 0, 0, 0, &time.Location{})
	veryFutureTime = time.Date(2900, 1, 1, 0, 0, 0, 0, &time.Location{})
)

type MetricsForTimespan struct {
	MetricsForPointInTime map[time.Time][]*telegraf.Metric
	TimeStart             time.Time
	TimeEnd               time.Time
}

func NewMetricsForTimeSpan() *MetricsForTimespan {
	return &MetricsForTimespan{
		TimeStart:             veryFutureTime,
		TimeEnd:               veryOldTime,
		MetricsForPointInTime: make(map[time.Time][]*telegraf.Metric),
	}
}

func (m *MetricsForTimespan) Add(metric telegraf.Metric) {
	m.MetricsForPointInTime[metric.Time()] = append(m.MetricsForPointInTime[metric.Time()], &metric)

	if metric.Time().Sub(m.TimeStart) < 0 {
		m.TimeStart = metric.Time()
	}

	if metric.Time().Sub(m.TimeEnd) > 0 {
		m.TimeEnd = metric.Time()
	}
}

func (m *MetricsForTimespan) Export() string {
	return fmt.Sprintf("%v\n", m)
}
