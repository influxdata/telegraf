package stats

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/require"
)

func MakeInputMetrics(num int) []telegraf.Metric {
	allOfEm := make([]telegraf.Metric, 0)
	for i := 1; i <= num; i++ {
		m, _ := metric.New("test", map[string]string{}, map[string]interface{}{"sample_field": i}, time.Now())
		allOfEm = append(allOfEm, m)
	}
	return allOfEm
}

func MakeOutputMetrics(num int) []telegraf.Metric {
	allOfEm := make([]telegraf.Metric, 0)
	for i := 1; i <= num; i++ {
		m, _ := metric.New("test", map[string]string{}, map[string]interface{}{"sample_field": i}, time.Now())
		allOfEm = append(allOfEm, m)
	}
	return allOfEm
}

func TestStats(t *testing.T) {
	tests := []struct {
		StatsField string
		WindowSize int
		name       string
		input      []telegraf.Metric
		want       []telegraf.Metric
		now        func() time.Time
		wantErr    bool
	}{
		{
			name:  "Log parser fmt returns all fields",
			now:   func() time.Time { return time.Unix(0, 0) },
			input: MakeInputMetrics(16),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Stats{
				WindowSize: 4,
				StatsField: "sample_field",
			}
			got := s.Apply(tt.input...)
			if got != nil {
				require.Equal(t, got, tt.want)
			}
		})
	}
}
