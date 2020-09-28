package models

import (
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
)

type RunningProcessor struct {
	sync.Mutex
	log       telegraf.Logger
	Processor telegraf.StreamingProcessor
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

func NewRunningProcessor(processor telegraf.StreamingProcessor, config *ProcessorConfig) *RunningProcessor {
	tags := map[string]string{"processor": config.Name}
	if config.Alias != "" {
		tags["alias"] = config.Alias
	}

	processErrorsRegister := selfstat.Register("process", "errors", tags)
	logger := NewLogger("processors", config.Name, config.Alias)
	logger.OnErr(func() {
		processErrorsRegister.Incr(1)
	})
	SetLoggerOnPlugin(processor, logger)

	return &RunningProcessor{
		Processor: processor,
		Config:    config,
		log:       logger,
	}
}

func (rp *RunningProcessor) metricFiltered(metric telegraf.Metric) {
	metric.Drop()
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

func (r *RunningProcessor) Log() telegraf.Logger {
	return r.log
}

func (r *RunningProcessor) LogName() string {
	return logName("processors", r.Config.Name, r.Config.Alias)
}

func (r *RunningProcessor) MakeMetric(metric telegraf.Metric) telegraf.Metric {
	return metric
}

func (r *RunningProcessor) Start(acc telegraf.Accumulator) error {
	return r.Processor.Start(acc)
}

func (r *RunningProcessor) Add(m telegraf.Metric, acc telegraf.Accumulator) error {
	if ok := r.Config.Filter.Select(m); !ok {
		// pass downstream
		acc.AddMetric(m)
		return nil
	}

	r.Config.Filter.Modify(m)
	if len(m.FieldList()) == 0 {
		// drop metric
		r.metricFiltered(m)
		return nil
	}

	return r.Processor.Add(m, acc)
}

func (r *RunningProcessor) Stop() {
	r.Processor.Stop()
}
