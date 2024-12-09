//go:generate ../../../tools/readme_config_includer/generator
package regex

import (
	_ "embed"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type Regex struct {
	Tags         []converter     `toml:"tags"`
	Fields       []converter     `toml:"fields"`
	TagRename    []converter     `toml:"tag_rename"`
	FieldRename  []converter     `toml:"field_rename"`
	MetricRename []converter     `toml:"metric_rename"`
	Log          telegraf.Logger `toml:"-"`
}

func (*Regex) SampleConfig() string {
	return sampleConfig
}

func (r *Regex) Init() error {
	// Compile the regular expressions
	for i := range r.Tags {
		if err := r.Tags[i].setup(convertTags, r.Log); err != nil {
			return fmt.Errorf("'tags' %w", err)
		}
	}
	for i := range r.Fields {
		if err := r.Fields[i].setup(convertFields, r.Log); err != nil {
			return fmt.Errorf("'fields' %w", err)
		}
	}

	for i, c := range r.TagRename {
		if c.Key != "" {
			r.Log.Info("'tag_rename' section contains a key which is ignored during processing")
		}
		if err := r.TagRename[i].setup(convertTagRename, r.Log); err != nil {
			return fmt.Errorf("'tag_rename' %w", err)
		}
	}

	for i, c := range r.FieldRename {
		if c.Key != "" {
			r.Log.Info("'field_rename' section contains a key which is ignored during processing")
		}

		if err := r.FieldRename[i].setup(convertFieldRename, r.Log); err != nil {
			return fmt.Errorf("'field_rename' %w", err)
		}
	}

	for i, c := range r.MetricRename {
		if c.Key != "" {
			r.Log.Info("'metric_rename' section contains a key which is ignored during processing")
		}

		if c.ResultKey != "" {
			r.Log.Info("'metric_rename' section contains a 'result_key' ignored during processing as metrics will ALWAYS the name")
		}

		if err := r.MetricRename[i].setup(convertMetricRename, r.Log); err != nil {
			return fmt.Errorf("'metric_rename' %w", err)
		}
	}

	return nil
}

func (r *Regex) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		for _, c := range r.Tags {
			c.apply(metric)
		}

		for _, c := range r.Fields {
			c.apply(metric)
		}

		for _, c := range r.TagRename {
			c.apply(metric)
		}

		for _, c := range r.FieldRename {
			c.apply(metric)
		}

		for _, c := range r.MetricRename {
			c.apply(metric)
		}
	}

	return in
}

func init() {
	processors.Add("regex", func() telegraf.Processor { return &Regex{} })
}
