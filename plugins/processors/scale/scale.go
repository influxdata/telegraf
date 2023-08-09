//go:generate ../../../tools/readme_config_includer/generator
package scale

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
	InMin  *float64 `toml:"input_minimum"`
	InMax  *float64 `toml:"input_maximum"`
	OutMin *float64 `toml:"output_minimum"`
	OutMax *float64 `toml:"output_maximum"`
	Factor *float64 `toml:"factor"`
	Offset *float64 `toml:"offset"`
	Fields []string `toml:"fields"`

	fieldFilter filter.Filter
	scale       float64
	shiftIn     float64
	shiftOut    float64
}

type Scale struct {
	Scalings []Scaling       `toml:"scaling"`
	Log      telegraf.Logger `toml:"-"`
}

func (s *Scaling) Init() error {
	s.scale, s.shiftOut, s.shiftIn = float64(1.0), float64(0.0), float64(0.0)
	allMinMaxSet := s.OutMax != nil && s.OutMin != nil && s.InMax != nil && s.InMin != nil
	anyMinMaxSet := s.OutMax != nil || s.OutMin != nil || s.InMax != nil || s.InMin != nil
	factorSet := s.Factor != nil || s.Offset != nil
	if anyMinMaxSet && factorSet {
		return fmt.Errorf("cannot use factor/offset and minimum/maximum at the same time for fields %s",
			strings.Join(s.Fields, ","))
	} else if anyMinMaxSet && !allMinMaxSet {
		return fmt.Errorf("all minimum and maximum values need to be set for fields %s", strings.Join(s.Fields, ","))
	} else if !anyMinMaxSet && !factorSet {
		return fmt.Errorf("no scaling defined for fields %s", strings.Join(s.Fields, ","))
	} else if allMinMaxSet {
		if *s.InMax == *s.InMin {
			return fmt.Errorf("input minimum and maximum are equal for fields %s", strings.Join(s.Fields, ","))
		}

		if *s.OutMax == *s.OutMin {
			return fmt.Errorf("output minimum and maximum are equal for fields %s", strings.Join(s.Fields, ","))
		}

		s.scale = (*s.OutMax - *s.OutMin) / (*s.InMax - *s.InMin)
		s.shiftOut = *s.OutMin
		s.shiftIn = *s.InMin
	} else {
		if s.Factor != nil {
			s.scale = *s.Factor
		}
		if s.Offset != nil {
			s.shiftOut = *s.Offset
		}
	}

	scalingFilter, err := filter.Compile(s.Fields)
	if err != nil {
		return fmt.Errorf("could not compile fields filter: %w", err)
	}
	s.fieldFilter = scalingFilter

	return nil
}

// scale a float according to the input and output range
func (s *Scaling) process(value float64) float64 {
	return s.scale*(value-s.shiftIn) + s.shiftOut
}

func (s *Scale) Init() error {
	if s.Scalings == nil {
		return errors.New("no valid scaling defined")
	}

	allFields := make(map[string]bool)
	for i := range s.Scalings {
		for _, field := range s.Scalings[i].Fields {
			// only generate a warning for the first duplicate field filter
			if warn, ok := allFields[field]; ok && warn {
				s.Log.Warnf("Filter field %q used twice in scalings", field)
				allFields[field] = false
			} else {
				allFields[field] = true
			}
		}

		if err := s.Scalings[i].Init(); err != nil {
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
				s.Log.Errorf("Error converting %q to float: %v", field.Key, err)
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
