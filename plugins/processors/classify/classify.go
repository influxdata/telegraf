//go:generate ../../../tools/readme_config_includer/generator
package classify

import (
	_ "embed"
	"errors"
	"fmt"
	"regexp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type Classify struct {
	// Selector: which tag/field value determines the regex group to use.
	// Mutually exclusive; omit both to use DefaultRegexGroup unconditionally.
	SelectorTag   string `toml:"selector_tag"`
	SelectorField string `toml:"selector_field"`

	// Ordered pattern-to-group mapping applied to the selector value.
	SelectorMapping []selectorMapping `toml:"selector_mapping"`

	// Regex group to fall back to when no selector mapping matches.
	DefaultRegexGroup string `toml:"default_regex_group"`

	// Match: which tag/field value is tested against category regexes.
	// Exactly one must be defined.
	MatchTag   string `toml:"match_tag"`
	MatchField string `toml:"match_field"`

	// DefaultCategory is used when no regex matches. Metrics are dropped if empty.
	DefaultCategory string `toml:"default_category"`

	// DropCategories lists categories whose matched metrics are dropped.
	DropCategories []string `toml:"drop_categories"`

	// Result: where to write the classification outcome.
	// Exactly one must be defined.
	ResultTag   string `toml:"result_tag"`
	ResultField string `toml:"result_field"`

	// MappedSelectorRegexes maps each group name to its ordered categories.
	MappedSelectorRegexes map[string][]category `toml:"mapped_selector_regexes"`

	Log telegraf.Logger `toml:"-"`

	// Internal state derived from config fields during Init.
	selectorRules    []selectorRule
	categoryRules    map[string][]categoryRule
	dropThisCategory map[string]bool
}

type selectorMapping struct {
	Pattern string `toml:"pattern"`
	Group   string `toml:"group"`
}

type category struct {
	Category string   `toml:"category"`
	Patterns []string `toml:"patterns"`
}

type selectorRule struct {
	re    *regexp.Regexp
	group string
}

type categoryRule struct {
	name    string
	regexes []*regexp.Regexp
}

