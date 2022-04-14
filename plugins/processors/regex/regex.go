package regex

import (
	"fmt"
	"regexp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Regex struct {
	Tags         []converter     `toml:"tags"`
	Fields       []converter     `toml:"fields"`
	TagRename    []converter     `toml:"tag_rename"`
	FieldRename  []converter     `toml:"field_rename"`
	MetricRename []converter     `toml:"metric_rename"`
	Log          telegraf.Logger `toml:"-"`
	regexCache   map[string]*regexp.Regexp
}

type converter struct {
	Key         string `toml:"key"`
	Pattern     string `toml:"pattern"`
	Replacement string `toml:"replacement"`
	ResultKey   string `toml:"result_key"`
	Append      bool   `toml:"append"`
}

func (r *Regex) Init() error {
	r.regexCache = make(map[string]*regexp.Regexp)

	// Compile the regular expressions
	for _, c := range r.Tags {
		if _, compiled := r.regexCache[c.Pattern]; !compiled {
			r.regexCache[c.Pattern] = regexp.MustCompile(c.Pattern)
		}
	}
	for _, c := range r.Fields {
		if _, compiled := r.regexCache[c.Pattern]; !compiled {
			r.regexCache[c.Pattern] = regexp.MustCompile(c.Pattern)
		}
	}

	resultOptions := []string{"overwrite", "keep"}
	for _, c := range r.TagRename {
		if c.Key != "" {
			r.Log.Info("'tag_rename' section contains a key which is ignored during processing")
		}

		if c.ResultKey == "" {
			c.ResultKey = "keep"
		}
		if err := choice.Check(c.ResultKey, resultOptions); err != nil {
			return fmt.Errorf("invalid metrics result_key: %v", err)
		}

		if _, compiled := r.regexCache[c.Pattern]; !compiled {
			r.regexCache[c.Pattern] = regexp.MustCompile(c.Pattern)
		}
	}

	for _, c := range r.FieldRename {
		if c.Key != "" {
			r.Log.Info("'field_rename' section contains a key which is ignored during processing")
		}

		if c.ResultKey == "" {
			c.ResultKey = "keep"
		}
		if err := choice.Check(c.ResultKey, resultOptions); err != nil {
			return fmt.Errorf("invalid metrics result_key: %v", err)
		}

		if _, compiled := r.regexCache[c.Pattern]; !compiled {
			r.regexCache[c.Pattern] = regexp.MustCompile(c.Pattern)
		}
	}

	for _, c := range r.MetricRename {
		if c.Key != "" {
			r.Log.Info("'metric_rename' section contains a key which is ignored during processing")
		}

		if c.ResultKey != "" {
			r.Log.Info("'metric_rename' section contains a 'result_key' ignored during processing as metrics will ALWAYS the name")
		}

		if _, compiled := r.regexCache[c.Pattern]; !compiled {
			r.regexCache[c.Pattern] = regexp.MustCompile(c.Pattern)
		}
	}

	return nil
}

func (r *Regex) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		for _, converter := range r.Tags {
			if converter.Key == "*" {
				for _, tag := range metric.TagList() {
					regex := r.regexCache[converter.Pattern]
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
			regex := r.regexCache[converter.Pattern]
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
			regex := r.regexCache[converter.Pattern]
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
			regex := r.regexCache[converter.Pattern]
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
	regex := r.regexCache[c.Pattern]

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
