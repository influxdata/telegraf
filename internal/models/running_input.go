package models

import (
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
)

type RunningInput struct {
	Input  telegraf.Input
	Config *InputConfig

	trace       bool
	debug       bool
	defaultTags map[string]string
}

// InputConfig containing a name, interval, and filter
type InputConfig struct {
	Name              string
	NameOverride      string
	MeasurementPrefix string
	MeasurementSuffix string
	Tags              map[string]string
	Filter            Filter
	Interval          time.Duration
}

func (r *RunningInput) Name() string {
	return "inputs." + r.Config.Name
}

// MakeMetric either returns a metric, or returns nil if the metric doesn't
// need to be created (because of filtering, an error, etc.)
func (r *RunningInput) MakeMetric(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	mType telegraf.ValueType,
	t time.Time,
) telegraf.Metric {
	m := makemetric(
		measurement,
		fields,
		tags,
		r.Config.NameOverride,
		r.Config.MeasurementPrefix,
		r.Config.MeasurementSuffix,
		r.Config.Tags,
		r.defaultTags,
		r.Config.Filter,
		true,
		r.debug,
		mType,
		t,
	)

	if r.trace && m != nil {
		fmt.Println("> " + m.String())
	}

	return m
}

func (r *RunningInput) Debug() bool {
	return r.debug
}

func (r *RunningInput) SetDebug(debug bool) {
	r.debug = debug
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
