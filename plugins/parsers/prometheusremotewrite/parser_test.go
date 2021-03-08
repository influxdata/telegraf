package prometheusremotewrite

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	// prometheus time series
	input := prompb.WriteRequest{
		Timeseries: []*prompb.TimeSeries{
			{
				Labels: []*prompb.Label{
					{Name: "__name__", Value: "foo"},
				},
				Samples: []prompb.Sample{
					{Value: 1, Timestamp: time.Date(2020, 4, 1, 0, 0, 0, 0, time.UTC).UnixNano()},
				},
			},
		},
	}
	// Marshal it
	inoutBytes, err := input.Marshal()

	// Expected telegraf metric
	expected := []telegraf.Metric{
		testutil.MustMetric(
			"prometheusremotewrite",
			map[string]string{},
			map[string]interface{}{
				"foo": float64(1),
			},
			time.Unix(0, 0),
		),
	}
	parser := Parser{
		DefaultTags: map[string]string{},
	}
	// hand it to parser
	metrics, err := parser.Parse(inoutBytes)
	// fmt.Println("metrics", metrics[0])
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	testutil.RequireMetricsEqual(t, expected, metrics, testutil.IgnoreTime(), testutil.SortMetrics())
}