func (r categoryRule) matches(s string) bool {
	for _, re := range r.regexes {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}

func (*Classify) SampleConfig() string {
	return sampleConfig
}

// Init validates config and compiles all regular expressions.
func (cl *Classify) Init() error {
	if cl.SelectorTag != "" && cl.SelectorField != "" {
		return errors.New("selector_tag and selector_field cannot both be defined")
	}
	if cl.MatchTag == "" && cl.MatchField == "" {
		return errors.New("either match_tag or match_field must be defined")
	}
	if cl.MatchTag != "" && cl.MatchField != "" {
		return errors.New("match_tag and match_field cannot both be defined")
	}
	if cl.ResultTag == "" && cl.ResultField == "" {
		return errors.New("either result_tag or result_field must be defined")
	}
	if cl.ResultTag != "" && cl.ResultField != "" {
		return errors.New("result_tag and result_field cannot both be defined")
	}

	known := make(map[string]bool)
	cl.categoryRules = make(map[string][]categoryRule, len(cl.MappedSelectorRegexes))
	for group, categories := range cl.MappedSelectorRegexes {
		if len(categories) == 0 {
			return fmt.Errorf("mapped_selector_regexes group %q has no categories", group)
		}
		seen := make(map[string]bool, len(categories))
		for _, c := range categories {
			if c.Category == "" {
				return fmt.Errorf("empty category name in mapped_selector_regexes group %q", group)
			}
			if seen[c.Category] {
				return fmt.Errorf("duplicate category %q in mapped_selector_regexes group %q", c.Category, group)
			}
			seen[c.Category] = true
			if len(c.Patterns) == 0 {
				return fmt.Errorf("mapped_selector_regexes group %q category %q has no patterns", group, c.Category)
			}
			rule := categoryRule{name: c.Category}
			for _, pattern := range c.Patterns {
				if pattern == "" {
					return fmt.Errorf("empty pattern in mapped_selector_regexes group %q category %q", group, c.Category)
				}
				re, err := regexp.Compile(pattern)
				if err != nil {
					return fmt.Errorf("invalid pattern %q in mapped_selector_regexes group %q category %q: %w", pattern, group, c.Category, err)
				}
				rule.regexes = append(rule.regexes, re)
			}
			known[c.Category] = true
			cl.categoryRules[group] = append(cl.categoryRules[group], rule)
		}
	}
	if len(cl.categoryRules) == 0 {
		return errors.New("mapped_selector_regexes has no groups defined")
	}

	if cl.DefaultRegexGroup != "" {
		if _, ok := cl.categoryRules[cl.DefaultRegexGroup]; !ok {
			return fmt.Errorf("default_regex_group %q is not a group in mapped_selector_regexes", cl.DefaultRegexGroup)
		}
	} else if cl.SelectorTag == "" && cl.SelectorField == "" {
		return errors.New("default_regex_group must be set when neither selector_tag nor selector_field is defined")
	}

	seenPattern := make(map[string]bool, len(cl.SelectorMapping))
	cl.selectorRules = make([]selectorRule, 0, len(cl.SelectorMapping))
	for _, m := range cl.SelectorMapping {
		if m.Pattern == "" {
			return fmt.Errorf("empty pattern in selector_mapping for group %q", m.Group)
		}
		if seenPattern[m.Pattern] {
			return fmt.Errorf("duplicate selector_mapping pattern %q", m.Pattern)
		}
		seenPattern[m.Pattern] = true
		// "*" passes the selector value through as the group name and ""
		// drops the metric, so only other names must be defined groups.
		if _, ok := cl.categoryRules[m.Group]; !ok && m.Group != "*" && m.Group != "" {
			return fmt.Errorf("selector_mapping group %q is not a group in mapped_selector_regexes", m.Group)
		}
		re, err := regexp.Compile(m.Pattern)
		if err != nil {
			return fmt.Errorf("invalid selector_mapping pattern %q for group %q: %w", m.Pattern, m.Group, err)
		}
		cl.selectorRules = append(cl.selectorRules, selectorRule{re: re, group: m.Group})
	}

	cl.dropThisCategory = make(map[string]bool, len(cl.DropCategories))
	for _, c := range cl.DropCategories {
		if c == "" {
			return errors.New("drop_categories contains an empty string")
		}
		if !known[c] && c != cl.DefaultCategory {
			return fmt.Errorf("%q in drop_categories is not a known category or default_category", c)
		}
		cl.dropThisCategory[c] = true
	}

	return nil
}

// Apply classifies the metrics and returns the ones that are kept.
func (cl *Classify) Apply(in ...telegraf.Metric) []telegraf.Metric {
	out := make([]telegraf.Metric, 0, len(in))
	for _, metric := range in {
		category, ok := cl.classify(metric)
		if !ok {
			metric.Drop()
			continue
		}
		if cl.ResultTag != "" {
			metric.AddTag(cl.ResultTag, category)
		} else {
			metric.AddField(cl.ResultField, category)
		}
		out = append(out, metric)
	}
	return out
}

// classify resolves the category of a metric and reports whether to keep it.
func (cl *Classify) classify(metric telegraf.Metric) (string, bool) {
	var selector string
	group := cl.DefaultRegexGroup
	if cl.SelectorTag != "" || cl.SelectorField != "" {
		var ok bool
		if selector, ok = cl.lookupString(metric, cl.SelectorTag, cl.SelectorField, "selector"); !ok {
			return "", false
		}
		group = cl.selectGroup(selector)
	}
	if group == "" {
		cl.Log.Tracef("Dropping point (selector item value %q maps to an empty regex group)", selector)
		return "", false
	}

	rules, ok := cl.categoryRules[group]
	if !ok {
		// Only a "*" passthrough can yield an unknown group, and Init
		// guarantees that a non-empty default_regex_group is a known group.
		if cl.DefaultRegexGroup == "" {
			cl.Log.Tracef("Dropping point (selector %q maps to %q, which is not a known regex group)", selector, group)
			return "", false
		}
		rules = cl.categoryRules[cl.DefaultRegexGroup]
		group = cl.DefaultRegexGroup
	}

	value, ok := cl.lookupString(metric, cl.MatchTag, cl.MatchField, "match")
	if !ok {
		return "", false
	}

	category := cl.DefaultCategory
	for _, rule := range rules {
		if rule.matches(value) {
			category = rule.name
			break
		}
	}
	cl.Log.Tracef("Regex group %q classified the point as %q", group, category)

	if category == "" {
		cl.Log.Trace("Dropping point (no match and default_category is not set)")
		return "", false
	}
	if cl.dropThisCategory[category] {
		cl.Log.Tracef("Dropping point (category %q is in drop_categories)", category)
		return "", false
	}
	return category, true
}

// selectGroup maps a selector value to a regex group; the first matching
// selector_mapping entry wins.
func (cl *Classify) selectGroup(selector string) string {
	for _, rule := range cl.selectorRules {
		if rule.re.MatchString(selector) {
			if rule.group == "*" {
				return selector
			}
			return rule.group
		}
	}
	cl.Log.Tracef("Selector item value %q does not match anything in selector_mapping", selector)
	return cl.DefaultRegexGroup
}

// lookupString reads a string value from the given tag, or else from the
// given field, logging why the point is dropped when it is unavailable.
func (cl *Classify) lookupString(metric telegraf.Metric, tag, field, role string) (string, bool) {
	if tag != "" {
		value, ok := metric.GetTag(tag)
		if !ok {
			cl.Log.Tracef("Dropping point (%s tag %q is missing)", role, tag)
		}
		return value, ok
	}
	raw, ok := metric.GetField(field)
	if !ok {
		cl.Log.Tracef("Dropping point (%s field %q is missing)", role, field)
		return "", false
	}
	value, ok := raw.(string)
	if !ok {
		cl.Log.Tracef("Dropping point (%s field %q is not a string)", role, field)
	}
	return value, ok
}

func init() {
	processors.Add("classify", func() telegraf.Processor {
		return &Classify{}
	})
}
