//go:generate ../../../tools/readme_config_includer/generator
package Scale

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

func (*Scale) SampleConfig() string {
	return sampleConfig
}

type Scaling struct {
	InMin  float64  `toml:"input_minimum"`
	InMax  float64  `toml:"input_maximum"`
	OutMin float64  `toml:"output_minimum"`
	OutMax float64  `toml:"output_maximum"`
	Fields []string `toml:"fields"`

	factor      float64
	fieldFilter filter.Filter
}

type Scale struct {
	Scalings []Scaling       `toml:"scaling"`
	Log      telegraf.Logger `toml:"-"`
}

func (s *Scaling) init() error {
	if s.InMax == s.InMin {
		return fmt.Errorf("input minimum and maximum are equal for fields %s", strings.Join(s.Fields, ","))
	}

	if s.OutMax == s.OutMin {
		return fmt.Errorf("output minimum and maximum are equal for fields %s", strings.Join(s.Fields, ","))
	}

	scalingFilter, err := filter.Compile(s.Fields)
	if err != nil {
		return fmt.Errorf("could not compile fields filter: %w", err)
	}
	s.fieldFilter = scalingFilter

	s.factor = (s.OutMax - s.OutMin) / (s.InMax - s.InMin)
	return nil
}

// scale a float according to the input and output range
func (s *Scaling) process(value float64) float64 {
	return (value-s.InMin)*s.factor + s.OutMin
}

func (s *Scale) Init() error {
	if s.Scalings == nil {
		return errors.New("no valid scalings defined")
	}

	allFields := make(map[string]bool)
	for i := range s.Scalings {
		for _, field := range s.Scalings[i].Fields {
			// only generate a warning for the first duplicate field filter
			if warn, ok := allFields[field]; ok && warn {
				s.Log.Warnf("filter field %q used twice in scalings", field)
				allFields[field] = false
			} else {
				allFields[field] = true
			}
		}

		if err := s.Scalings[i].init(); err != nil {
			return fmt.Errorf("scaling %d: %w", i+1, err)
		}
	}
	return nil
}

// handle the scaling process
func (s *Scale) scaleValues(metric telegraf.Metric) {
	fields := metric.FieldList()

	for _, scaling := range s.Scalings {
		for _, field := range fields {
			if !scaling.fieldFilter.Match(field.Key) {
				continue
			}

			v, err := internal.ToFloat64(field.Value)
			if err != nil {
				s.Log.Errorf("error converting %q to float: %w\n", field.Key, err)
				continue
			}

			// scale the field values using the defined scaler
			field.Value = scaling.process(v)
		}
	}
}

func (s *Scale) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		s.scaleValues(metric)
	}
	return in
}

func init() {
	processors.Add("scale", func() telegraf.Processor {
		return &Scale{}
	})
}
