package metric

import (
	"time"

	"github.com/influxdata/telegraf"
)

type TimeFunc func() time.Time

type Builder struct {
	TimeFunc
	TimePrecision time.Duration

	*metric
}

func NewBuilder() *Builder {
	b := &Builder{
		TimeFunc:      time.Now,
		TimePrecision: 1 * time.Nanosecond,
	}
	b.Reset()
	return b
}

func (b *Builder) SetName(name string) {
	b.name = name
}

func (b *Builder) AddTag(key string, value string) {
	b.metric.AddTag(key, value)
}

func (b *Builder) AddField(key string, value interface{}) {
	b.metric.AddField(key, value)
}

func (b *Builder) SetTime(tm time.Time) {
	b.tm = tm
}

func (b *Builder) Reset() {
	b.metric = &metric{
		tp: telegraf.Untyped,
	}
}

func (b *Builder) Metric() (telegraf.Metric, error) {
	if b.tm.IsZero() {
		b.tm = b.TimeFunc().Truncate(b.TimePrecision)
	}

	return b.metric, nil
}
