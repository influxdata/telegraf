package http

import (
	"strings"
	"testing"
	"time"

	"github.com/signalfx/golib/datapoint"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/require"
)

func getMetric() telegraf.Metric {
	m, err := metric.New(
		"cpu",
		map[string]string{
			"tagKey": "tagValue",
		},
		map[string]interface{}{
			"time_user":       43.0,
			"time_system":     44.0,
			"time_idle":       45.0,
			"time_nice":       46.0,
			"time_iowait":     47.0,
			"time_irq":        48.0,
			"time_softirq":    49.0,
			"time_steal":      50.0,
			"time_guest":      51.0,
			"time_guest_nice": 52.0,
		},
		time.Unix(0, 0),
	)
	if err != nil {
		panic(err)
	}
	return m
}

func TestTelegrafTypeToSignalFXDatapoint(t *testing.T) {
	m := getMetric()
	d := telegrafMetricToSignalFXDatapoints(m)
	require.Len(t, d, 10, "len mismatch")

	for _, datapoint := range d {
		nameParts := strings.Split(datapoint.Metric, ".")
		require.True(t, (nameParts[0] == m.Name()))
		require.True(t, m.HasField(nameParts[1]))
	}
}

func TestTelegrafTypeToSignalFXType(t *testing.T) {
	tests := map[telegraf.ValueType]datapoint.MetricType{
		telegraf.Histogram: datapoint.Gauge,
		telegraf.Summary:   datapoint.Gauge,
		telegraf.Gauge:     datapoint.Gauge,
		telegraf.Counter:   datapoint.Counter,
		telegraf.Untyped:   datapoint.Gauge,
	}

	for arg, expected := range tests {
		require.Equal(t, telegrafTypeToSignalFXType(arg), expected)
	}
}
