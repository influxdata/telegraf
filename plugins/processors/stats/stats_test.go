package stats

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

func MakeOutputMetric(num int, dev float64, variance float64, mean float64) telegraf.Metric {
	m, _ := metric.New("test", map[string]string{}, map[string]interface{}{"sample": num,
		"sample_deviation": dev,
		"sample_variance":  variance,
		"sample_mean":      mean}, time.Now())
	return m
}

func MakeAllOutputMetrics(count int) []telegraf.Metric {
	outputs := MakeInputMetrics(2)
	type Stat struct {
		val      int
		dev      float64
		variance float64
		mean     float64
	}

	testGoodies := []Stat{
		{
			val:      3,
			dev:      0.7905694150420949,
			variance: 0.625,
			mean:     2,
		},
		{
			val:      4,
			dev:      1.0801234497346435,
			variance: 1.1666666666666667,
			mean:     2.5,
		},
		{
			val:      5,
			dev:      0.7905694150420949,
			variance: 0.625,
			mean:     4,
		},
		{
			val:      6,
			dev:      1.0801234497346435,
			variance: 1.1666666666666667,
			mean:     4.5,
		},
		{
			val:      7,
			dev:      0.7905694150420949,
			variance: 0.625,
			mean:     6,
		},
		{
			val:      8,
			dev:      1.0801234497346435,
			variance: 1.1666666666666667,
			mean:     6.5,
		},
		{
			val:      9,
			dev:      0.7905694150420949,
			variance: 0.625,
			mean:     8,
		},
		{
			val:      10,
			dev:      1.0801234497346435,
			variance: 1.1666666666666667,
			mean:     8.5,
		},
		{
			val:      11,
			dev:      0.7905694150420949,
			variance: 0.625,
			mean:     10,
		},
		{
			val:      12,
			dev:      1.0801234497346435,
			variance: 1.1666666666666667,
			mean:     10.5,
		},
		{
			val:      13,
			dev:      0.7905694150420949,
			variance: 0.625,
			mean:     12,
		},
		{
			val:      14,
			dev:      1.0801234497346435,
			variance: 1.1666666666666667,
			mean:     12.5,
		},
		{
			val:      15,
			dev:      0.7905694150420949,
			variance: 0.625,
			mean:     14,
		},
		{
			val:      16,
			dev:      1.0801234497346435,
			variance: 1.1666666666666667,
			mean:     14.5,
		},
	}

	for _, s := range testGoodies {
		outputs = append(outputs, MakeOutputMetric(s.val, s.dev, s.variance, s.mean))
	}

	return outputs
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
