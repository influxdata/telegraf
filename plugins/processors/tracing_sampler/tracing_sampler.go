package tracing_sampler

import (
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/processors/sampler"
	"github.com/influxdata/telegraf/plugins/processors/stats"
	"github.com/influxdata/telegraf/plugins/processors/threshold"
)

type TracingSampler struct {
	WindowSize       int     `toml:"window_size"`
	StatsField       string  `toml:"stats_field"`
	SampleField      string  `toml:"sample_field"`
	OutlierDistance  float64 `toml:"outlier_distance"`
	PercentOfMetrics int     `toml:"percent_of_metrics"`
	Stats            stats.Stats
	Threshold        threshold.Threshold
	Sampler          sampler.Sampler
	compiled         bool
}

func (a *TracingSampler) SampleConfig() string {
	return `
[[processors.allofem]]

## percent of total metrics to be returned as a sample
percent_of_metrics = 5

## field to have stats compiled for
stats_field = "trace_id"

## number of metrics considered for stats
## to be calculated
window_size = 6

## Determine the number of standard deviations
## away you want your outlier to be from the mean
outlier_distance = 2`
}

func (a *TracingSampler) Description() string {
	return "will add mean, variance, and standard deviation stats to metrics, mark outliers in the data set, and return a sample percentage of metrics along with any outliers"
}

func (a *TracingSampler) compile() error {
	if a.StatsField == "" {
		return fmt.Errorf("[processor.allofem] stats_field must be set")
	}
	if a.WindowSize <= 0 {
		return fmt.Errorf("[processor.allofem] window_size is invalid, cannot be zero or negative ")
	}
	if a.OutlierDistance == 0 {
		return fmt.Errorf("[processor.allofem] outlier_distance is invalid, must be more than 0")
	}
	if a.PercentOfMetrics < 0 {
		return fmt.Errorf("[processor.allofem] percent_of_metrics can't be negative")
	}

	a.Stats = stats.Stats{
		StatsField: a.StatsField,
		WindowSize: a.WindowSize,
	}
	a.Threshold = threshold.Threshold{
		FieldName:       a.StatsField,
		OutlierDistance: a.OutlierDistance,
	}
	a.Sampler = sampler.Sampler{
		SampleField:      a.StatsField,
		PercentOfMetrics: a.PercentOfMetrics,
	}
	a.compiled = true
	return nil
}

func (a *TracingSampler) Apply(in ...telegraf.Metric) []telegraf.Metric {
	if !a.compiled {
		a.compile()
	}
	in = a.Stats.Apply(in...)
	in = a.Threshold.Apply(in...)
	in = a.Sampler.Apply(in...)
	return in
}

func init() {
	processors.Add("tracing_sampler", func() telegraf.Processor {
		return &TracingSampler{}
	})
}
