package average

import (
	"log"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Average struct {
	AverageField string `toml:"average_field"`
	Count        int
	Sum          float64
}

func (s *Average) SampleConfig() string {
	return `
[[processors.average]]

## field to compile a running average of
average_field = "trace_id"`
}

func (a *Average) Description() string {
	return "will append a field to each metric indicating the running average of the specified field"
}

func (a *Average) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		if metric.Fields()[a.AverageField] != nil {

			fVal, err := strconv.ParseFloat(metric.Fields()[a.AverageField].(string), 64)
			if err != nil {
				log.Printf("E! %v must be a float or integer value, %v", a.AverageField, err)
				continue
			}
			a.Sum += fVal
			a.Count++
			ave := a.Sum / float64(a.Count)
			metric.AddField(a.AverageField+"_mean", ave)
		}
	}
	return in
}

func init() {
	processors.Add("average", func() telegraf.Processor {
		return &Average{}
	})
}
