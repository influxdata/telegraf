package stats

import (
	"log"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Stats struct {
	StatsField string `toml:"stats_field"`
	WindowSize int    `toml:"window_size"`
}

type Window struct {
	Count    int
	Sum      float64
	Mean     float64
	Variance float64
	Std      float64
}

func (s *Stats) SampleConfig() string {
	return `
[[processors.stats]]

## field to compile a running average of
stats_field = "trace_id"`
}

func (s *Stats) Description() string {
	return "will append a field to each metric indicating the running average of the specified field"
}

func (s *Stats) Apply(in ...telegraf.Metric) []telegraf.Metric {
	var window Window
	var buffer Window
	for _, metric := range in {
		value, err := strconv.ParseFloat(metric.Fields()[s.StatsField].(string), 64)
		if err != nil {
			log.Printf("E! %v", err)
			continue
		}

		// when window count is more than half, fill the buffer and window
		if window.Count >= s.WindowSize/2 {
			window.Count++
			buffer.Count++

			window.Sum += value
			buffer.Sum += value
		}
	}
	return in
}

func init() {
	processors.Add("stats", func() telegraf.Processor {
		return &Stats{}
	})
}
