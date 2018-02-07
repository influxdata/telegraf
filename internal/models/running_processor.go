package models

import (
	"sync"

	"github.com/influxdata/telegraf"
)

type RunningProcessor struct {
	Name string

	sync.Mutex
	Processor telegraf.Processor
	Config    *ProcessorConfig
}

type RunningProcessors []*RunningProcessor

func (rp RunningProcessors) Len() int           { return len(rp) }
func (rp RunningProcessors) Swap(i, j int)      { rp[i], rp[j] = rp[j], rp[i] }
func (rp RunningProcessors) Less(i, j int) bool { return rp[i].Config.Order < rp[j].Config.Order }

// FilterConfig containing a name and filter
type ProcessorConfig struct {
	Name   string
	Order  int64
	Filter Filter
}

func (rp *RunningProcessor) Apply(in ...telegraf.Metric) []telegraf.Metric {
	rp.Lock()
	defer rp.Unlock()

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
