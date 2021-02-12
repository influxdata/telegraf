package models

import (
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
)

type RunningProcessor struct {
	ID uint64
	sync.Mutex
	log       telegraf.Logger
	Processor telegraf.StreamingProcessor
	Config    *ProcessorConfig
	State
}

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

	p := &RunningProcessor{
		ID:        nextPluginID(),
		Processor: processor,
		Config:    config,
		log:       logger,
	}
	p.setState(PluginStateCreated)
	return p
}

func (rp *RunningProcessor) metricFiltered(metric telegraf.Metric) {
	//TODO(steve): rp.MetricsFiltered.Incr(1)
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
	r.setState(PluginStateStarting)
	err := r.Processor.Start(acc)
	if err != nil {
		return err
	}
	r.setState(PluginStateRunning)
	return nil
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
	r.setState(PluginStateStopping)
	r.Processor.Stop()
	r.setState(PluginStateDead)
}

func (r *RunningProcessor) Order() int64 {
	return r.Config.Order
}

func (r *RunningProcessor) GetID() uint64 {
	return r.ID
}
