package regex

import (
	"fmt"
	"regexp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/processors"
)

type Regex struct {
	Tags       []converter     `toml:"tags"`
	Fields     []converter     `toml:"fields"`
	Metrics    []converter     `toml:"metrics"`
	Log        telegraf.Logger `toml:"-"`
	regexCache map[string]*regexp.Regexp
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

  ## Replace metric element names such as "measurement", "tag" or "field" name
  # [[processors.regex.metrics]]
  #   ## Element name to change, can be "measurement", "tags" or "fields"
  #   key = "fields"
  #   ## Regular expression to match on a element name
  #   pattern = "^value_(\\d)_\\d_\\d$"
  #   ## Matches of the pattern will be replaced with this string.  Use ${1}
  #   ## notation to use the text of the first submatch.
  #   replacement = "${1}"
  #   ## If the new tag or field name is already present, you can either
  #   ## "overwrite" or "keep" the existing tag or field.
  #   # result_key = "keep"
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
	for _, c := range r.Metrics {
		if err := choice.Check(c.Key, []string{"measurement", "fields", "tags"}); err != nil {
			return fmt.Errorf("invalid metrics key: %v", err)
		}

		if c.ResultKey == "" {
			c.ResultKey = "keep"
		}
		if err := choice.Check(c.ResultKey, []string{"overwrite", "keep"}); err != nil {
			return fmt.Errorf("invalid metrics result_key: %v", err)
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

		for _, converter := range r.Metrics {
			regex := r.regexCache[converter.Pattern]

			switch converter.Key {
			case "measurement":
				value := metric.Name()
				if regex.MatchString(value) {
					newValue := regex.ReplaceAllString(value, converter.Replacement)
					metric.SetName(newValue)
				}
			case "fields":
				replacements := make(map[string]string)
				for _, field := range metric.FieldList() {
					name := field.Key
					if regex.MatchString(name) {
						newName := regex.ReplaceAllString(name, converter.Replacement)

						if !metric.HasField(newName) {
							// There is no colliding field, we can just change the name.
							field.Key = newName
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
					metric.RemoveField(newName)
					metric.AddField(newName, value)
					metric.RemoveField(oldName)
				}
			case "tags":
				replacements := make(map[string]string)
				for _, tag := range metric.TagList() {
					name := tag.Key
					if regex.MatchString(name) {
						newName := regex.ReplaceAllString(name, converter.Replacement)

						if !metric.HasTag(newName) {
							// There is no colliding tag, we can just change the name.
							tag.Key = newName
						}

						if converter.ResultKey == "overwrite" {
							// We got a colliding tag, remember the replacement and do it later
							replacements[name] = newName
						}
					}
				}
				// We needed to postpone the replacement as we cannot modify the field-list
				// while iterating it as this will result in invalid memory dereference panic.
				for oldName, newName := range replacements {
					value, ok := metric.GetTag(oldName)
					if !ok {
						// Just in case the field got removed in the meantime
						continue
					}
					metric.RemoveTag(newName)
					metric.AddTag(newName, value)
					metric.RemoveTag(oldName)
				}
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
