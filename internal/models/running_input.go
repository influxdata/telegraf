package models

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/selfstat"
)

var GlobalMetricsGathered = selfstat.Register("agent", "metrics_gathered", map[string]string{})

type RunningInput struct {
	Input  telegraf.Input
	Config *InputConfig

	trace       bool
	defaultTags map[string]string

	MetricsGathered selfstat.Stat
}

func NewRunningInput(
	input telegraf.Input,
	config *InputConfig,
) *RunningInput {
	return &RunningInput{
		Input:  input,
		Config: config,
		MetricsGathered: selfstat.Register(
			"gather",
			"metrics_gathered",
			map[string]string{"input": config.Name},
		),
	}
}

// InputConfig is the common config for all inputs.
type InputConfig struct {
	Name     string
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

func (r *RunningInput) MakeMetric(metric telegraf.Metric) telegraf.Metric {
	if ok := r.Config.Filter.Select(metric); !ok {
		return nil
	}

	r.Config.Filter.Modify(metric)
	if len(metric.FieldList()) == 0 {
		return nil
	}

	m := makemetric(
		metric,
		r.Config.NameOverride,
		r.Config.MeasurementPrefix,
		r.Config.MeasurementSuffix,
		r.Config.Tags,
		r.defaultTags)

	if r.trace && m != nil {
		s := influx.NewSerializer()
		s.SetFieldSortOrder(influx.SortFields)
		octets, err := s.Serialize(m)
		if err == nil {
			fmt.Print("> " + string(octets))
		}
	}

	r.MetricsGathered.Incr(1)
	GlobalMetricsGathered.Incr(1)
	return m
}

func (r *RunningInput) Trace() bool {
	return r.trace
}

func (r *RunningInput) SetTrace(trace bool) {
	r.trace = trace
}

func (r *RunningInput) SetDefaultTags(tags map[string]string) {
	r.defaultTags = tags
}
