package models

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/pubsub"
	"github.com/influxdata/telegraf/selfstat"
)

type RunningService struct {
	Service telegraf.Service
	Config  *ServiceConfig

	defaultTags map[string]string

	MetricsGathered selfstat.Stat
	GatherTime      selfstat.Stat
}

func NewRunningService(service telegraf.Service, config *ServiceConfig) *RunningService {
	return &RunningService{
		Service: service,
		Config:  config,
		MetricsGathered: selfstat.Register(
			"gather",
			"metrics_gathered",
			map[string]string{"service": config.Name},
		),
		GatherTime: selfstat.RegisterTiming(
			"gather",
			"gather_time_ns",
			map[string]string{"service": config.Name},
		),
	}
}

// InputConfig is the common config for all inputs.
type ServiceConfig struct {
	Name     string
	Interval time.Duration

	NameOverride      string
	MeasurementPrefix string
	MeasurementSuffix string
	Tags              map[string]string
	Filter            Filter
}

func (r *RunningService) Name() string {
	return "service." + r.Config.Name
}

func (r *RunningService) metricFiltered(metric telegraf.Metric) {
	metric.Drop()
}

func (r *RunningService) MakeMetric(metric telegraf.Metric) telegraf.Metric {
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

func (r *RunningService) Run(msgbus *pubsub.PubSub) error {
	start := time.Now()
	err := r.Service.Run(msgbus)
	elapsed := time.Since(start)
	r.GatherTime.Incr(elapsed.Nanoseconds())
	return err
}

func (r *RunningService) Connect() error {
	err := r.Service.Connect()
	return err
}

func (r *RunningService) Close() error {
	err := r.Service.Close()
	return err
}
func (r *RunningService) SetDefaultTags(tags map[string]string) {
	r.defaultTags = tags
}
