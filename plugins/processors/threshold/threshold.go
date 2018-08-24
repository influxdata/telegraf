package threshold

import (
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Threshold struct {
	FieldName       string  `toml:"field_name"`
	OutlierDistance float64 `toml:"outlier_distance"`
}

func (d *Threshold) SampleConfig() string {
	return `
## must run metrics through average processor before this processor
[[processors.threshold]]

## field to compile a standard deviation of
## the processor will assume the average of the field
## can be found in the field deviation_field"_mean"
field_name = "trace_id"

## Determine the number of standard deviations
## away you want your outlier to be
outlier_distance = "2"`
}

func (d *Threshold) Description() string {
	return "will append a field to each metric indicating whether it is a outlier or not"
}

func (d *Threshold) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		if metric.Fields()[d.FieldName] != nil {
			fVal, err := strconv.ParseFloat(fmt.Sprintf("%v", metric.Fields()[d.FieldName]), 64)
			if err != nil {
				log.Printf("E! %v must be a float or integer value, %v", d.FieldName, err)
				continue
			}
			if metric.Fields()[d.FieldName+"_mean"] == nil {
				log.Printf("E! missing field: %v from [processor.stats]", d.FieldName+"_mean")
				continue
			}
			if metric.Fields()[d.FieldName+"_deviation"] == nil {
				log.Printf("E! missing field: %v from [processor.stats]", d.FieldName+"_deviation")
				continue
			}
			mean, err := strconv.ParseFloat(fmt.Sprintf("%v", metric.Fields()[d.FieldName+"_mean"]), 64)
			if err != nil {
				log.Printf("E! %v must be a float or integer value, %v", d.FieldName+"_mean", err)
				continue
			}

			deviation, err := strconv.ParseFloat(fmt.Sprintf("%v", metric.Fields()[d.FieldName+"_deviation"]), 64)
			if err != nil {
				log.Printf("E! %v must be a float or integer value, %v", d.FieldName+"_deviation", err)
				continue
			}

			if math.Abs(mean-fVal) >= d.OutlierDistance*deviation {
				numOutliers := math.Abs(mean-fVal) / d.OutlierDistance

				// adds a field to mark outliers
				metric.AddField("stddev_away", numOutliers)
			}
		}
	}
	return in
}

func init() {
	processors.Add("threshold", func() telegraf.Processor {
		return &Threshold{}
	})
}
