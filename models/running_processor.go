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

// ProcessorConfig containing a name and filter
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

func (rp *RunningProcessor) Init() error {
	if p, ok := rp.Processor.(telegraf.Initializer); ok {
		err := p.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

func (rp *RunningProcessor) Log() telegraf.Logger {
	return rp.log
}

func (rp *RunningProcessor) LogName() string {
	return logName("processors", rp.Config.Name, rp.Config.Alias)
}

func (rp *RunningProcessor) MakeMetric(metric telegraf.Metric) telegraf.Metric {
	return metric
}

func (rp *RunningProcessor) Start(acc telegraf.Accumulator) error {
	return rp.Processor.Start(acc)
}

func (rp *RunningProcessor) Add(m telegraf.Metric, acc telegraf.Accumulator) error {
	if ok := rp.Config.Filter.Select(m); !ok {
		// pass downstream
		acc.AddMetric(m)
		return nil
	}

	rp.Config.Filter.Modify(m)
	if len(m.FieldList()) == 0 {
		// drop metric
		rp.metricFiltered(m)
		return nil
	}

	return rp.Processor.Add(m, acc)
}

func (rp *RunningProcessor) Stop() {
	rp.Processor.Stop()
}
