//go:generate ../../../tools/readme_config_includer/generator
package scaler

import (
	_ "embed"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

func (*Scaler) SampleConfig() string {
	return sampleConfig
}

type Scaling struct {
	InMin  float64  `toml:"input_minimum"`
	InMax  float64  `toml:"input_maximum"`
	OutMin float64  `toml:"output_minimum"`
	OutMax float64  `toml:"output_maximum"`
	Fields []string `toml:"fields"`
}

type Scaler struct {
	Scalings   []Scaling       `toml:"scaling"`
	Log        telegraf.Logger `toml:"-"`
	scalingMap map[filter.Filter]*Scaling
}

func (s *Scaler) Init() error {
	s.scalingMap = make(map[filter.Filter]*Scaling)

	// convert filter list to filter map for better performance
	for i, element := range s.Scalings {
		filter, err := filter.Compile(element.Fields)

		if err != nil {
			s.Log.Errorf("Could not compile filter: %v\n", err)
			return nil
		}

		if element.InMax != element.InMin {
			s.scalingMap[filter] = &s.Scalings[i]
		} else {
			s.Log.Error("Found scaling with equal input_minimum and input_maximum. Skipping it.")
		}
	}

	return nil
}

// scale a float according to the input and output range
func Scale(value float64, in_min float64, in_max float64, out_min float64, out_max float64) float64 {
	return (value-in_min)*(out_max-out_min)/(in_max-in_min) + out_min
}

// convert a numeric value to float
func toFloat(v interface{}) (float64, bool) {
	switch value := v.(type) {
	case int64:
		return float64(value), true
	case uint64:
		return float64(value), true
	case float64:
		return value, true
	}
	return 0.0, false
}

// handle the scaling process
func (s *Scaler) ScaleValues(metric telegraf.Metric) {
	if s.Scalings == nil || s.scalingMap == nil || len(s.scalingMap) == 0 {
		s.Log.Errorf("No valid scalings defined. Skipping scaling")
		return
	}

	fields := metric.Fields()

	for key := range fields {
		for filter, scaling := range s.scalingMap {
			if filter != nil && filter.Match(key) {

				// This call will always succeed as we are only using the fields from this specific metric
				value, _ := metric.GetField(key)

				v, ok := toFloat(value)

				if !ok {
					metric.RemoveField(key)
					s.Log.Errorf("error converting to float [%T]: %v\n", value, value)
					continue
				}

				// replace filed with the new value (the name remains the same)
				metric.RemoveField(key)
				res := Scale(v, scaling.InMin, scaling.InMax, scaling.OutMin, scaling.OutMax)
				metric.AddField(key, res)
			}
		}
	}
}

func (p *Scaler) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		p.ScaleValues(metric)
	}
	return in
}

func init() {
	processors.Add("scaler", func() telegraf.Processor {
		return &Scaler{}
	})
}
