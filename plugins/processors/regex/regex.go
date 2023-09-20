//go:generate ../../../tools/readme_config_includer/generator
package regex

import (
	_ "embed"
	"fmt"
	"regexp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type converterType int

const (
	convertTags = iota
	convertFields
	convertTagRename
	convertFieldRename
	convertMetricRename
)

type Regex struct {
	Tags         []converter     `toml:"tags"`
	Fields       []converter     `toml:"fields"`
	TagRename    []converter     `toml:"tag_rename"`
	FieldRename  []converter     `toml:"field_rename"`
	MetricRename []converter     `toml:"metric_rename"`
	Log          telegraf.Logger `toml:"-"`
}

type converter struct {
	Key         string `toml:"key"`
	Pattern     string `toml:"pattern"`
	Replacement string `toml:"replacement"`
	ResultKey   string `toml:"result_key"`
	Append      bool   `toml:"append"`

	re *regexp.Regexp
}

func (c *converter) setup(ct converterType) error {
	switch ct {
	case convertTagRename, convertFieldRename:
		switch c.ResultKey {
		case "":
			c.ResultKey = "keep"
		case "overwrite", "keep":
			// Do nothing as those are valid choices
		default:
			return fmt.Errorf("invalid metrics result_key %q", c.ResultKey)
		}
	}

	var err error
	c.re, err = regexp.Compile(c.Pattern)

	return err
}

func (*Regex) SampleConfig() string {
	return sampleConfig
}

func (r *Regex) Init() error {
	// Compile the regular expressions
	for i := range r.Tags {
		if err := r.Tags[i].setup(convertTags); err != nil {
			return fmt.Errorf("'tags' %w", err)
		}
	}
	for i := range r.Fields {
		if err := r.Fields[i].setup(convertFields); err != nil {
			return fmt.Errorf("'fields' %w", err)
		}
	}

	for i, c := range r.TagRename {
		if c.Key != "" {
			r.Log.Info("'tag_rename' section contains a key which is ignored during processing")
		}
		if err := r.TagRename[i].setup(convertTagRename); err != nil {
			return fmt.Errorf("'tag_rename' %w", err)
		}
	}

	for i, c := range r.FieldRename {
		if c.Key != "" {
			r.Log.Info("'field_rename' section contains a key which is ignored during processing")
		}

		if err := r.FieldRename[i].setup(convertFieldRename); err != nil {
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

		if err := r.MetricRename[i].setup(convertMetricRename); err != nil {
			return fmt.Errorf("'metric_rename' %w", err)
		}
	}

	return nil
}

func (r *Regex) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		for _, converter := range r.Tags {
			if converter.Key == "*" {
				for _, tag := range metric.TagList() {
					regex := converter.re
					if regex.MatchString(tag.Value) {
						newValue := regex.ReplaceAllString(tag.Value, converter.Replacement)
						updateTag(converter, metric, tag.Key, newValue)
					}
				}
			} else if value, ok := metric.GetTag(converter.Key); ok {
				if key, newValue := r.convert(converter, value); newValue != "" {
					updateTag(converter, metric, key, newValue)
				}
			}
		}

		for _, converter := range r.Fields {
			if value, ok := metric.GetField(converter.Key); ok {
				if v, ok := value.(string); ok {
					if key, newValue := r.convert(converter, v); newValue != "" {
						metric.AddField(key, newValue)
					}
				}
			}
		}

		for _, converter := range r.TagRename {
			regex := converter.re
			replacements := make(map[string]string)
			for _, tag := range metric.TagList() {
				name := tag.Key
				if regex.MatchString(name) {
					newName := regex.ReplaceAllString(name, converter.Replacement)

					if !metric.HasTag(newName) {
						// There is no colliding tag, we can just change the name.
						tag.Key = newName
						continue
					}

					if converter.ResultKey == "overwrite" {
						// We got a colliding tag, remember the replacement and do it later
						replacements[name] = newName
					}
				}
			}
			// We needed to postpone the replacement as we cannot modify the tag-list
			// while iterating it as this will result in invalid memory dereference panic.
			for oldName, newName := range replacements {
				value, ok := metric.GetTag(oldName)
				if !ok {
					// Just in case the tag got removed in the meantime
					continue
				}
				metric.AddTag(newName, value)
				metric.RemoveTag(oldName)
			}
		}

		for _, converter := range r.FieldRename {
			regex := converter.re
			replacements := make(map[string]string)
			for _, field := range metric.FieldList() {
				name := field.Key
				if regex.MatchString(name) {
					newName := regex.ReplaceAllString(name, converter.Replacement)

					if !metric.HasField(newName) {
						// There is no colliding field, we can just change the name.
						field.Key = newName
						continue
					}

					if converter.ResultKey == "overwrite" {
						// We got a colliding field, remember the replacement and do it later
						replacements[name] = newName
					}
				}
			}
			// We needed to postpone the replacement as we cannot modify the field-list
			// while iterating it as this will result in invalid memory dereference panic.
			for oldName, newName := range replacements {
				value, ok := metric.GetField(oldName)
				if !ok {
					// Just in case the field got removed in the meantime
					continue
				}
				metric.AddField(newName, value)
				metric.RemoveField(oldName)
			}
		}

		for _, converter := range r.MetricRename {
			regex := converter.re
			value := metric.Name()
			if regex.MatchString(value) {
				newValue := regex.ReplaceAllString(value, converter.Replacement)
				metric.SetName(newValue)
			}
		}
	}

	return in
}

func (r *Regex) convert(c converter, src string) (key string, value string) {
	regex := c.re

	if c.ResultKey == "" || regex.MatchString(src) {
		value = regex.ReplaceAllString(src, c.Replacement)
	}

	if c.ResultKey != "" {
		return c.ResultKey, value
	}

	return c.Key, value
}

func updateTag(converter converter, metric telegraf.Metric, key string, newValue string) {
	if converter.Append {
		if v, ok := metric.GetTag(key); ok {
			newValue = v + newValue
		}
	}
	metric.AddTag(key, newValue)
}

func init() {
	processors.Add("regex", func() telegraf.Processor { return &Regex{} })
}
