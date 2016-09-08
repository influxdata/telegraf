package models

import (
	"github.com/influxdata/telegraf"
)

type RunningProcessor struct {
	Name      string
	Processor telegraf.Processor
	Config    *ProcessorConfig
}

// FilterConfig containing a name and filter
type ProcessorConfig struct {
	Name   string
	Filter Filter
}

func (rp *RunningProcessor) Apply(in ...telegraf.Metric) []telegraf.Metric {
	ret := []telegraf.Metric{}

	for _, metric := range in {
		if rp.Config.Filter.IsActive() {
			// check if the filter should be applied to this metric
			if ok := rp.Config.Filter.Apply(metric.Name(), metric.Fields(), metric.Tags()); !ok {
				// this means filter should not be applied
				ret = append(ret, metric)
				continue
			}
		}
		// This metric should pass through the filter, so call the filter Apply
		// function and append results to the output slice.
		ret = append(ret, rp.Processor.Apply(metric)...)
	}

	return ret
}
