//go:generate ../../../tools/readme_config_includer/generator
package Scale

import (
	_ "embed"
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

func (s *Scaling) Init() error {
	scalingFilter, err := filter.Compile(s.Fields)

	if err != nil {
		return fmt.Errorf("could not compile filter: %w", err)
	}

	s.fieldFilter = scalingFilter

	if s.InMax == s.InMin {
		return fmt.Errorf("minumum and maximum are equal for fields %s", strings.Join(s.Fields, ","))
	}

	s.factor = (s.OutMax - s.OutMin) / (s.InMax - s.InMin)
	return nil
}

func (s *Scale) Init() error {
	if s.Scalings == nil {
		return fmt.Errorf("no valid scalings defined. Skipping scaling")
	}

	allFields := make(map[string]bool, len(s.Scalings[0].Fields))
	for i := range s.Scalings {
		for _, field := range s.Scalings[i].Fields {
			if _, ok := allFields[field]; ok {
				return fmt.Errorf("filter field '%s' use twice in scalings", field)
			}

			allFields[field] = true
		}

		if res := s.Scalings[i].Init(); res != nil {
			return res
		}
	}
	return nil
}

// scale a float according to the input and output range
func (s *Scaling) Process(value float64) float64 {
	return (value-s.InMin)*s.factor + s.OutMin
}

// handle the scaling process
func (s *Scale) ScaleValues(metric telegraf.Metric) {
	fields := metric.FieldList()

	for _, scaling := range s.Scalings {
		for _, field := range fields {
			if !scaling.fieldFilter.Match(field.Key) {
				continue
			}

			v, err := internal.ToFloat64(field.Value)
			if err != nil {
				s.Log.Errorf("error converting '%v' to float: %v\n", field.Key, err)
				continue
			}

			// replace field with the new value (the name remains the same)
			field.Value = scaling.Process(v)
		}
	}
}

func (s *Scale) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		s.ScaleValues(metric)
	}
	return in
}

func init() {
	processors.Add("scale", func() telegraf.Processor {
		return &Scale{}
	})
}
