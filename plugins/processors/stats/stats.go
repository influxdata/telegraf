package stats

import (
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Stats struct {
	StatsField string `toml:"stats_field"`
	WindowSize int    `toml:"window_size"`
	Window     Window
	Buffer     Window
}

type Window struct {
	Count    int
	ValueSum float64
	DiffSum  float64
	Mean     float64
	Variance float64
	Std      float64
}

func (s *Stats) SampleConfig() string {
	return `
[[processors.stats]]

## field to compile stats for
stats_field = "trace_id"

## number of metrics considered for stats at a time
window_size = 8`
}

func (s *Stats) Description() string {
	return "will append a field to each metric indicating the running average of the specified field"
}

func (s *Stats) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		value := metric.Fields()[s.StatsField]
		sVal := fmt.Sprintf("%v", value)
		fVal, err := strconv.ParseFloat(sVal, 64)
		if err != nil {
			log.Printf("E! cannot convert field: %v to float, %v", s.StatsField, err)
		}

		// for warmup
		if s.Window.Count < s.WindowSize/2 {
			s.Window.Count++
			s.Window.ValueSum += fVal
			s.Window.Mean = s.Window.ValueSum / float64(s.Window.Count)
			s.Window.Variance = varianceCalculator(fVal, &s.Window)
			s.Window.Std = math.Sqrt(s.Window.Variance)
			continue
		}

		if s.Window.Count >= s.WindowSize {
			s.Window = s.Buffer
			s.Buffer = Window{}
		}

		// when window count is more than half, fill the buffer and window
		s.Window.Count++
		s.Buffer.Count++

		s.Window.ValueSum += fVal
		s.Buffer.ValueSum += fVal

		s.Window.Mean = s.Window.ValueSum / float64(s.Window.Count)
		s.Buffer.Mean = s.Buffer.ValueSum / float64(s.Buffer.Count)

		s.Window.Variance = varianceCalculator(fVal, &s.Window)
		s.Buffer.Variance = varianceCalculator(fVal, &s.Buffer)

		s.Window.Std = math.Sqrt(s.Window.Variance)
		s.Buffer.Std = math.Sqrt(s.Buffer.Variance)

		metric.AddField(s.StatsField+"_mean", s.Window.Mean)
		metric.AddField(s.StatsField+"_deviation", s.Window.Std)
		metric.AddField(s.StatsField+"_variance", s.Window.Variance)
	}
	return in
}

func varianceCalculator(currentVal float64, w *Window) float64 {
	diff := currentVal - w.Mean
	sqrDiff := math.Pow(diff, 2)
	w.DiffSum = w.DiffSum + sqrDiff
	if w.Count-1 == 0 {
		return 0
	}
	variance := w.DiffSum / float64(w.Count-1)
	return variance
}

func init() {
	processors.Add("stats", func() telegraf.Processor {
		return &Stats{}
	})
}
