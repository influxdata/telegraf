package threshold

import (
	"log"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/require"
)

intArray := []int{1,2,3,4,5,6,7,8,8,9,10,100,50}

func MakeInputMetrics(num []int) []telegraf.Metric {
	allOfEm := make([]telegraf.Metric, 0)
	for i := range num {
		m, _ := metric.New("test", map[string]string{}, map[string]interface{}{"sample": i}, time.Now())
		allOfEm = append(allOfEm, m)
	}
	return allOfEm
}

func MakeOutputMetric(num int, dev float64, variance float64, mean float64) telegraf.Metric {
	m, _ := metric.New("test", map[string]string{}, map[string]interface{}{"sample": num,
		"sample_deviation": dev,
		"sample_variance":  variance,
		"sample_mean":      mean}, time.Now())
	return m
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
			name:  "Test deviation, mean, and variant fields are added",
			now:   func() time.Time { return time.Unix(0, 0) },
			input: MakeInputMetrics(16),
			want:  MakeAllOutputMetrics(16)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Stats{
				WindowSize: 4,
				StatsField: "sample",
			}
			got := s.Apply(tt.input...)
			for i, m := range got {
				std := m.Fields()["sample_deviation"]
				variance := m.Fields()["sample_variance"]
				mean := m.Fields()["sample_mean"]
				log.Printf("m[%v]: std: %v, var: %v, mean: %v", i, std, variance, mean)
			}
			for i, m := range tt.want {
				require.Equal(t, m.Fields(), got[i].Fields())
			}
		})
	}
}
