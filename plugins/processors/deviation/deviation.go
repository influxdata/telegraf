package threshold

import (
	"log"
	"math"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Threshold struct {
	FieldName         string  `toml:"field_name"`
	DeviationDistance float64 `toml:"outlier_distance"`
	Count             int
	Sum               float64
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
	nMetrics := make([]telegraf.Metric, 0)
	for _, metric := range in {
		if metric.Fields()[d.FieldName] != nil {
			fVal, err := strconv.ParseFloat(metric.Fields()[d.FieldName].(string), 64)
			if err != nil {
				log.Printf("E! %v must be a float or integer value, %v", d.FieldName, err)
				continue
			}
			if metric.Fields()[d.FieldName+"_mean"] != nil {
				log.Printf("E! missing field %v from [processor.average]", d.FieldName+"_mean")
				continue
			}
			aveVal, err := strconv.ParseFloat(metric.Fields()[d.FieldName+"_mean"].(string), 64)
			if err != nil {
				log.Printf("E! %v must be a float or integer value, %v", d.FieldName+"_mean", err)
				continue
			}
			diff := fVal - aveVal
			sqrDiff := math.Pow(diff, 2)
			d.Sum += sqrDiff
			d.Count++
			std := math.Sqrt(d.Sum / float64(d.Count))

			if math.Abs(diff) > 2.0*std {
				nMetrics = append(nMetrics, metric)
			}

		}
	}
	return nMetrics
}

func init() {
	processors.Add("threshold", func() telegraf.Processor {
		return &Threshold{}
	})
}
