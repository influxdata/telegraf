package deviation

import (
	"log"
	"math"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Deviation struct {
	DeviationField string `toml:"deviation_field"`
	Count          int
	Sum            float64
}

func (d *Deviation) SampleConfig() string {
	return `
## must run metrics through average processor before this processor
[[processors.deviation]]

## field to compile a standard deviation of
## the processor will assume the average of the field
## can be found in the field deviation_field"_mean"
deviation_field = "trace_id"`
}

func (d *Deviation) Description() string {
	return "will append a field to each metric indicating the running average of the specified field"
}

func (d *Deviation) Apply(in ...telegraf.Metric) []telegraf.Metric {
	nMetrics := make([]telegraf.Metric, 0)
	for _, metric := range in {
		if metric.Fields()[d.DeviationField] != nil {
			fVal, err := strconv.ParseFloat(metric.Fields()[d.DeviationField].(string), 64)
			if err != nil {
				log.Printf("E! %v must be a float or integer value, %v", d.DeviationField, err)
				continue
			}
			if metric.Fields()[d.DeviationField+"_mean"] != nil {
				log.Printf("E! missing field %v from [processor.average]", d.DeviationField+"_mean")
				continue
			}
			aveVal, err := strconv.ParseFloat(metric.Fields()[d.DeviationField+"_mean"].(string), 64)
			if err != nil {
				log.Printf("E! %v must be a float or integer value, %v", d.DeviationField+"_mean", err)
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
	processors.Add("deviation", func() telegraf.Processor {
		return &Deviation{}
	})
}
