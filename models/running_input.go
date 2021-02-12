package models

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
)

var (
	GlobalMetricsGathered = selfstat.Register("agent", "metrics_gathered", map[string]string{})
	GlobalGatherErrors    = selfstat.Register("agent", "gather_errors", map[string]string{})
)

type RunningInput struct {
	ID     uint64
	Input  telegraf.Input
	Config *InputConfig

	log         telegraf.Logger
	defaultTags map[string]string

	MetricsGathered selfstat.Stat
	GatherTime      selfstat.Stat
	State
}

func NewRunningInput(input telegraf.Input, config *InputConfig) *RunningInput {
	tags := map[string]string{"input": config.Name}
	if config.Alias != "" {
		tags["alias"] = config.Alias
	}

	inputErrorsRegister := selfstat.Register("gather", "errors", tags)
	logger := NewLogger("inputs", config.Name, config.Alias)
	logger.OnErr(func() {
		inputErrorsRegister.Incr(1)
		GlobalGatherErrors.Incr(1)
	})
	SetLoggerOnPlugin(input, logger)

	return &RunningInput{
		ID:     nextPluginID(),
		Input:  input,
		Config: config,
		MetricsGathered: selfstat.Register(
			"gather",
			"metrics_gathered",
			tags,
		),
		GatherTime: selfstat.RegisterTiming(
			"gather",
			"gather_time_ns",
			tags,
		),
		log: logger,
	}
}

// InputConfig is the common config for all inputs.
type InputConfig struct {
	Name             string        `toml:"name" json:"name"`
	Alias            string        `toml:"alias" json:"alias"`
	Interval         time.Duration `toml:"interval" json:"interval"`
	CollectionJitter time.Duration `toml:"collection_jitter" json:"collection_jitter"`
	Precision        time.Duration `toml:"precision" json:"precision"`

	NameOverride      string            `toml:"name_override" json:"name_override"`
	MeasurementPrefix string            `toml:"measurement_prefix" json:"measurement_prefix"`
	MeasurementSuffix string            `toml:"measurement_suffix" json:"measurement_suffix"`
	Tags              map[string]string `toml:"tags" json:"tags"`
	Filter            Filter            `toml:"filter" json:"filter"`
}

func (r *RunningInput) metricFiltered(metric telegraf.Metric) {
	metric.Drop()
}

func (r *RunningInput) LogName() string {
	return logName("inputs", r.Config.Name, r.Config.Alias)
}

func (r *RunningInput) Init() error {
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

func (r *RunningInput) Log() telegraf.Logger {
	return r.log
}

func (r *RunningInput) Start(acc telegraf.Accumulator) error {
	r.setState(PluginStateStarting)
	if si, ok := r.Input.(telegraf.ServiceInput); ok {
		if err := si.Start(acc); err != nil {
			return err
		}
	}
	r.setState(PluginStateRunning)
	return nil
}

func (r *RunningInput) Stop() {
	r.setState(PluginStateStopping)
	if si, ok := r.Input.(telegraf.ServiceInput); ok {
		si.Stop()
	}
	r.setState(PluginStateDead)
}

func (r *RunningInput) GetID() uint64 {
	return r.ID
}
