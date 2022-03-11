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

const sampleConfig = `
  ## Tag and field conversions defined in a separate sub-tables
  # [[processors.regex.tags]]
  #   ## Tag to change
  #   key = "resp_code"
  #   ## Regular expression to match on a tag value
  #   pattern = "^(\\d)\\d\\d$"
  #   ## Matches of the pattern will be replaced with this string.  Use ${1}
  #   ## notation to use the text of the first submatch.
  #   replacement = "${1}xx"

  # [[processors.regex.fields]]
  #   ## Field to change
  #   key = "request"
  #   ## All the power of the Go regular expressions available here
  #   ## For example, named subgroups
  #   pattern = "^/api(?P<method>/[\\w/]+)\\S*"
  #   replacement = "${method}"
  #   ## If result_key is present, a new field will be created
  #   ## instead of changing existing field
  #   result_key = "method"

  ## Multiple conversions may be applied for one field sequentially
  ## Let's extract one more value
  # [[processors.regex.fields]]
  #   key = "request"
  #   pattern = ".*category=(\\w+).*"
  #   replacement = "${1}"
  #   result_key = "search_category"

  ## Rename metric fields
  # [[processors.regex.field_rename]]
  #   ## Regular expression to match on a field name
  #   pattern = "^search_(\\w+)d$"
  #   ## Matches of the pattern will be replaced with this string.  Use ${1}
  #   ## notation to use the text of the first submatch.
  #   replacement = "${1}"
  #   ## If the new field name already exists, you can either "overwrite" the
  #   ## existing one with the value of the renamed field OR you can "keep"
  #   ## both the existing and source field.
  #   # result_key = "keep"

  ## Rename metric tags
  # [[processors.regex.tag_rename]]
  #   ## Regular expression to match on a tag name
  #   pattern = "^search_(\\w+)d$"
  #   ## Matches of the pattern will be replaced with this string.  Use ${1}
  #   ## notation to use the text of the first submatch.
  #   replacement = "${1}"
  #   ## If the new tag name already exists, you can either "overwrite" the
  #   ## existing one with the value of the renamed tag OR you can "keep"
  #   ## both the existing and source tag.
  #   # result_key = "keep"

  ## Rename metrics
  # [[processors.regex.metric_rename]]
  #   ## Regular expression to match on an metric name
  #   pattern = "^search_(\\w+)d$"
  #   ## Matches of the pattern will be replaced with this string.  Use ${1}
  #   ## notation to use the text of the first submatch.
  #   replacement = "${1}"
`

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

func (r *Regex) SampleConfig() string {
	return sampleConfig
}

func (r *Regex) Description() string {
	return "Transforms tag and field values as well as measurement, tag and field names with regex pattern"
}

func (r *Regex) Apply(in ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range in {
		for _, converter := range r.Tags {
			if value, ok := metric.GetTag(converter.Key); ok {
				if key, newValue := r.convert(converter, value); newValue != "" {
					if converter.Append {
						if v, ok := metric.GetTag(key); ok {
							newValue = v + newValue
						}
					}
					metric.AddTag(key, newValue)
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

func init() {
	processors.Add("regex", func() telegraf.Processor { return &Regex{} })
}
