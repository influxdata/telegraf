package azure_blob

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
)

func TestMetricsForTimeSpan(t *testing.T) {
	m := NewMetricsForTimeSpan()

	iterations := 10
	baseTime := time.Date(2019, 10, 9, 0, 0, 0, 0, &time.Location{})
	currentTime := baseTime

	for i := 0; i < iterations; i++ {
		m.Add(newTimeMetric(currentTime))

		if currentTime.Minute()%2 == 0 {
			for j := 0; j < i; j++ {
				m.Add(newTimeMetric(currentTime))
			}
		}
		currentTime = currentTime.Add(time.Minute)
	}

	// after these two loops, state should be
	// m[0] should have 1 metric - after 0 minutes
	// m[1] should have 1 metric - after 1 minute
	// m[2] should have 3 metric - after 2 minutes
	// m[3] should have 1 metrics  - after 3 minutes
	// m[4] should have 5 metric
	// m[5] should have 1 metrics
	// ...etc

	require.Lenf(t, m.MetricsForPointInTime, iterations, "Incorrect amount of metrics")
	require.Equal(t, m.TimeStart, baseTime)
	require.Equal(t, m.TimeEnd, baseTime.Add(time.Duration(iterations-1)*time.Minute))

	require.Lenf(t, m.MetricsForPointInTime[baseTime.Add(time.Duration(1)*time.Minute)], 1, "Incorrect amount of metrics") //m[1]
	require.Lenf(t, m.MetricsForPointInTime[baseTime.Add(time.Duration(2)*time.Minute)], 3, "Incorrect amount of metrics") //m[2]
	require.Lenf(t, m.MetricsForPointInTime[baseTime.Add(time.Duration(3)*time.Minute)], 1, "Incorrect amount of metrics") //m[3]
	require.Lenf(t, m.MetricsForPointInTime[baseTime.Add(time.Duration(4)*time.Minute)], 5, "Incorrect amount of metrics") //m[4]
}

func newTimeMetric(time time.Time) telegraf.Metric {
	return testutil.MustMetric("", nil, nil, time)
}
