package stats

import (
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
stats_field = "trace_id"`
}

func (s *Stats) Description() string {
	return "will append a field to each metric indicating the running average of the specified field"
}

func (s *Stats) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		value, err := strconv.ParseFloat(metric.Fields()[s.StatsField].(string), 64)
		if err != nil {
			log.Printf("E! %v", err)
			continue
		}

		// for warmup
		if s.Window.Count < s.WindowSize/2 {
			s.Window.Count++
			s.Window.ValueSum += value
			s.Window.Mean = s.Window.ValueSum / float64(s.Window.Count)
			s.Window.Std = stdCalculator(value, s.Window)
			s.Window.Variance = math.Pow(s.Window.Std, 2)
			continue
		}

		if s.Window.Count >= s.WindowSize {
			s.Window = s.Buffer
			s.Buffer = Window{}
		}

		// when window count is more than half, fill the buffer and window
		s.Window.Count++
		s.Buffer.Count++

		s.Window.ValueSum += value
		s.Buffer.ValueSum += value

		s.Window.Mean = s.Window.ValueSum / float64(s.Window.Count)
		s.Buffer.Mean = s.Buffer.ValueSum / float64(s.Buffer.Count)

		s.Window.Std = stdCalculator(value, s.Window)
		s.Buffer.Std = stdCalculator(value, s.Buffer)

		s.Window.Variance = math.Pow(s.Window.Std, 2)
		s.Buffer.Variance = math.Pow(s.Buffer.Std, 2)

		metric.AddField(s.StatsField+"_mean", s.Window.Mean)
		metric.AddField(s.StatsField+"_deviation", s.Window.Std)
		metric.AddField(s.StatsField+"_variance", s.Window.Variance)
	}
	return in
}

func stdCalculator(currentVal float64, w Window) float64 {
	diff := currentVal - w.Mean
	sqrDiff := math.Pow(diff, 2)
	w.DiffSum += sqrDiff
	std := math.Sqrt(w.DiffSum / float64(w.Count))
	return std
}

func init() {
	processors.Add("stats", func() telegraf.Processor {
		return &Stats{}
	})
}
