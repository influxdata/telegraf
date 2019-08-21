package models

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
)

var GlobalMetricsGathered = selfstat.Register("agent", "metrics_gathered", map[string]string{})

type RunningInput struct {
	Input  telegraf.Input
	Config *InputConfig

	defaultTags map[string]string

	MetricsGathered selfstat.Stat
	GatherTime      selfstat.Stat
}

func NewRunningInput(input telegraf.Input, config *InputConfig) *RunningInput {
	return &RunningInput{
		Input:  input,
		Config: config,
		MetricsGathered: selfstat.Register(
			"gather",
			"metrics_gathered",
			map[string]string{"input": config.Name, "alias": config.Alias},
		),
		GatherTime: selfstat.RegisterTiming(
			"gather",
			"gather_time_ns",
			map[string]string{"input": config.Name, "alias": config.Alias},
		),
	}
}

// InputConfig is the common config for all inputs.
type InputConfig struct {
	Name     string
	Alias    string
	Interval time.Duration

	NameOverride      string
	MeasurementPrefix string
	MeasurementSuffix string
	Tags              map[string]string
	Filter            Filter
}

func (r *RunningInput) Name() string {
	return "inputs." + r.Config.Name
}

func (r *RunningInput) LogName() string {
	if r.Config.Alias == "" {
		return r.Name()
	}
	return r.Name() + "::" + r.Config.Alias
}

func (r *RunningInput) metricFiltered(metric telegraf.Metric) {
	metric.Drop()
}

func (r *RunningInput) Init() error {
	setLogIfExist(r.Input, &Logger{
		Name: r.LogName(),
		Errs: selfstat.Register("gather", "errors",
			map[string]string{"input": r.Config.Name, "alias": r.Config.Alias}),
	})

	if p, ok := r.Input.(telegraf.Initializer); ok {
		err := p.Init()
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RunningInput) MakeMetric(metric telegraf.Metric) telegraf.Metric {
	if ok := r.Config.Filter.Select(metric); !ok {
		r.metricFiltered(metric)
		return nil
	}

	m := makemetric(
		metric,
		r.Config.NameOverride,
		r.Config.MeasurementPrefix,
		r.Config.MeasurementSuffix,
		r.Config.Tags,
		r.defaultTags)

	r.Config.Filter.Modify(metric)
	if len(metric.FieldList()) == 0 {
		r.metricFiltered(metric)
		return nil
	}

	r.MetricsGathered.Incr(1)
	GlobalMetricsGathered.Incr(1)
	return m
}

func (r *RunningInput) Gather(acc telegraf.Accumulator) error {
	start := time.Now()
	err := r.Input.Gather(acc)
	elapsed := time.Since(start)
	r.GatherTime.Incr(elapsed.Nanoseconds())
	return err
}

func (r *RunningInput) SetDefaultTags(tags map[string]string) {
	r.defaultTags = tags
}
