package rename_eval

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
`

// Rename - renames
type Rename struct {
	Tag          string `toml:"tag"`
	Dest         string `toml:"dest"`
	DropOriginal bool   `toml:"drop_original"`
	Position     int    `toml:"prefix=1 , postfix=2, 0=replace"`
	init         bool
	evaluate     func(string, string, telegraf.Metric) string
}

// SampleConfig - SampleConfig
func (r *Rename) SampleConfig() string {
	return sampleConfig
}

//Description - Description
func (r *Rename) Description() string {
	return "Rename measurements, tags, and fields that pass through this filter."
}

func (r *Rename) initOnce() {
	if r.init {
		return
	}
	r.evaluate = suffixPrefixEvaluator(r.Position)
	r.init = true
}

func suffixPrefixEvaluator(s int) func(string, string, telegraf.Metric) string {
	switch s {
	case 1:
		return func(tag string, measurement string, point telegraf.Metric) string {
			if value, ok := point.GetTag(tag); ok {
				return value + "_" + point.Name()
			}
			return point.Name()

		}
	case 2:
		return func(tag string, measurement string, point telegraf.Metric) string {
			if value, ok := point.GetTag(tag); ok {
				return point.Name() + "_" + value
			}
			return point.Name()
		}
	default:
		return func(tag string, measurement string, point telegraf.Metric) string {
			return measurement
		}
	}
}

//Apply - Apply
func (r *Rename) Apply(in ...telegraf.Metric) []telegraf.Metric {
	r.initOnce()
	results := []telegraf.Metric{}

	for _, point := range in {
		if !r.DropOriginal {
			results = append(results, point)
			point = metric.FromMetric(point)
		}
		newName := r.evaluate(r.Tag, r.Dest, point)
		if newName != point.Name() {
			point.SetName(newName)
			results = append(results, point)
		}
	}

	return results
}

func init() {
	processors.Add("rename_eval", func() telegraf.Processor {
		return &Rename{}
	})
}
