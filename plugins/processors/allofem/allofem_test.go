package allofem

import (
	"log"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/require"
)

func MakeInputMetrics(num int) []telegraf.Metric {
	allOfEm := make([]telegraf.Metric, 0)
	for i := 1; i <= num; i++ {
		m, _ := metric.New("test", map[string]string{}, map[string]interface{}{"sample": i}, time.Now())
		allOfEm = append(allOfEm, m)
	}
	return allOfEm
}

func MakeOutlier(val int) telegraf.Metric {
	m, _ := metric.New("test", make(map[string]string), map[string]interface{}{"sample": val}, time.Now())
	return m
}

func TestAllOfEm(t *testing.T) {
	tests := []struct {
		fieldName       string
		windowSize      int
		percent         int
		outlierDistance float64
		name            string
		input           []telegraf.Metric
		want            int
		now             func() time.Time
		wantErr         bool
	}{
		{
			name:            "Test no percentiles outputted",
			now:             func() time.Time { return time.Unix(0, 0) },
			input:           MakeInputMetrics(20),
			percent:         0,
			fieldName:       "sample",
			outlierDistance: 2,
			windowSize:      6,
			want:            0,
		},
		{
			name:            "Test outliers outputted",
			now:             func() time.Time { return time.Unix(0, 0) },
			input:           append(MakeInputMetrics(20), MakeOutlier(3000)),
			percent:         0,
			fieldName:       "sample",
			outlierDistance: 2,
			windowSize:      6,
			want:            1,
		},
		{
			name:            "Test outliers+percentile outputted",
			now:             func() time.Time { return time.Unix(0, 0) },
			input:           append(MakeInputMetrics(100), MakeOutlier(3000)),
			percent:         5,
			fieldName:       "sample",
			outlierDistance: 2,
			windowSize:      6,
			want:            6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AllOfEm{
				WindowSize:       tt.windowSize,
				StatsField:       tt.fieldName,
				PercentOfMetrics: tt.percent,
				OutlierDistance:  tt.outlierDistance,
			}
			got := a.Apply(tt.input...)
			for i, m := range got {
				log.Printf("m[%v]: %v", i, m)
			}
			require.Equal(t, tt.want, len(got))
		})
	}
}
