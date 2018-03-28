package metric

import (
	"time"

	"github.com/influxdata/telegraf"
)

type TimeFunc func() time.Time

type Builder struct {
	TimeFunc

	*metric
}

func NewBuilder() *Builder {
	b := &Builder{
		TimeFunc: time.Now,
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
	b.metric = &metric{}
}

func (b *Builder) Metric() (telegraf.Metric, error) {
	if b.tm.IsZero() {
		b.tm = b.TimeFunc()
	}

	return b.metric, nil
}
