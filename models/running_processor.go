package models

import (
	"fmt"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
)

type RunningProcessor struct {
	sync.Mutex
	log telegraf.Logger
	// Processor can be a Processor or a StreamingProcessor
	Processor telegraf.PluginDescriber
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

func NewRunningProcessor(processor telegraf.PluginDescriber, config *ProcessorConfig) *RunningProcessor {
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

func (r *RunningProcessor) ApplyFilters(metric telegraf.Metric) telegraf.Metric {
	if ok := r.Config.Filter.Select(metric); !ok {
		r.metricFiltered(metric)
		return nil
	}

	r.Config.Filter.Modify(metric)
	if len(metric.FieldList()) == 0 {
		r.metricFiltered(metric)
		return nil
	}

	return metric
}

func (r *RunningProcessor) Log() telegraf.Logger {
	return r.log
}

func (r *RunningProcessor) Start(acc telegraf.StreamingAccumulator) error {
	if sp, ok := r.Processor.(telegraf.StreamingProcessor); ok {
		return sp.Start(acc)
	} else if p, ok := r.Processor.(telegraf.Processor); ok {
		// wrap a standard processor to work like a streaming one.
		sp := telegraf.NewProcessorWrapper(p)
		return sp.Start(acc)
	} else {
		// unknown processor type.
		return fmt.Errorf("Unknown processor type %T", r.Processor)
	}
	return nil
}

type metricModifierFn func(m telegraf.Metric) telegraf.Metric

// NewStreamingAccumulatorWrapper wraps a streaming accumulator, calling a function
// on each metric that passes through the stream.
func NewStreamingAccumulatorWrapper(acc telegraf.StreamingAccumulator, f metricModifierFn) telegraf.StreamingAccumulator {
	sa := &streamingAccumulatorWrapper{
		acc: acc,
		f:   f,
	}
	return sa
}

type streamingAccumulatorWrapper struct {
	acc telegraf.StreamingAccumulator
	f   func(telegraf.Metric) telegraf.Metric
}

func (sa *streamingAccumulatorWrapper) PassMetric(m telegraf.Metric) {
	sa.acc.PassMetric(m)
}

func (sa *streamingAccumulatorWrapper) GetNextMetric() telegraf.Metric {
retry:
	m := sa.acc.GetNextMetric()
	if m == nil {
		if sa.acc.IsStreamClosed() {
			return nil
		}
		goto retry
	}
	m2 := sa.f(m)
	if m2 == nil {
		// metric was dropped, pass through to next layer
		sa.acc.PassMetric(m)
		goto retry
	}
	return m
}
func (sa *streamingAccumulatorWrapper) IsMetricAvailable() bool {
	return sa.acc.IsMetricAvailable()
}
func (sa *streamingAccumulatorWrapper) IsStreamClosed() bool {
	return sa.acc.IsStreamClosed()
}
