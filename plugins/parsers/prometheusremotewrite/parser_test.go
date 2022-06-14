package prometheusremotewrite

import (
	"testing"
	"time"

	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func TestParse(t *testing.T) {
	prompbInput := prompb.WriteRequest{
		Timeseries: []prompb.TimeSeries{
			{
				Labels: []prompb.Label{
					{Name: "__name__", Value: "go_gc_duration_seconds"},
					{Name: "quantile", Value: "0.99"},
				},
				Samples: []prompb.Sample{
					{Value: 4.63, Timestamp: time.Date(2020, 4, 1, 0, 0, 0, 0, time.UTC).UnixNano()},
				},
			},
			{
				Labels: []prompb.Label{
					{Name: "__name__", Value: "prometheus_target_interval_length_seconds"},
					{Name: "job", Value: "prometheus"},
				},
				Samples: []prompb.Sample{
					{Value: 14.99, Timestamp: time.Date(2020, 4, 1, 0, 0, 0, 0, time.UTC).UnixNano()},
				},
			},
		},
	}

	inoutBytes, err := prompbInput.Marshal()
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"prometheus_remote_write",
			map[string]string{
				"quantile": "0.99",
			},
			map[string]interface{}{
				"go_gc_duration_seconds": float64(4.63),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"prometheus_remote_write",
			map[string]string{
				"job": "prometheus",
			},
			map[string]interface{}{
				"prometheus_target_interval_length_seconds": float64(14.99),
			},
			time.Unix(0, 0),
		),
	}

	parser := Parser{
		DefaultTags: map[string]string{},
	}

	metrics, err := parser.Parse(inoutBytes)
	require.NoError(t, err)
	require.Len(t, metrics, 2)
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestDefaultTags(t *testing.T) {
	prompbInput := prompb.WriteRequest{
		Timeseries: []prompb.TimeSeries{
			{
				Labels: []prompb.Label{
					{Name: "__name__", Value: "foo"},
					{Name: "__eg__", Value: "bar"},
				},
				Samples: []prompb.Sample{
					{Value: 1, Timestamp: time.Date(2020, 4, 1, 0, 0, 0, 0, time.UTC).UnixNano()},
				},
			},
		},
	}

	inoutBytes, err := prompbInput.Marshal()
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"prometheus_remote_write",
			map[string]string{
				"defaultTag": "defaultTagValue",
				"__eg__":     "bar",
			},
			map[string]interface{}{
				"foo": float64(1),
			},
			time.Unix(0, 0),
		),
	}

	parser := Parser{
		DefaultTags: map[string]string{
			"defaultTag": "defaultTagValue",
		},
	}

	metrics, err := parser.Parse(inoutBytes)
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.IgnoreTime(), testutil.SortMetrics())
}

func TestMetricsWithTimestamp(t *testing.T) {
	testTime := time.Date(2020, time.October, 4, 17, 0, 0, 0, time.UTC)
	testTimeUnix := testTime.UnixNano() / int64(time.Millisecond)
	prompbInput := prompb.WriteRequest{
		Timeseries: []prompb.TimeSeries{
			{
				Labels: []prompb.Label{
					{Name: "__name__", Value: "foo"},
					{Name: "__eg__", Value: "bar"},
				},
				Samples: []prompb.Sample{
					{Value: 1, Timestamp: testTimeUnix},
				},
			},
		},
	}

	inoutBytes, err := prompbInput.Marshal()
	require.NoError(t, err)

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"prometheus_remote_write",
			map[string]string{
				"__eg__": "bar",
			},
			map[string]interface{}{
				"foo": float64(1),
			},
			testTime,
		),
	}
	parser := Parser{
		DefaultTags: map[string]string{},
	}

	metrics, err := parser.Parse(inoutBytes)
	require.NoError(t, err)
	require.Len(t, metrics, 1)
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.SortMetrics())
}
