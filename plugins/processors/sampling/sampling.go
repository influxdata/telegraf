package sampling

import (
	"time"

	"bytes"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Sampling struct {
	Period    internal.Duration
	FieldKeys bool

	sampled     map[string]struct{}
	samplingEnd time.Time
}

var empty struct{}

var sampleConfig = `
`

func (s *Sampling) SampleConfig() string {
	return sampleConfig
}

func (s *Sampling) Description() string {
	return "Samples all metrics based on a specified time period."
}

func (s *Sampling) Apply(in ...telegraf.Metric) []telegraf.Metric {
	if s.samplingEnd.Before(time.Now()) {
		s.samplingEnd = time.Now().Add(s.Period.Duration)
		s.sampled = make(map[string]struct{})
	}

	out := []telegraf.Metric{}

	for _, metric := range in {
		key := s.metricKey(metric)
		if _, ok := s.sampled[key]; !ok {
			out = append(out, metric)
			s.sampled[key] = empty
		}
	}

	return out
}

func (s *Sampling) metricKey(metric telegraf.Metric) string {
	key := bytes.NewBufferString("")

	key.WriteString(metric.Name())
	key.WriteString(":")
	for k, v := range metric.Tags() {
		key.WriteString(k)
		key.WriteString("=")
		key.WriteString(v)
	}

	return key.String()
}

func init() {
	processors.Add("sampling", func() telegraf.Processor {
		return &Sampling{
			sampled: make(map[string]struct{}),
		}
	})
}
