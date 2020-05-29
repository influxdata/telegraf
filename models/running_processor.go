package models

import (
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
)

type RunningProcessor struct {
	sync.Mutex
	log       telegraf.Logger
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
	Alias  string
	Order  int64
	Filter Filter
}

func NewRunningProcessor(processor telegraf.Processor, config *ProcessorConfig) *RunningProcessor {
	tags := map[string]string{"processor": config.Name}
	if config.Alias != "" {
		tags["alias"] = config.Alias
	}

	processErrorsRegister := selfstat.Register("process", "errors", tags)
	logger := NewLogger("processors", config.Name, config.Alias)
	logger.OnErr(func() {
		processErrorsRegister.Incr(1)
	})
	setLogIfExist(processor, logger)

	return &RunningProcessor{
		Processor: processor,
		Config:    config,
		log:       logger,
	}
}

func (rp *RunningProcessor) metricFiltered(metric telegraf.Metric) {
	metric.Drop()
}

func containsMetric(item telegraf.Metric, metrics []telegraf.Metric) bool {
	for _, m := range metrics {
		if item == m {
			return true
		}
	}
	return false
}

func (r *RunningProcessor) Init() error {
	if p, ok := r.Processor.(telegraf.Initializer); ok {
		err := p.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

func (rp *RunningProcessor) Apply(in ...telegraf.Metric) []telegraf.Metric {
	rp.Lock()
	defer rp.Unlock()

	ret := []telegraf.Metric{}

	for _, metric := range in {
		// In processors when a filter selects a metric it is sent through the
		// processor.  Otherwise the metric continues downstream unmodified.
		if ok := rp.Config.Filter.Select(metric); !ok {
			ret = append(ret, metric)
			continue
		}

		rp.Config.Filter.Modify(metric)
		if len(metric.FieldList()) == 0 {
			rp.metricFiltered(metric)
			continue
		}

		// This metric should pass through the filter, so call the filter Apply
		// function and append results to the output slice.
		ret = append(ret, rp.Processor.Apply(metric)...)
	}

	return ret
}

func (r *RunningProcessor) Log() telegraf.Logger {
	return r.log
}
