//go:generate ../../../tools/readme_config_includer/generator
package split

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type Split struct {
	Templates    []template `toml:"template"`
	DropOriginal bool       `toml:"drop_original"`
}

type template struct {
	Name   string   `toml:"name"`
	Tags   []string `toml:"tags"`
	Fields []string `toml:"fields"`

	fieldFilters filter.Filter
	tagFilters   filter.Filter
}

func (*Split) SampleConfig() string {
	return sampleConfig
}

func (s *Split) Init() error {
	if len(s.Templates) == 0 {
		return errors.New("at least one template required")
	}

	for index, template := range s.Templates {
		if template.Name == "" {
			return errors.New("metric name cannot be empty")
		}

		if len(template.Fields) == 0 {
			return errors.New("at least one field is required for a valid metric")
		}
		f, err := filter.Compile(template.Fields)
		if err != nil {
			return fmt.Errorf("failed to create new field filter: %w", err)
		}
		s.Templates[index].fieldFilters = f

		if len(template.Tags) != 0 {
			f, err := filter.Compile(template.Tags)
			if err != nil {
				return fmt.Errorf("failed to create new tag filter: %w", err)
			}
			s.Templates[index].tagFilters = f
		}
	}

	return nil
}

func (s *Split) Apply(in ...telegraf.Metric) []telegraf.Metric {
	newMetrics := []telegraf.Metric{}

	for _, point := range in {
		if s.DropOriginal {
			point.Accept()
		} else {
			newMetrics = append(newMetrics, point)
		}

		for _, template := range s.Templates {
			fields := make(map[string]any, len(point.FieldList()))
			for _, field := range point.FieldList() {
				if template.fieldFilters.Match(field.Key) {
					fields[field.Key] = field.Value
				}
			}

			tags := make(map[string]string, len(point.TagList()))
			if len(template.Tags) != 0 {
				for _, tag := range point.TagList() {
					if template.tagFilters.Match(tag.Key) {
						tags[tag.Key] = tag.Value
					}
				}
			}

			// metric with no fields should be skipped
			if len(fields) == 0 {
				continue
			}

			m := metric.New(template.Name, tags, fields, point.Time())
			newMetrics = append(newMetrics, m)
		}
	}

	return newMetrics
}

func init() {
	processors.Add("split", func() telegraf.Processor {
		return &Split{}
	})
}
