package request_aggregates

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestTimeWindow_Add(t *testing.T) {
	tw := &TimeWindow{}

	tw.Add(&Request{Time: 123.45})
	require.Equal(t, 1, len(tw.TimesTotal))
	require.Equal(t, 1, len(tw.TimesSuccess))
	require.Equal(t, 0, len(tw.TimesFail))
	require.Equal(t, float64(123.45), tw.TimesTotal[0])
	require.Equal(t, float64(123.45), tw.TimesSuccess[0])

	tw.Add(&Request{Time: 100, Failure: false})
	require.Equal(t, 2, len(tw.TimesTotal))
	require.Equal(t, 2, len(tw.TimesSuccess))
	require.Equal(t, 0, len(tw.TimesFail))
	require.Equal(t, float64(100), tw.TimesTotal[1])
	require.Equal(t, float64(100), tw.TimesSuccess[1])

	tw.Add(&Request{Time: 200, Failure: true})
	require.Equal(t, 3, len(tw.TimesTotal))
	require.Equal(t, 2, len(tw.TimesSuccess))
	require.Equal(t, 1, len(tw.TimesFail))
	require.Equal(t, float64(200), tw.TimesTotal[2])
	require.Equal(t, float64(200), tw.TimesFail[0])
}

func TestTimeWindow_Start(t *testing.T) {
	now := time.Now()
	tw := &TimeWindow{StartTime: now}
	require.Equal(t, now, tw.Start())
}

func TestTimeWindow_End(t *testing.T) {
	now := time.Now()
	tw := &TimeWindow{EndTime: now}
	require.Equal(t, now, tw.End())
}

func TestTimeWindow_Aggregate_All(t *testing.T) {
	start := time.Now()
	end := start.Add(time.Duration(60))
	tw := &TimeWindow{StartTime: start, EndTime: end, OnlyTotal: false}
	metrics, err := tw.Aggregate()
	require.NoError(t, err)
	require.Equal(t, 3, len(metrics))
	require.Equal(t, end, metrics[0].Time())
	require.Equal(t, MeasurementTime, metrics[0].Name())
	require.Equal(t, end, metrics[1].Time())
	require.Equal(t, MeasurementTimeFail, metrics[1].Name())
	require.Equal(t, end, metrics[0].Time())
	require.Equal(t, MeasurementTimeSuccess, metrics[2].Name())
}

func TestTimeWindow_Aggregate_OnlyTotal(t *testing.T) {
	start := time.Now()
	end := start.Add(time.Duration(60))
	tw := &TimeWindow{StartTime: start, EndTime: end, OnlyTotal: true}
	metrics, err := tw.Aggregate()
	require.NoError(t, err)
	require.Equal(t, 1, len(metrics))
	require.Equal(t, end, metrics[0].Time())
	require.Equal(t, MeasurementTime, metrics[0].Name())
}

func TestTimeWindow_aggregateTimes(t *testing.T) {
	end := time.Now()
	metric, err := aggregateTimes(MeasurementTime, []float64{500, 900, 300, 1000, 100, 600, 700, 800, 200, 400},
		[]float32{60, 80, 99.9}, end)
	require.NoError(t, err)
	require.Equal(t, MeasurementTime, metric.Name())
	require.Equal(t, int64(10), metric.Fields()[FieldTimeRequests])
	require.Equal(t, float64(1000), metric.Fields()[FieldTimeMax])
	require.Equal(t, float64(100), metric.Fields()[FieldTimeMin])
	require.Equal(t, float64(550), metric.Fields()[FieldTimeMean])
	require.Equal(t, float64(700), metric.Fields()[FieldTimePerc+"60"])
	require.Equal(t, float64(900), metric.Fields()[FieldTimePerc+"80"])
	require.Equal(t, float64(1000), metric.Fields()[FieldTimePerc+"99_9"])
}

func TestThroughputWindow_Add(t *testing.T) {
	tw := &ThroughputWindow{}

	tw.Add(&Request{})
	require.Equal(t, int64(1), tw.RequestsTotal)
	require.Equal(t, int64(0), tw.RequestsFail)

	tw.Add(&Request{Failure: false})
	require.Equal(t, int64(2), tw.RequestsTotal)
	require.Equal(t, int64(0), tw.RequestsFail)

	tw.Add(&Request{Failure: true})
	require.Equal(t, int64(3), tw.RequestsTotal)
	require.Equal(t, int64(1), tw.RequestsFail)
}

func TestThroughputWindow_Start(t *testing.T) {
	now := time.Now()
	tw := &ThroughputWindow{StartTime: now}
	require.Equal(t, now, tw.Start())
}

func TestThroughputWindow_End(t *testing.T) {
	now := time.Now()
	tw := &ThroughputWindow{EndTime: now}
	require.Equal(t, now, tw.End())
}

func TestThroughputWindow_Aggregate(t *testing.T) {
	start := time.Now()
	end := start.Add(time.Duration(60))
	tw := &ThroughputWindow{StartTime: start, EndTime: end, RequestsTotal: 33, RequestsFail: 11}
	metrics, err := tw.Aggregate()
	require.NoError(t, err)
	require.Equal(t, 1, len(metrics))
	require.Equal(t, end, metrics[0].Time())
	require.Equal(t, MeasurementThroughput, metrics[0].Name())
	require.Equal(t, int64(33), metrics[0].Fields()[FieldThroughputTotal])
	require.Equal(t, int64(11), metrics[0].Fields()[FieldThroughputFailed])
}
